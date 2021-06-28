package models

import "github.com/dgrijalva/jwt-go"

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
