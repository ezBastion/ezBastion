package models

import "gopkg.in/ldap.v2"

type StaUser struct {
	User          string          `json:"user"`
	UserSid       string          `json:"usersid"`
	UserGroups    []string        `json:"usergroups"`
	Sign_key      string          `json:"signkey"`
	ExtProperties *IntrospectUser `json:"introspectuser"`
}

type IntrospectUser struct {
	Jti               string   `json:"jti"`
	Ou                string   `json:"ou"`
	Groups            []string `json:"groups"`
	Ntaccount         string   `json:"ntaccount"`
	Samaccountname    string   `json:"samaccountname"`
	Description       string   `json:"description"`
	Displayname       string   `json:"displayname"`
	Distinguishedname string   `json:"distinguishedname"`
	Emailaddress      string   `json:"emailaddress"`
	Givenname         string   `json:"givenname"`
}

type Ldapinfo struct {
	Base         string     `json:"base"`
	Host         string     `json:"host"`
	Port         int        `json:"port"`
	UseSSL       bool       `json:"usessl"`
	SkipTLS      bool       `json:"skiptls"`
	BindDN       string     `json:"binddn"`
	BindUser     string     `json:"binduser"`
	BindPassword string     `json:"bindpassword"`
	ServerName   string     `json:"servername"`
	LDAPcrt      string     `json:"ldapcrt"`
	LDAPpk       string     `json:"ldappk"`
	UserFilter   string     `json:"userfilter"`
	GroupFilter  string     `json:"groupfilter"`
	Attributes   []string   `json:"attributes"`
	LConn        *ldap.Conn `json:"lconn"`
}
