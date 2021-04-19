package Models

import (
	"github.com/jinzhu/gorm"
)

type CSREntry struct {
	gorm.Model
	UUID         string `gorm:"not null;unique;index:uuid" json:"uuid"`
	Name         string `json:"name"`
	Signed       int    `json:"signed"`
	SerialNumber string `gorm:"not null;unique;index:serialnumber" json:"serialnumber"`
}
