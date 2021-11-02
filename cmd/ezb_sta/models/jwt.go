package models

import (
	"bytes"
	"encoding/gob"
	"ezBastion/cmd/ezb_srv/cache"
	"ezBastion/pkg/confmanager"
	"github.com/dgrijalva/jwt-go"
	"os"
	"path/filepath"
)

type Payload struct {
	*jwt.StandardClaims
	JTI string `json:"jti"`
	ISS string `json:"iss"`
	SUB string `json:"sub"`
	AUD string `json:"aud"`
	EXP int64  `json:"exp"`
	IAT int    `json:"iat"`
}

type Introspec struct {
	User       string   `json:"user"`
	UserSID    string   `json:"userSid"`
	UserGroups []string `json:"userGroups"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Bearer struct {
	ExpireIn    int    `json:"expire_in"`
	ExpireAt    int64  `json:"expire_at"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

var storage cache.Storage
var conf confmanager.Configuration
var exPath string

func init() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)
}

func GetUser(s cache.Storage, key string) (u *StaUser, err error) {
	u = new(StaUser)
	content := s.Get(key)
	if content != nil {
		byteCtrl := bytes.NewBuffer(content)
		dec := gob.NewDecoder(byteCtrl)
		err := dec.Decode(&u)
		if err == nil {
			return u, err
		}
	}
	return u, nil
}
