package models

type StaUser struct {
	User       string `json:"user"`
	UserSid    string `json:"usersid"`
	UserGroups string `json:"usergroups"`
	Sign_key   string `json:"signkey"`
}
