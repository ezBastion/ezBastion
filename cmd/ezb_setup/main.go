//go:generate goversioninfo

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
	"ezBastion/pkg/confmanager"
	"ezBastion/pkg/setupmanager"
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
	"path"
)

var (
	exePath, confPath string
	conf              confmanager.Configuration
	err               error
	isVerbose         bool
)

const (
	//VERSION string - use semver.org
	VERSION = "0.1.0"
	//SERVICENAME string - name used as windows service name
	SERVICENAME = "ezb_setup"
	//SERVICEFULLNAME string - windows service description
	SERVICEFULLNAME = "ezBastion setup tooling."
	//CONFFILE string - config file path stay hardcoded
	CONFFILE = "conf/config.toml"
)

func init() {
	exePath, err = setupmanager.ExePath()
	if err != nil {
		log.Fatalf("Path error: %v", err)
	}
	confPath = path.Join(exePath, CONFFILE)
}

func main() {
	//All hardcoded path MUST be ONLY in main.go, it's bad enough.

	app := cli.NewApp()
	app.Name = SERVICENAME
	app.Version = VERSION
	app.Usage = SERVICEFULLNAME
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose, v",
			Destination: &isVerbose,
		},
	}
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, ver",
		Usage: "print the version",
	}
	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Generate default config file.",
			Action: func(c *cli.Context) error {
				startMsg("Get config file:")
				if err = setupmanager.Setup(exePath, confPath, SERVICENAME); err != nil {
					endMsg(err)
					return err
				}
				endMsg(nil)
				return nil
			},
		}, {
			Name:  "config",
			Usage: "Check config file structure.",
			Action: func(c *cli.Context) error {
				return checkToml()
			},
		}, {
			Name:  "pki",
			Usage: "Test ezb_pki TCP connection.",
			Action: func(c *cli.Context) error {
				conf, err = confmanager.CheckConfig(confPath, exePath)
				if err != nil {
					startMsg("Get config file:")
					endMsg(err)
					return err
				}
				return tcpPing(conf.EZBPKI.Network.FQDN, conf.EZBPKI.Network.Port)
			},
		}, {
			Name:  "install",
			Usage: "Add windows service.",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	}
	// ascii art url: http://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=ezBastion
	cli.AppHelpTemplate = fmt.Sprintf(`

███████╗███████╗██████╗  █████╗ ███████╗████████╗██╗ ██████╗ ███╗   ██╗
██╔════╝╚══███╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██║██╔═══██╗████╗  ██║
█████╗    ███╔╝ ██████╔╝███████║███████╗   ██║   ██║██║   ██║██╔██╗ ██║
██╔══╝   ███╔╝  ██╔══██╗██╔══██║╚════██║   ██║   ██║██║   ██║██║╚██╗██║
███████╗███████╗██████╔╝██║  ██║███████║   ██║   ██║╚██████╔╝██║ ╚████║
╚══════╝╚══════╝╚═════╝ ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
                                                                       
                ███████╗███████╗████████╗██╗   ██╗██████╗              
                ██╔════╝██╔════╝╚══██╔══╝██║   ██║██╔══██╗             
                ███████╗█████╗     ██║   ██║   ██║██████╔╝             
                ╚════██║██╔══╝     ██║   ██║   ██║██╔═══╝              
                ███████║███████╗   ██║   ╚██████╔╝██║                  
                ╚══════╝╚══════╝   ╚═╝    ╚═════╝ ╚═╝                            
																			  
%s
INFO:
		http://www.ezbastion.com		
		support@ezbastion.com
		`, cli.AppHelpTemplate)
	if err = app.Run(os.Args); err != nil {
		if !isVerbose {
			log.Fatal(err)
		}
	}

}
