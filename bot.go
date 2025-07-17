package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gorilla/websocket"
)

func bpWebsocket(ctx context.Context, cp ChanPkg, wg *sync.WaitGroup, url string) {
	defer wg.Done()
	slog.Info("bpWebsocket started")
	/* The bot is the oly thing that needs to be cleaned up therefore the bot */
	/* listens for SIGINT and SIGTERM, it then requests all go routines to    */
	/* stop*/

	var conn *websocket.Conn
	var err error
	// loop forever!
	for {
		select {
		case <-ctx.Done():
			/* Allow time for workers to stop */
			slog.Info("bpWebsocket shutting down")
			return
		default:
			slog.Info("Attempting websocket connection")
			conn, err = connectWebsocket(url)
			if err != nil {
				slog.Error("Error connecting to websocket", "err", err.Error())
				time.Sleep(time.Duration(WebsocketTimeout) * time.Second)
				continue
			}
			wg.Add(1)
			handleWebsocket(ctx, wg, conn, cp)

		}
	}
}

// returns true if parent should quit
func handleWebsocket(ctx context.Context, wg *sync.WaitGroup, conn *websocket.Conn, cp ChanPkg) {
	defer wg.Done()
	defer conn.Close()

	slog.Info("handleWebsocket started")

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				/* Only print a warning if context IS NOT cancelled */
				slog.Warn("Failed to read message from socket", "error", err)
			}
			cp.ByteSlice <- message

		}
	}()

	<-ctx.Done()
	/* Allow time for websocket to close cleanly */
	time.Sleep(time.Millisecond * 500)
	slog.Warn("handleWebsocket shutting down")
}

func connectWebsocket(url string) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Start(ctx context.Context, wg *sync.WaitGroup, cnf *Config, cp ChanPkg) error {
	defer wg.Done()

	var workers int = ByteWorker

	slog.Debug("Starting byte handlers")
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go handleBytes(ctx, cnf, wg, i, cp)
	}

	stream := fmt.Sprintf("%s%s%s", ServerArgsPre, cnf.JetStreamServer, ServerArgsPost)
	slog.Debug("Starting the websocket listener")
	wg.Add(1)
	go bpWebsocket(ctx, cp, wg, stream)

	// Only report buffer status during debug
	if LogLevel == slog.LevelDebug {
		go reportBsChanBuf(cp)
	}

	slog.Info("Bot startup complete")
	return nil
}

func handleBytes(ctx context.Context, cnf *Config, wg *sync.WaitGroup, id int, cp ChanPkg) {
	defer wg.Done()
	slog.Info("handleBytes worker starting", "WorkerId", id)
	for {
		select {
		case ba := <-cp.ByteSlice:
			slog.Debug("Data received", "WorkerId", id, "Length", len(ba))
			var msg Message
			err := json.Unmarshal(ba, &msg)
			if err != nil {
				slog.Error("Couldn't unmarshal message", "error", err.Error())
				return
			}
			err = handleMessage(cnf, id, &msg, cp)
			if err != nil {
				slog.Warn("Problem handling message", "WorkerId", id, "error", err.Error())
				//A single message failure does not require us to quit
			}
		case <-ctx.Done():
			slog.Info("handleBytes worker shutting down", "WorkerId", id)
			return
		}
	}
}

func handleMessage(cnf *Config, id int, msg *Message, cp ChanPkg) error {
	/* We only want to repost and like original posts, not replies. */
	/* Check if there is a reply path */
	if msg.Commit.Record.Reply.Parent.URI != "" {
		slog.Debug("Post is a reply, will not process", "WorkerId", id)
		return nil
	}
	if checkForTerms(cnf, msg) {
		slog.Info("Found a match", "WorkerId", id)
		d, err := resolveDID(msg.DID)
		if err != nil {
			slog.Warn("Could not resolve did of message")
		} else {
			/*Just use the first alias found for now*/
			if len(d.AlsoKnownAs) > 0 {
				uname := strings.TrimPrefix(d.AlsoKnownAs[0], "at://")
				slog.Info("Match info", "user", uname, "post", msg.Commit.Record.Text)
			} else {
				slog.Info("Match info", "post", msg.Commit.Record.Text)
			}
		}
		//TODO new post office logic user client/server model
		err = Ral(cnf, msg, cp)
		if err != nil {
			slog.Error("Repost failed", "err", err.Error())
		}
	}

	return nil
}

func checkForTerms(cnf *Config, msg *Message) bool {
	if msg.Commit.Record.Text == "" {
		return false // don't waste time on an empty record
	}

	// If DebugPosts is true and logging is set to Debug, this will print the
	// content of posts to the log
	if DebugPosts && LogLevel == slog.LevelDebug {
		slog.Debug("Post data", "text", msg.Commit.Record.Text)
	}

	strLower := strings.ToLower(msg.Commit.Record.Text)
	for _, v := range cnf.Terms {

		// Convert both strings to lowercase for case-insensitive comparison
		substrLower := strings.ToLower(v)

		// Convert to rune slices to handle Unicode properly
		strRunes := []rune(strLower)
		substrRunes := []rune(substrLower)

		// Search for the substring
		for i := 0; i <= len(strRunes)-len(substrRunes); i++ {
			// Check if substring matches at position i
			match := true
			for j := 0; j < len(substrRunes); j++ {
				if strRunes[i+j] != substrRunes[j] {
					match = false
					break
				}
			}

			if match {
				// Check the character immediately before the substring
				if i > 0 {
					prevChar := strRunes[i-1]
					if unicode.IsLetter(prevChar) || unicode.IsDigit(prevChar) {
						continue // This match is invalid, keep searching
					}
				}

				// Check the character immediately after the substring
				nextCharIndex := i + len(substrRunes)
				if nextCharIndex < len(strRunes) {
					nextChar := strRunes[nextCharIndex]
					if unicode.IsLetter(nextChar) || unicode.IsDigit(nextChar) {
						continue // This match is invalid, keep searching
					}
				}

				// If we get here, both boundaries are valid
				return true
			}
		}
	}
	return false
}

func reportBsChanBuf(cp ChanPkg) {
	// Run forever just spitting out the current byteslice buffer every minute
	tick := time.NewTicker(time.Second * 60)
	defer tick.Stop()
	timeout := time.Second * 70
	for {
		select {
		case <-tick.C:
			slog.Debug("ByteSlice channel stats", "BufferSize", "10", "ItemsInBuffer", len(cp.ByteSlice))
		case <-time.After(timeout):
			slog.Warn("ByteSlice channel stats", "msg", "Failed To Read in time")
		}
	}
}

func resolveDID(did string) (*DIDDoc, error) {

	u := fmt.Sprintf("%s/%s", DidLookUpEndpoint, did)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		slog.Error("Error creating request", "error", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to resolve did", "Error", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code", "status", resp)
		return nil, fmt.Errorf("unexpected status code")
	}

	var result DIDDoc
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("Failed to marshall respond to DIDResponse type")
		return nil, err
	}

	return &result, nil
}
