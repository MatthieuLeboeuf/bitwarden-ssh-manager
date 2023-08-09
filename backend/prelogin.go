package backend

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/viper"
	"io"
	"net/http"
)

type PreLoginRequest struct {
	Email string `json:"email"`
}

type PreLoginResponse struct {
	Kdf            KDFType `json:"Kdf"`
	KdfIterations  int     `json:"KdfIterations"`
	KdfMemory      int     `json:"KdfMemory"`
	KdfParallelism int     `json:"KdfParallelism"`
}

func PreLogin() {
	baseUrl := viper.GetString("vault.url")

	reqData := PreLoginRequest{
		Email: viper.GetString("vault.email"),
	}
	reqDataMarshalled, _ := json.Marshal(reqData)

	r, _ := http.NewRequest("POST", baseUrl+"/identity/accounts/prelogin", bytes.NewReader(reqDataMarshalled))
	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, _ := client.Do(r)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	resData := &PreLoginResponse{}
	_ = json.NewDecoder(res.Body).Decode(resData)

	GlobalData.PreLogin = *resData
}
