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
	"github.com/urfave/cli"
	"golang.org/x/sys/windows/svc"
	"os"
	"path"
	"strings"
)

var (
	exePath string
	conf    confmanager.Configuration
	err     error
)

const (
	VERSION         = "0.1.0"                                 //use semver.org
	SERVICENAME     = "ezb_sta"                               // name used as windows service name
	SERVICEFULLNAME = "ezBastion Secure Token Authorization." // windows service description
	CONFFILE        = "conf/config.toml"                      //config file path stay hardcoded
	LOGFILE         = "log/ezb_sta.log"                       // static path too
)

func init() {
	exePath, err = setupmanager.ExePath()
	if err != nil {
		log.Fatalf("Path error: %v", err)
	}
	// Get the current user
	userdomain := os.Getenv("USERDNSDOMAIN")
	cdomain := strings.Split(userdomain, ".")
	b_dn := ""
	for _, dcbloc := range cdomain {
		if b_dn != "" {
			b_dn += ","
		}
		b_dn += "dc=" + dcbloc
	}
	// Logonserver is like \\server, removing \\
	dcname := os.Getenv("LOGONSERVER")
	dcname = dcname[2 : len(dcname)-2]

	/*cfg := &ldap.Config{
		BaseDN:       b_dn,
		BindDN:       "cn=LDAP viewer,ou=Services,ou=Accounts,dc=EZB,dc=local",
		Port:         "389",
		Host:         dcname,
		BindPassword: "P@ssw0rd!EZB",
		Filter:       "(uid=%s)",
	}
	*/
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
	// ascii art url: http://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=ezBastion
	cli.AppHelpTemplate = fmt.Sprintf(`

███████╗███████╗██████╗  █████╗ ███████╗████████╗██╗ ██████╗ ███╗   ██╗
██╔════╝╚══███╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██║██╔═══██╗████╗  ██║
█████╗    ███╔╝ ██████╔╝███████║███████╗   ██║   ██║██║   ██║██╔██╗ ██║
██╔══╝   ███╔╝  ██╔══██╗██╔══██║╚════██║   ██║   ██║██║   ██║██║╚██╗██║
███████╗███████╗██████╔╝██║  ██║███████║   ██║   ██║╚██████╔╝██║ ╚████║
╚══════╝╚══════╝╚═════╝ ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
                                                                       
							███████╗████████╗ █████╗ 
							██╔════╝╚══██╔══╝██╔══██╗
							███████╗   ██║   ███████║
							╚════██║   ██║   ██╔══██║
							███████║   ██║   ██║  ██║
							╚══════╝   ╚═╝   ╚═╝  ╚═╝            
																			  
%s
INFO:
		http://www.ezbastion.com		
		support@ezbastion.com
		`, cli.AppHelpTemplate)
	app.Run(os.Args)
}
