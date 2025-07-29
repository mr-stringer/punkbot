package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

func sessionServer(ctx context.Context, wg *sync.WaitGroup, cnf *Config, cp ChanPkg) {
	defer wg.Done()

	slog.Info("Starting the SessionServer")
	slog.Info("Creating Session")
	d, err := getToken(cnf)
	if err != nil {
		slog.Error("Failed to start session", "err", err.Error())
		/*Don't clean up, just exit*/
		os.Exit(ExitGetToken)
	}

	/*Configure refresh ticker*/
	ticker := time.NewTicker(time.Second * 60)

	for {
		slog.Debug("sessionServer, in the loop", "AccessTokenHash", StrHash(d.AccessJwt))
		select {
		case <-cp.ReqDidResp:
			slog.Info("Request for session", "AccessTokenHash", StrHash(d.AccessJwt))
			// dereference to send copy of DID Response
			cp.Session <- *d
		case <-ticker.C:
			slog.Debug("Attempting to refresh access token")
			for i := 0; i < TokenRefreshAttempts; i++ {
				err = getRefresh(&d)
				if err != nil {
					slog.Warn("Failed to refresh token", "Attempt", i+1)
				} else {
					// if err != nil, we can break out of the retry loop
					break
				}
				time.Sleep(time.Second * TokenRefreshTimeout)
			}
			if err != nil {
				slog.Error("Could not refresh token", "error", err.Error())
				os.Exit(ExitGetToken)
			}
		case <-ctx.Done():
			/* No need to decrement the wait groups, it's already deferred    */
			slog.Info("SessionServer shutting down")
			return
		}
	}

}

// Ral (RePost and Like), uses the configured user to re-posts and like the
// provided message
func Ral(cnf *Config, msg *Message, cp ChanPkg) error {
	//Request a copy of the latest session
	cp.ReqDidResp <- true
	dr := <-cp.Session

	uri := fmt.Sprintf("at://%s/app.bsky.feed.post/%s", msg.DID, msg.Commit.RKey)

	resource := &CreateRecordProps{
		DIDResponse: &dr,
		Resource:    "app.bsky.feed.repost",
		URI:         uri,
		CID:         msg.Commit.CID,
	}

	err := createRecord(resource)
	if err != nil {
		slog.Error("Error creating record", "error", err, "resource", resource.Resource)
		return err
	}

	resource.Resource = "app.bsky.feed.like"
	err = createRecord(resource)
	if err != nil {
		slog.Error("Error creating record", "error", err, "resource", resource.Resource)
		return err
	}

	return nil
}

func createRecord(r *CreateRecordProps) error {
	body := map[string]interface{}{
		"$type":      r.Resource,
		"collection": r.Resource,
		"repo":       r.DIDResponse.DID,
		"record": map[string]interface{}{
			"subject": map[string]interface{}{
				"uri": r.URI,
				"cid": r.CID,
			},
			"createdAt": time.Now(),
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		slog.Error("Error marshalling request", "error", err, "resource", r.Resource)
		return err
	}

	url := fmt.Sprintf("%s/%s", ApiUrl, CreatePostEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		slog.Error("Error creating request", "error", err, "r.Resource", r.Resource)
		return nil
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.DIDResponse.AccessJwt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error sending request", "error", err, "r.Resource", r.Resource)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code", "status", resp, "r.Resource", r.Resource)
		return nil
	}

	slog.Info("Published successfully", "resource", r.Resource)

	return nil
}
