package backend

type RunningData struct {
	PreLogin     PreLoginResponse
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	MasterKey    []byte `json:"master_key"`
	Sync         SyncResponse
}

var GlobalData RunningData
var secrets secretCache
