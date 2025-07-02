package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/gorilla/websocket"
)

func bpWebsocket(cp ChanPkg, wg *sync.WaitGroup, url string) {
	slog.Info("bpWebsocket started")
	/* The bot is the oly thing that needs to be cleaned up therefore the bot */
	/* listens for SIGINT and SIGTERM, it then requests all go routines to    */
	/* stop*/

	var conn *websocket.Conn
	var err error
	// loop forever!
	for {
		slog.Info("Attempting websocket connection")
		conn, err = connectWebsocket(url)
		if err != nil {
			slog.Error("Error connecting to websocket", "err", err.Error())
			time.Sleep(time.Duration(WebsocketTimeout) * time.Second)
			continue
		}
		quit := handleWebsocket(conn, cp)
		if quit {
			return
		}
	}
}

// returns true if parent should quit
func handleWebsocket(conn *websocket.Conn, cp ChanPkg) bool {
	defer conn.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sig:
			slog.Warn("Instruction to quit received, shutting down")
			for i := 0; i < ByteWorker; i++ {
				cp.Cancel <- true
			}
			return true
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return false
			}
			cp.ByteSlice <- message
		}
	}
}

func connectWebsocket(url string) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Start(cnf *Config, cp ChanPkg) error {

	var workers int = ByteWorker
	var wg sync.WaitGroup
	slog.Debug("Starting byte handlers")
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go handleBytes(cnf, &wg, i, cp)
	}

	stream := fmt.Sprintf("%s%s%s", ServerArgsPre, cnf.JetStreamServer, ServerArgsPost)
	slog.Debug("Starting the websocket listener")
	go bpWebsocket(cp, &wg, stream)

	// Only report buffer status during debug
	if LogLevel == slog.LevelDebug {
		go reportBsChanBuf(cp)
	}

	wg.Wait()
	slog.Info("Bot shutdown complete")
	return nil
}

func handleBytes(cnf *Config, wg *sync.WaitGroup, id int, cp ChanPkg) {
	slog.Debug("Worker starting", "WorkerId", id)
	for {
		select {
		case <-cp.Cancel:
			slog.Info("Cancel received", "WorkerId", id)
			wg.Done()
			return
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
		slog.Info("Found a match", "WorkerId", id, "Msg", msg.Commit.Record.Text)
		//TODO new post office logic user client/server model
		err := Ral(cnf, msg, cp)
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
