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
	"time"
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

func GetBearers(s cache.Storage, conf *confmanager.Configuration) (bearers []Bearer, err error) {
	content := s.Get("bearers")
	if content != nil {
		byteCtrl := bytes.NewBuffer(content)
		dec := gob.NewDecoder(byteCtrl)
		err := dec.Decode(&bearers)
		if err != nil {
			fmt.Println(err)
			return bearers, err
		}
	} else {
		if len(bearers) == 0 {
			return bearers, fmt.Errorf("BEARERS NOT FOUND")
		}
		var byteCtrl bytes.Buffer
		enc := gob.NewEncoder(&byteCtrl)
		encErr := enc.Encode(bearers)
		if encErr != nil {
			return bearers, encErr
		}
		s.Set("bearers", byteCtrl.Bytes(), time.Duration(conf.EZBSTA.JWT.TTL)*time.Second)
	}
	return bearers, nil
}
