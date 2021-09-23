package models

import (
	"bytes"
	"encoding/gob"
	"ezBastion/cmd/ezb_srv/cache"
	"ezBastion/pkg/confmanager"
	"fmt"
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

func GetJWT(s cache.Storage, conf *confmanager.Configuration, key string) (j *jwt.Token, err error) {
	j = new(jwt.Token)
	content := s.Get(key)
	if content == nil {
		byteCtrl := bytes.NewBuffer(content)
		dec := gob.NewDecoder(byteCtrl)
		err := dec.Decode(&j)
		if err != nil {
			fmt.Println(err)
			return j, err
		}
	}
	return j, nil
}
