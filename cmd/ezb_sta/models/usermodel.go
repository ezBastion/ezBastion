package models

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
