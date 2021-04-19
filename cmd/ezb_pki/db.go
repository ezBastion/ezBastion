package main

import (
	m "ezBastion/cmd/ezb_pki/Models"
	"ezBastion/pkg/confmanager"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"path"
)

type GormLogger struct{}

func (*GormLogger) Print(v ...interface{}) {
	if v[0] == "sql" {
		log.WithFields(log.Fields{"module": "gorm", "type": "sql"}).Print(v[3])
	}
	if v[0] == "log" {
		log.WithFields(log.Fields{"module": "gorm", "type": "log"}).Print(v[2])
	}
}
func InitDB(conf confmanager.Configuration, exePath string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	db, err = gorm.Open("sqlite3", path.Join(exePath, "db/ca.db"))

	if err != nil {
		log.Fatal("sql.Open err: %s\n", err)
		return nil, err
	}
	db.Exec("PRAGMA foreign_keys = OFF")
	db.SetLogger(&GormLogger{})
	db.SingularTable(true)
	if !db.HasTable(&m.CSREntry{}) {
		db.CreateTable(&m.CSREntry{})
	}
	db.AutoMigrate(&m.CSREntry{})
	return db, nil
}
