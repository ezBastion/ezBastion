package models

type StaUser struct {
	User       string `json:"user"`
	UserSid    string `json:"usersid"`
	UserGroups string `json:"usergroups"`
	Sign_key   string `json:"signkey"`
}

type ADUser struct {
	User       string `json:"user"`
	Domain     string `json:"domain"`
	UserSid    string `json:"usersid"`
	UserGroups string `json:"usergroups"`
}
