package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var goodJson string = `{` +
	`"did": "did:plc:abcdefghijklmnopqrstuvwx",` +
	`"didDoc": {` +
	`  "@context": [` +
	`    "https://www.w3.org/ns/did/v1",` +
	`    "https://w3id.org/security/multikey/v1",` +
	`    "https://w3id.org/security/suites/secp256k1-2019/v1"` +
	`  ], ` +
	`  "id": "did:plc:abcdefghijklmnopqrstuvwx",` +
	`  "alsoKnownAs": [` +
	`    "at://atestuser.bsky.social"` +
	`  ],` +
	`  "verificationMethod": [` +
	`    {` +
	`      "id": "did:plc:abcdefghijklmnopqrstuvwx#atproto",` +
	`      "type": "Multikey",` +
	`      "controller": "did:plc:abcdefghijklmnopqrstuvwx",` +
	`      "publicKeyMultibase": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"` +
	`    }` +
	`  ],` +
	`  "service": [` +
	`    {` +
	`      "id": "#atproto_pds",` +
	`      "type": "AtprotoPersonalDataServer",` +
	`      "serviceEndpoint": "https://cortinarius.us-west.host.bsky.network"` +
	`    }` +
	`  ]` +
	`},` +
	`"handle": "atestuser.bsky.social",` +
	`"email": "example@example.com",` +
	`"emailConfirmed": true,` +
	`"emailAuthFactor": false,` +
	`"accessJwt": "f81b0e1dcd47a54b194af9efd084331055d1f9fb76289c5b91c74e59a80960a41ce637167a622c219b0d36f2bd599d08a28dbf629de6c6d684c5dc56d06d7de5",` +
	`"refreshJwt": "3cfcca922ae4c279e615d8bc0ceefb306526ee080bca1b961096309eda1dace63905bcc2300bc661c3e8850e79aa604103e719b54994d4db2c310d0e90aa742c",` +
	`"active": true` +
	`}`

func Test_getToken(t *testing.T) {
	goodCnf := Config{
		Identifier: "goodTest",
		password:   "blah-blah-blah-blah",
	}

	badCnf := Config{
		Identifier: "badTest",
		password:   "blah-blah-blah-blah",
	}

	notJsonCnf := Config{
		Identifier: "notJson",
		password:   "blah-blah-blah-blah",
	}

	var goodReturn DIDResponse
	err := json.Unmarshal([]byte(goodJson), &goodReturn)
	if err != nil {
		t.Errorf("couldn't unmarshal json string to DIDResponse for test")
		return
	}

	tstSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var tc tokenCreate
		err := dec.Decode(&tc)
		if err != nil {
			t.Errorf("couldn't decode json to string map")
		}
		switch tc.Identifier {
		case "goodTest":
			fmt.Fprint(w, goodJson)
		case "badTest":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Something went wrong"}`))
		}

	}))
	defer tstSrv.Close()

	type args struct {
		cnf *Config
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    *DIDResponse
		wantErr bool
	}{
		{"Good01", args{&goodCnf, tstSrv.URL}, &goodReturn, false},
		{"InternalServerError", args{&badCnf, tstSrv.URL}, nil, true},
		{"NotJsonError", args{&notJsonCnf, tstSrv.URL}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getToken(tt.args.cnf, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("getToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRefresh(t *testing.T) {

	var goodCurrentDID, badCurrentDID DIDResponse
	goodCurrentDID.RefreshJwt = "good"
	badCurrentDID.RefreshJwt = "bad"

	pGoodCurrentDID := &goodCurrentDID
	pBadCurrentDID := &badCurrentDID

	tstSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		str := r.Header.Get("Authorization")
		switch str {
		case "Bearer good":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, goodJson)
		case "Bearer bad":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Something went wrong"}`))
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "BadRequest"}`))
		}

	}))
	defer tstSrv.Close()

	type args struct {
		current **DIDResponse
		url     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Good01", args{&pGoodCurrentDID, tstSrv.URL}, false},
		{"Bad01", args{&pBadCurrentDID, tstSrv.URL}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := getRefresh(tt.args.current, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("getRefresh() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
