package main

import "log/slog"

const (
	ExitCmdLineArgsFailure = iota
	ExitConfigFailure
	ExitNoPasswordSet
	ExitJetStreamFailure
	ExitPostOfficeFailure
	ExitBotFailure
	ExitGetToken
)

const ServerArgsPre string = "wss://"
const ServerArgsPost string = "/subscribe?wantedCollections=app.bsky.feed.post"
const ApiUrl string = "https://bsky.social/xrpc"
const CreateSessionEndpoint string = "com.atproto.server.createSession"
const CreatePostEndpoint string = "com.atproto.repo.createRecord"
const RefreshEndpoint string = "com.atproto.server.refreshSession"
const DidLookUpEndpoint string = "https://plc.directory"
const WebsocketTimeout int = 5
const ByteWorker int = 4
const TokenRefreshAttempts = 5
const TokenRefreshTimeout = 5
const ByteSliceBufferSize = 10

/*Build info*/
var ReleaseVersion string = "Development"
var BuildTime string

/* Global var for logging level & Debug Posts */
/* I don't really like global vars, but it gets written once and is read once */
/* so it should be safe */
var LogLevel slog.Level = slog.LevelInfo
var DebugPosts bool = false

type ClArgs struct {
	LogLevel       slog.Level
	ConfigFilePath string
	LogPath        string
	JsonLog        bool
}
type Config struct {
	Identifier string
	Terms      []string
	//Add jetstream instance, allowing users to set their preferred instance
	//Also add the ability to auto select public instance automatically based on
	//latency
	JetStreamServer string
	password        string //unexported

}

type Commit struct {
	CID        string `json:"cid"`
	Collection string `json:"collection"`
	Operation  string `json:"operation"`
	Record     Record `json:"record"`
	Rev        string `json:"rev"`
	RKey       string `json:"rkey"`
}

type Record struct {
	Type      string   `json:"$type"`
	CreatedAt string   `json:"createdAt"`
	Langs     []string `json:"langs"`
	Reply     Reply    `json:"reply"`
	Text      string   `json:"text"`
}

type Reply struct {
	Parent Parent `json:"parent"`
	Root   Parent `json:"root"`
}

type Parent struct {
	CID string `json:"cid"`
	URI string `json:"uri"`
}

type Message struct {
	Commit Commit `json:"commit"`
	DID    string `json:"did"`
	Kind   string `json:"kind"`
	TimeUs int64  `json:"time_us"`
}

type DIDResponse struct {
	DID             string `json:"did"`
	DIDDoc          DIDDoc `json:"didDoc"`
	Handle          string `json:"handle"`
	Email           string `json:"email"`
	EmailConfirmed  bool   `json:"emailConfirmed"`
	EmailAuthFactor bool   `json:"emailAuthFactor"`
	AccessJwt       string `json:"accessJwt"`
	RefreshJwt      string `json:"refreshJwt"`
	Active          bool   `json:"active"`
}

type DIDDoc struct {
	Context            []string `json:"@context"`
	ID                 string   `json:"id"`
	AlsoKnownAs        []string `json:"alsoKnownAs"`
	VerificationMethod []struct {
		ID                 string `json:"id"`
		Type               string `json:"type"`
		Controller         string `json:"controller"`
		PublicKeyMultibase string `json:"publicKeyMultibase"`
	} `json:"verificationMethod"`
	Service []struct {
		ID              string `json:"id"`
		Type            string `json:"type"`
		ServiceEndpoint string `json:"serviceEndpoint"`
	} `json:"service"`
}

type CreateRecordProps struct {
	DIDResponse *DIDResponse
	Resource    string
	URI         string
	CID         string
}

type TokenServer struct {
	Request  chan bool
	Response chan DIDResponse
}

type ChanPkg struct {
	ByteSlice  chan []byte
	ReqDidResp chan bool
	Session    chan DIDResponse
}
