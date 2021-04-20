//go:generate  goversioninfo -64 -platform-specific=false

// This file is part of ezBastion.
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

package main

import (
	"ezBastion/cmd/ezb_pki/Models"
	"ezBastion/pkg/confmanager"
	"ezBastion/pkg/ez_cli"
	"ezBastion/pkg/logmanager"
	"ezBastion/pkg/servicemanager"
	"ezBastion/pkg/setupmanager"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"os"
	"path"

	"github.com/urfave/cli"
	"golang.org/x/sys/windows/svc"
)

var (
	exePath string
	conf    confmanager.Configuration
	err     error
	db      *gorm.DB
)

const (
	VERSION         = "1.0.0"
	SERVICENAME     = "ezb_pki"
	SERVICEFULLNAME = "ezBastion internal PKI"
	CONFFILE        = "conf/config.toml"
	LOGFILE         = "log/ezb_pki.log"
)

func init() {
	exePath, err = setupmanager.ExePath()
	if err != nil {
		log.Fatalf("Path error: %v", err)
	}
	db, err = InitDB(conf, exePath)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func main() {
	//All hardcoded path MUST be ONLY in main.go, it's bad enough.
	confPath := path.Join(exePath, CONFFILE)
	conf, err = confmanager.CheckConfig(confPath, exePath)
	if err == nil {
		IsWindowsService, err := svc.IsWindowsService()
		if err != nil {
			log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
		}
		logmanager.SetLogLevel(conf.Logger.LogLevel, exePath, LOGFILE, conf.Logger.MaxSize, conf.Logger.MaxBackups, conf.Logger.MaxAge, IsWindowsService)
		if IsWindowsService {
			servicemanager.RunService(SERVICENAME, false, mainService{})
			return
		}
	}
	app := cli.NewApp()
	app.Name = SERVICENAME
	app.Version = VERSION
	app.Usage = SERVICEFULLNAME
	app.Commands = ez_cli.EZCli(SERVICENAME, SERVICEFULLNAME, exePath, confPath, mainService{})

	cli.AppHelpTemplate = fmt.Sprintf(`

		███████╗███████╗██████╗  █████╗ ███████╗████████╗██╗ ██████╗ ███╗   ██╗
		██╔════╝╚══███╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██║██╔═══██╗████╗  ██║
		█████╗    ███╔╝ ██████╔╝███████║███████╗   ██║   ██║██║   ██║██╔██╗ ██║
		██╔══╝   ███╔╝  ██╔══██╗██╔══██║╚════██║   ██║   ██║██║   ██║██║╚██╗██║
		███████╗███████╗██████╔╝██║  ██║███████║   ██║   ██║╚██████╔╝██║ ╚████║
		╚══════╝╚══════╝╚═════╝ ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝

								██████╗ ██╗  ██╗██╗
								██╔══██╗██║ ██╔╝██║
								██████╔╝█████╔╝ ██║
								██╔═══╝ ██╔═██╗ ██║
								██║     ██║  ██╗██║
								╚═╝     ╚═╝  ╚═╝╚═╝

%s
INFO:
		https://www.ezbastion.com
		support@ezbastion.com
		`, cli.AppHelpTemplate)

	app.Commands = append(app.Commands, cli.Command{
		Name:  "cert",
		Usage: "Certificate management",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "Show unaccepted certificate request.",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "all,a",
						Usage: "Show all certificate.",
					},
				},
				Action: func(c *cli.Context) error {
					csrs := []Models.CSREntry{}
					db.Find(&csrs)
					for _, csr := range csrs {
						tag := "-"
						if csr.Signed == 1 {
							tag = "+"
						}
						if c.Bool("all") {
							fmt.Println(tag, csr.Name, " (uuid) ", csr.UUID, "from", csr.CreatedAt.Format("2006-01-02T15:04:05-0700"))
						} else {
							if csr.Signed == 0 {
								fmt.Println(tag, csr.Name, " (uuid) ", csr.UUID, "from", csr.CreatedAt.Format("2006-01-02T15:04:05-0700"))
							}
						}
					}
					return nil
				},
			},
			{
				Name:      "sign",
				Usage:     "Set ready to sign a certificate request.",
				ArgsUsage: "uuid",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "all,a",
						Usage: "Accept all pending certificate request.",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("all") {
						db.Model(&Models.CSREntry{}).Where("signed = 0").Update("signed", 1)
					} else {
						if len(c.Args()) == 1 {
							db.Model(&Models.CSREntry{}).Where("uuid= ?", c.Args().First()).Update("signed", 1)
						}
					}
					return nil
				},
			},
			{
				Name:      "clean",
				Usage:     "remove a certificate request",
				ArgsUsage: "uuid",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "all,a",
						Usage: "Remove all pending certificate request.",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("all") {
						db.Where("signed = 0").Delete(&Models.CSREntry{})
					} else {
						if len(c.Args()) == 1 {
							db.Where("uuid= ?", c.Args().First()).Delete(&Models.CSREntry{})
						}
					}
					return nil
				},
			},
		},

		// sign, clean, -a, revoke

	})

	app.Run(os.Args)
}
