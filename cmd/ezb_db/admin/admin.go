//     This file is part of ezBastion.
//     ezBastion is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     ezBastion is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.

package admin

import (
	"crypto/sha256"
	"encoding/json"
	"ezBastion/pkg/confmanager"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"ezBastion/cmd/ezb_db/configuration"
	m "ezBastion/cmd/ezb_db/models"
	"ezBastion/cmd/ezb_db/tools"
	"ezBastion/pkg/setupmanager"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func FirstSTA(exePath string, conf confmanager.Configuration, staUrl string) error {
	var db *gorm.DB
	db, err := configuration.InitDB(conf, exePath)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	firstIAM := m.EzbStas{}
	//Name: "Default", Enable: true, Type: 0, Comment: "First IAM", EndPoint: staUrl, Issuer: "changeME", Default: true}
	db.FirstOrCreate(&firstIAM, m.EzbStas{Name: "Default", Enable: true, Type: 0, Comment: "First IAM", Default: true})
	firstIAM.EndPoint = staUrl
	firstIAM.Issuer = conf.EZBSTA.JWT.Issuer
	db.Save(firstIAM)
	fmt.Println("Sta added, you can login ezb_admin console.")
	return nil
}

func ResetPWD(exePath string, conf confmanager.Configuration) error {
	var db *gorm.DB
	db, err := configuration.InitDB(conf, exePath)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	newAdmin := fmt.Sprintf("adm%s", tools.RandString(6, "abcdefghijklmnopqrstuvwxyz"))
	newPasswd := tools.RandString(7, "")
	salt := tools.RandString(5, "")
	fmt.Println("A new admin account will be created.")
	for {
		c := setupmanager.AskForConfirmation("Create new admin account?")
		if c {
			break
		} else {
			return nil
		}
	}

	var Adm m.EzbAccounts
	defpwd := fmt.Sprintf("%x", sha256.Sum256([]byte(newPasswd+salt)))
	currentTime := time.Now()
	db.Where(m.EzbAccounts{Name: newAdmin}).Attrs(m.EzbAccounts{Enable: true, Comment: fmt.Sprintf("backup admin create by %s on %s", os.Getenv("USERNAME"), currentTime.Format("2006-01-02")), Salt: salt, Password: defpwd, Type: "i", Isadmin: true}).FirstOrCreate(&Adm)

	fmt.Println("Login with this new account to reset real one.")
	fmt.Printf("user: %s\n", newAdmin)
	fmt.Printf("password: %s\n", newPasswd)
	return nil
}

func DumpDB(exePath string, conf confmanager.Configuration) error {
	var db *gorm.DB
	db, err := configuration.InitDB(conf, exePath)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	f := map[string]interface{}{}

	var access []m.EzbAccess
	err = db.Find(&access).Error
	f["access"] = jsonEncode(access)

	var account []m.EzbAccounts
	err = db.Find(&account).Error
	f["account"] = jsonEncode(account)

	var action []m.EzbActions
	err = db.Find(&action).Error
	f["action"] = jsonEncode(action)

	var collection []m.EzbCollections
	err = db.Find(&collection).Error
	f["collection"] = jsonEncode(collection)

	var controller []m.EzbControllers
	err = db.Find(&controller).Error
	f["controller"] = jsonEncode(controller)

	var group []m.EzbGroups
	err = db.Find(&group).Error
	f["group"] = jsonEncode(group)

	var job []m.EzbJobs
	err = db.Find(&job).Error
	f["job"] = jsonEncode(job)

	var tag []m.EzbTags
	err = db.Find(&tag).Error
	f["tag"] = jsonEncode(tag)

	var worker []m.EzbWorkers
	err = db.Find(&worker).Error
	f["worker"] = jsonEncode(worker)

	var sta []m.EzbStas
	err = db.Find(&sta).Error
	f["sta"] = jsonEncode(sta)

	var bastion []m.EzbBastions
	err = db.Find(&bastion).Error
	f["bastion"] = jsonEncode(bastion)

	var license []m.EzbLicense
	err = db.Find(&license).Error
	f["license"] = jsonEncode(license)

	c, err := json.Marshal(f)
	statusFile := filepath.Join(exePath, "dbdump.json")
	err = ioutil.WriteFile(statusFile, c, 0600)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	fmt.Println("Database save to", statusFile)
	fmt.Println("/!\\ Sensitive data is in this file. /!\\")
	return nil
}
func jsonEncode(v interface{}) string {
	s, _ := json.Marshal(v)
	return string(s)
}
func RestoreDB(exePath string, conf confmanager.Configuration) error {

	var db *gorm.DB
	db, err := configuration.InitDB(conf, exePath)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	dbFile := filepath.Join(exePath, "dbdump.json")

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Fatal("dump file not found")
		panic(err)
	}

	raw, err := ioutil.ReadFile(dbFile)
	if err != nil {
		log.Fatal("error reading task dump file ", dbFile)
		panic(err)
	}

	var dumpDB map[string]string
	err = json.Unmarshal(raw, &dumpDB)

	var access []m.EzbAccess
	err = json.Unmarshal([]byte(dumpDB["access"]), &access)
	for _, a := range access {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var accounts []m.EzbAccounts
	err = json.Unmarshal([]byte(dumpDB["account"]), &accounts)
	for _, a := range accounts {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var action []m.EzbActions
	err = json.Unmarshal([]byte(dumpDB["action"]), &action)
	for _, a := range action {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var collection []m.EzbCollections
	err = json.Unmarshal([]byte(dumpDB["collection"]), &collection)
	for _, a := range collection {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var controller []m.EzbControllers
	err = json.Unmarshal([]byte(dumpDB["controller"]), &controller)
	for _, a := range controller {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var group []m.EzbGroups
	err = json.Unmarshal([]byte(dumpDB["group"]), &group)
	for _, a := range group {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var job []m.EzbJobs
	err = json.Unmarshal([]byte(dumpDB["job"]), &job)
	for _, a := range job {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var tag []m.EzbTags
	err = json.Unmarshal([]byte(dumpDB["tag"]), &tag)
	for _, a := range tag {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var worker []m.EzbWorkers
	err = json.Unmarshal([]byte(dumpDB["worker"]), &worker)
	for _, a := range worker {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var sta []m.EzbStas
	err = json.Unmarshal([]byte(dumpDB["sta"]), &sta)
	for _, a := range sta {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var bastion []m.EzbBastions
	err = json.Unmarshal([]byte(dumpDB["bastion"]), &bastion)
	for _, a := range bastion {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	var license []m.EzbLicense
	err = json.Unmarshal([]byte(dumpDB["license"]), &license)
	for _, a := range license {
		fmt.Print(".")
		if db.NewRecord(&a) {
			fmt.Println(a, "not found")
			if err := db.Create(&a).Error; err != nil {
				panic(err)
			}
		} else {
			if err := db.Save(&a).Error; err != nil {
				panic(err)
			}
		}
	}

	return nil
}
