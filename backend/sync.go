package backend

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Profile struct {
	Name       string
	Key        CipherString
	PrivateKey CipherString
}

type Folder struct {
	ID   uuid.UUID
	Name CipherString
}

type CipherString struct {
	Type CipherStringType

	IV, CT, MAC []byte
}

type Cipher struct {
	ID       uuid.UUID
	Name     CipherString
	Type     int
	FolderId uuid.UUID `json:",omitempty"`
	Login    struct {
		Username CipherString `json:",omitempty"`
		Password CipherString `json:",omitempty"`
		Uri      CipherString `json:",omitempty"`
	} `json:"Login"`
	Fields []struct {
		Name  CipherString
		Type  int
		Value CipherString
	} `json:"Fields"`
}

type SyncResponse struct {
	Profile Profile
	Folders []Folder
	Ciphers []Cipher
}

type SyncFrontend struct {
	Hosts   []FrontendHost   `json:"hosts"`
	Folders []FrontendFolder `json:"folders"`
}

type FrontendHost struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Port     string `json:"port"`
	Folder   string `json:"folder"`
}

type FrontendFolder struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Hosts int    `json:"hosts"`
}

type CipherStringType int

const (
	AesCbc256_B64                     CipherStringType = 0
	AesCbc128_HmacSha256_B64          CipherStringType = 1
	AesCbc256_HmacSha256_B64          CipherStringType = 2
	Rsa2048_OaepSha256_B64            CipherStringType = 3
	Rsa2048_OaepSha1_B64              CipherStringType = 4
	Rsa2048_OaepSha256_HmacSha256_B64 CipherStringType = 5
	Rsa2048_OaepSha1_HmacSha256_B64   CipherStringType = 6
)

var b64enc = base64.StdEncoding.Strict()

func (t CipherStringType) HasMAC() bool {
	return t != AesCbc256_B64
}

func (s CipherString) IsZero() bool {
	return s.Type == 0 && s.IV == nil && s.CT == nil && s.MAC == nil
}

func (s CipherString) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s CipherString) String() string {
	if s.IsZero() {
		return ""
	}
	if !s.Type.HasMAC() {
		return fmt.Sprintf("%d.%s|%s",
			s.Type,
			b64enc.EncodeToString(s.IV),
			b64enc.EncodeToString(s.CT),
		)
	}
	return fmt.Sprintf("%d.%s|%s|%s",
		s.Type,
		b64enc.EncodeToString(s.IV),
		b64enc.EncodeToString(s.CT),
		b64enc.EncodeToString(s.MAC),
	)
}

func (s *CipherString) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	i := bytes.IndexByte(data, '.')
	if i < 0 {
		return fmt.Errorf("cipher string does not contain a type: %q", data)
	}
	typStr := string(data[:i])
	var err error
	if t, err := strconv.Atoi(typStr); err != nil {
		return fmt.Errorf("invalid cipher string type: %q", typStr)
	} else {
		s.Type = CipherStringType(t)
	}
	switch s.Type {
	case AesCbc128_HmacSha256_B64, AesCbc256_HmacSha256_B64, AesCbc256_B64:
	default:
		return fmt.Errorf("unsupported cipher string type: %d", s.Type)
	}

	data = data[i+1:]
	parts := bytes.Split(data, []byte("|"))
	wantParts := 3
	if !s.Type.HasMAC() {
		wantParts = 2
	}
	if len(parts) != wantParts {
		return fmt.Errorf("cipher string type requires %d parts: %q", wantParts, data)
	}

	// TODO: do a single []byte allocation for all fields
	if s.IV, err = b64decode(parts[0]); err != nil {
		return err
	}
	if s.CT, err = b64decode(parts[1]); err != nil {
		return err
	}
	if s.Type.HasMAC() {
		if s.MAC, err = b64decode(parts[2]); err != nil {
			return err
		}
	}
	return nil
}

func b64decode(src []byte) ([]byte, error) {
	dst := make([]byte, b64enc.DecodedLen(len(src)))
	n, err := b64enc.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	dst = dst[:n]
	return dst, nil
}

func Sync() string {
	baseUrl := viper.GetString("vault.url")

	r, _ := http.NewRequest("GET", baseUrl+"/api/sync", nil)
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", GlobalData.TokenType+" "+GlobalData.AccessToken)
	r.URL.Query().Add("excludeDomains", "true")

	client := &http.Client{}
	res, _ := client.Do(r)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	resData := &SyncResponse{}
	_ = json.NewDecoder(res.Body).Decode(resData)

	GlobalData.Sync = *resData

	frontData := &SyncFrontend{}
	frontData.Hosts = []FrontendHost{}
	frontData.Folders = []FrontendFolder{}

	var folders []string

	for i := range resData.Folders {
		folder := &resData.Folders[i]
		temp, _ := secrets.decryptStr(folder.Name, nil)
		if strings.HasPrefix(temp, "bsm-") {
			hosts := 0
			for f := range resData.Ciphers {
				cipher := &resData.Ciphers[f]
				if cipher.FolderId.String() == folder.ID.String() {
					hosts++
				}
			}
			frontData.Folders = append(frontData.Folders, FrontendFolder{
				Id:    folder.ID.String(),
				Name:  temp[4:],
				Hosts: hosts,
			})
			folders = append(folders, folder.ID.String())
		}
	}

	var logins []*Cipher
	for i := range resData.Ciphers {
		cipher := &resData.Ciphers[i]
		if slices.Contains(folders, cipher.FolderId.String()) {
			logins = append(logins, cipher)
		}
	}

	for _, cipher := range logins {

		host := &FrontendHost{}
		host.Name, _ = secrets.decryptStr(cipher.Name, nil)
		host.Host, _ = secrets.decryptStr(cipher.Login.Uri, nil)
		host.Username, _ = secrets.decryptStr(cipher.Login.Username, nil)
		host.Password, _ = secrets.decryptStr(cipher.Login.Password, nil)
		host.Folder = cipher.FolderId.String()
		for _, field := range cipher.Fields {
			temp, _ := secrets.decryptStr(field.Name, nil)
			switch temp {
			case "port":
				host.Port, _ = secrets.decryptStr(field.Value, nil)
				break
			}
		}
		frontData.Hosts = append(frontData.Hosts, *host)
	}

	frontDataMarshalled, _ := json.Marshal(frontData)

	return string(frontDataMarshalled)
}
