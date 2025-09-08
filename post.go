package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

func sessionServer(tm TokenManagerInt, ctx context.Context, wg *sync.WaitGroup, cnf *Config, cp ChanPkg, tr time.Duration) {
	defer wg.Done()

	slog.Info("Starting")
	slog.Info("Creating Session")
	createUrl := fmt.Sprintf("%s/%s", ApiUrl, CreateSessionEndpoint)
	RefreshUrl := fmt.Sprintf("%s/%s", ApiUrl, RefreshEndpoint)

	d, err := tm.getToken(cnf, createUrl)
	if err != nil {
		slog.Error("Failed to start session", "err", err.Error())
		/*Don't clean up, just exit*/
		cp.Exit <- ExitGetToken
		return
	}
	slog.Info("Got Token")

	/*Configure refresh ticker*/
	ticker := time.NewTicker(tr)

	for {
		slog.Debug("In the loop", "AccessTokenHash", StrHash(d.AccessJwt))
		select {
		case <-cp.ReqDidResp:
			slog.Debug("Request for session", "AccessTokenHash", StrHash(d.AccessJwt))
			// dereference to send copy of DID Response
			cp.Session <- *d

		/* When the ticker sends to the channel, it's time to refresh the     */
		/*token                                                               */
		case <-ticker.C:
			slog.Debug("Attempting to refresh access token")
			err = tm.getRefresh(&d, RefreshUrl)
			slog.Error("Refreshing token failed")
			if err != nil {
				/*right now on error, the code quits, that's OK docker can    */
				/* restart it, but I might come back and improve later.       */
				cp.Exit <- ExitRefreshToken
				return
			}

		case <-ctx.Done():
			/* No need to decrement the wait groups, it's already deferred    */
			slog.Info("Shutting down")
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
