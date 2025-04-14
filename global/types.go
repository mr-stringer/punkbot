package global

const (
	ExitCmdLineArgsFailure = iota
	ExitConfigFailure
	ExitNoPasswordSet
	ExitJetStreamFailure
	ExitPostOfficeFailure
	ExitBotFailure
)

const ServerArgsPre string = "wss://"
const ServerArgsPost string = "/subscribe?wantedCollections=app.bsky.feed.post"
const ApiUrl string = "https://bsky.social/xrpc"
const CreateSessionEndpoint string = "com.atproto.server.createSession"
const CreatePostEndpoint string = "com.atproto.repo.createRecord"
const WebsocketTimeout int = 5
const ByteWorker int = 4

/*Build info*/
var ReleaseVersion string = "Development"
var BuildTime string

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
