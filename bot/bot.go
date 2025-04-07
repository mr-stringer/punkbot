package bot

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

	"github.com/gorilla/websocket"
	"github.com/mr-stringer/punkbot/config"
	"github.com/mr-stringer/punkbot/global"
	"github.com/mr-stringer/punkbot/postoffice"
)

func bpWebsocket(chSig <-chan os.Signal, chCancel chan<- bool, chBa chan<- []byte, wg *sync.WaitGroup, url string) {
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
			time.Sleep(time.Duration(global.WebsocketTimeout) * time.Second)
			continue
		}
		quit := handleWebsocket(conn, chBa, chCancel, chSig)
		if quit {
			return
		}
	}
}

// returns true if parent should quit
func handleWebsocket(conn *websocket.Conn, chBa chan<- []byte, chCancel chan<- bool, chSig <-chan os.Signal) bool {
	defer conn.Close()
	for {
		select {
		case <-chSig:
			slog.Warn("Instruction to quit received, shutting down")
			for i := 0; i < global.ByteWorker; i++ {
				chCancel <- true
			}
			return true
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return false
			}
			chBa <- message
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

func Start(cnf *config.Config) error {

	/* There is a problem here, if the websocket glitches the whole program   /*
	/* fails. A possible workaround would be to look out for the failure and  /*
	/* restart it. That would probably require a dedicated go-routine
	the */
	chBa := make(chan []byte)
	chCancel := make(chan bool)
	chSig := make(chan os.Signal, 1)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)

	var workers int = global.ByteWorker
	var wg sync.WaitGroup
	slog.Debug("Starting byte handlers")
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go handleBytes(cnf, &wg, i, chCancel, chBa)
	}

	stream := fmt.Sprintf("%s%s%s", global.ServerArgsPre, cnf.JetStreamServer, global.ServerArgsPost)
	slog.Debug("Starting the websocket listener")
	go bpWebsocket(chSig, chCancel, chBa, &wg, stream)

	wg.Wait()
	slog.Info("Bot shutdown complete")
	return nil
}

func handleBytes(cnf *config.Config, wg *sync.WaitGroup, id int, chCancel <-chan bool, chin <-chan []byte) {
	slog.Debug("Worker starting", "WorkerId", id)
	for {
		select {
		case <-chCancel:
			slog.Info("Cancel received", "WorkerId", id)
			wg.Done()
			return
		case ba := <-chin:
			slog.Debug("Data received", "WorkerId", id, "Length", len(ba))
			var msg global.Message
			err := json.Unmarshal(ba, &msg)
			if err != nil {
				slog.Error("Couldn't unmarshal message", "error", err.Error())
				return
			}
			err = handleMessage(cnf, id, &msg)
			if err != nil {
				slog.Warn("Problem handling message", "WorkerId", id, "error", err.Error())
				//A single message failure does not require us to quit
			}
		}
	}
}

func handleMessage(cnf *config.Config, id int, msg *global.Message) error {
	/* We only want to repost and like original posts, not replies. */
	/* Check if there is a reply path */
	if msg.Commit.Record.Reply.Parent.URI != "" {
		slog.Debug("Post is a reply, will not process", "WorkerId", id)
		return nil
	}
	if checkHashtags(cnf, msg) {
		slog.Info("Found a match", "WorkerId", id, "Msg", msg.Commit.Record.Text)
		err := postoffice.Ral(cnf, msg)
		if err != nil {
			slog.Error("Repost failed", "err", err.Error())
		}
	}

	return nil
}

func checkHashtags(cnf *config.Config, msg *global.Message) bool {
	/* Check if hastags are present in the message */
	for _, v := range cnf.Terms {
		if strings.Contains(strings.ToLower(msg.Commit.Record.Text), strings.ToLower(v)) {
			return true
		}
	}
	return false
}
