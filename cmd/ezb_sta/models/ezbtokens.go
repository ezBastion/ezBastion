package models

type EzbFormAuth struct {
	Grant_type string `json:"grant_type"`
	Username   string `json:"username"  `
	Password   string `json:"password"`
}
