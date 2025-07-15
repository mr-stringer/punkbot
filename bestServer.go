package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Config) FindFastestServer(ml MeasureLatencyType) error {

	servers := []string{
		"jetstream1.us-east.bsky.network", // /subscribe?wantedCollections=app.bsky.feed.post",
		"jetstream2.us-east.bsky.network",
		"jetstream1.us-west.bsky.network",
		"jetstream2.us-west.bsky.network",
	}

	lowest := time.Duration(0)
	lowestPtr := &lowest
	best := ""
	for _, v := range servers {

		t, err := ml(fmt.Sprintf("%s%s%s", ServerArgsPre, v, ServerArgsPost))

		if err != nil {
			slog.Error("Problem checking latency", "server", v, "err", err.Error())
			continue
		}
		if lowest == 0 || *t < *lowestPtr {
			best = v
			lowest = *t
		}
		slog.Info("Server latency", "server", v, "Milliseconds", t.Milliseconds())
	}

	if best == "" {
		slog.Error("All servers failed to provide a response")
		return fmt.Errorf("all servers failed to provide a response")
	}

	slog.Info("Best server found", "server", best)
	c.JetStreamServer = best

	return nil
}

type MeasureLatencyType func(address string) (*time.Duration, error)

func MeasureLatency(address string) (*time.Duration, error) {
	// Create a timeout for the connection attempt
	startTime := time.Now()

	conn, _, err := websocket.DefaultDialer.Dial(address, http.Header{})
	if err != nil {
		slog.Error("Failed to open websocket", "err", err.Error())
		return nil, err
	}
	conn.Close()

	// Calculate the latency
	latency := time.Since(startTime)
	return &latency, nil
}
