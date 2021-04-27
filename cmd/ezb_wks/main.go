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
	"ezBastion/pkg/ez_cli"
	"ezBastion/pkg/logmanager"
	"ezBastion/pkg/servicemanager"
	"ezBastion/pkg/setupmanager"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"

	"github.com/urfave/cli"
	"golang.org/x/sys/windows/svc"
)

var (
	exePath string
	conf    confmanager.Configuration
	err     error
)

const (
	VERSION         = "1.2.0"
	SERVICENAME     = "ezb_wks"
	SERVICEFULLNAME = "ezBastion worker service."
	CONFFILE        = "conf/config.toml"
	LOGFILE         = "log/ezb_wks.log"
)

func init() {
	exePath, err = setupmanager.ExePath()
	if err != nil {
		log.Fatalf("Path error: %v", err)
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

	app.Commands = ez_cli.EZCli(conf.EZBWKS.ServiceName, SERVICEFULLNAME, exePath, confPath, mainService{})

	cli.AppHelpTemplate = fmt.Sprintf(`

		███████╗███████╗██████╗  █████╗ ███████╗████████╗██╗ ██████╗ ███╗   ██╗
		██╔════╝╚══███╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██║██╔═══██╗████╗  ██║
		█████╗    ███╔╝ ██████╔╝███████║███████╗   ██║   ██║██║   ██║██╔██╗ ██║
		██╔══╝   ███╔╝  ██╔══██╗██╔══██║╚════██║   ██║   ██║██║   ██║██║╚██╗██║
		███████╗███████╗██████╔╝██║  ██║███████║   ██║   ██║╚██████╔╝██║ ╚████║
		╚══════╝╚══════╝╚═════╝ ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
																			   
							██╗    ██╗██╗  ██╗███████╗                         
							██║    ██║██║ ██╔╝██╔════╝                         
							██║ █╗ ██║█████╔╝ ███████╗                         
							██║███╗██║██╔═██╗ ╚════██║                         
							╚███╔███╔╝██║  ██╗███████║                         
							 ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝                         
																			  
%s
INFO:
		http://www.ezbastion.com		
		support@ezbastion.com
		`, cli.AppHelpTemplate)
	app.Run(os.Args)
}
