package postoffice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/mr-stringer/punkbot/config"
	"github.com/mr-stringer/punkbot/global"
)

// The purpose of PreFlightCheck is to ensure a JwtToken can be retrieved from
// the API, if not, we quit.
func PreFlightCheck(cnf *config.Config) error {
	slog.Info("Starting the postoffice")
	slog.Info("Get token")
	/*Getting a token checks we can authenticate, this saves us from waiting for
	a period for a hashtag and failing later*/
	d, err := getToken(cnf)
	if err != nil {
		slog.Error("Failed to get token")
		return err
	}
	slog.Info("Got token", "token", d.AccessJwt)
	return nil
}

func DIDResponseServer(cnf *config.Config, cp global.ChanPkg) {
	slog.Info("Starting the DID Response server")
	slog.Info("Getting token")
	d, err := getToken(cnf)
	if err != nil {
		slog.Error("Failed to get token")
		os.Exit(global.ExitGetToken)
	}

	/*Configure refresh ticker*/
	ticker := time.NewTicker(time.Second * 60)

	for {
		slog.Debug("DIDResponseServer, in the loop", "AccessTokenHash", global.StrHash(d.AccessJwt))
		select {
		case <-cp.ReqDidResp:
			slog.Info("Request for DID Response", "AccessTokenHash", global.StrHash(d.AccessJwt))
			// dereference to send copy of DID Response
			cp.DIDResp <- *d
		case <-ticker.C:
			slog.Debug("Attempting to refresh access token")
			for i := 0; i < global.TokenRefreshAttempts; i++ {
				err = getRefresh(&d)
				if err != nil {
					slog.Warn("Failed to refresh token", "Attempt", i+1)
				} else {
					// if err != nil, we can break out of the retry loop
					break
				}
				time.Sleep(time.Second * global.TokenRefreshTimeout)
			}
			if err != nil {
				slog.Error("Could not refresh token", "error", err.Error())
				os.Exit(global.ExitGetToken)
			}
		}
	}

}

// Ral (RePost and Like), uses the configured user to re-posts and like the
// provided message
func Ral(cnf *config.Config, msg *global.Message, cp global.ChanPkg) error {
	//Request a copy of the latest DID Response
	cp.ReqDidResp <- true
	dr := <-cp.DIDResp

	uri := fmt.Sprintf("at://%s/app.bsky.feed.post/%s", msg.DID, msg.Commit.RKey)

	resource := &global.CreateRecordProps{
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

func createRecord(r *global.CreateRecordProps) error {
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

	url := fmt.Sprintf("%s/%s", global.ApiUrl, global.CreatePostEndpoint)
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
