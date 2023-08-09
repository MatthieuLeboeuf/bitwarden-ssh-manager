package backend

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/spf13/viper"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type LoginResponse struct {
	AccessToken    string `json:"access_token"`
	ExpiresIn      int    `json:"expires_in"`
	TokenType      string `json:"token_type"`
	RefreshToken   string `json:"refresh_token"`
	TwoFactorToken string `json:"TwoFactorToken"`
}

type ErrorLoginResponse struct {
	Message          string `json:"Message"`
	ErrorDescription string `json:"error_description"`
}

// Login return status int
func Login(password string, otpCode int, otpRemember bool) int {
	email := viper.GetString("vault.email")

	masterKey, _ := deriveMasterKey(
		[]byte(password),
		email,
		GlobalData.PreLogin.Kdf,
		GlobalData.PreLogin.KdfIterations,
		GlobalData.PreLogin.KdfMemory,
		GlobalData.PreLogin.KdfParallelism,
	)

	hashedPassword := base64.StdEncoding.Strict().EncodeToString(pbkdf2.Key(masterKey, []byte(password),
		1, 32, sha256.New))

	values := urlValues(
		"grant_type", "password",
		"username", email,
		"password", hashedPassword,
		"scope", "api offline_access",
		"client_id", "connector",
		"deviceType", deviceType(),
		"deviceName", viper.GetString("device.name"),
		"deviceIdentifier", viper.GetString("device.identifier"),
	)

	totpToken := viper.GetString("vault.totp")
	if otpCode != 0 || totpToken != "" {
		if totpToken != "" {
			values.Set("twoFactorProvider", "5")
			values.Set("twoFactorToken", totpToken)
		} else {
			values.Set("twoFactorProvider", "0")
			values.Set("twoFactorToken", strconv.Itoa(otpCode))
		}
		if otpRemember || totpToken != "" {
			values.Set("twoFactorRemember", "1")
		} else {
			values.Set("twoFactorRemember", "0")
		}
	}

	baseUrl := viper.GetString("vault.url")

	r, _ := http.NewRequest("POST", baseUrl+"/identity/connect/token", strings.NewReader(values.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, _ := client.Do(r)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode == 400 {
		resData := &ErrorLoginResponse{}
		_ = json.NewDecoder(res.Body).Decode(resData)
		if resData.ErrorDescription == "Two factor required." {
			return LoginWaitOtp
		} else if strings.Contains(resData.Message, "Invalid TOTP code") {
			return LoginInvalidOtp
		} else if resData.Message == "Username or password is incorrect. Try again" {
			return LoginInvalidPassword
		} else {
			return LoginError
		}
	} else if res.StatusCode == 429 {
		return LoginToManyRequests
	}

	resData := &LoginResponse{}
	_ = json.NewDecoder(res.Body).Decode(resData)

	GlobalData.AccessToken = resData.AccessToken
	GlobalData.RefreshToken = resData.RefreshToken
	GlobalData.ExpiresIn = resData.ExpiresIn
	GlobalData.TokenType = resData.TokenType
	GlobalData.MasterKey = masterKey
	if (otpCode != 0 && otpRemember) || totpToken != "" {
		viper.Set("vault.totp", resData.TwoFactorToken)
		_ = viper.WriteConfig()
	}

	return LoginOk
}
