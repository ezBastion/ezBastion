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
	"github.com/jtblin/go-ldap-client"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/sys/windows/svc"
	"os"
	"path"
)

var (
	exePath    string
	conf       confmanager.Configuration
	confPath   string
	conferr    error
	err        error
	staservice mainService
	ldapclient *ldap.LDAPClient
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
	confPath = path.Join(exePath, CONFFILE)
	conf, conferr = confmanager.CheckConfig(confPath, exePath)
	if conferr == nil {
		// successful bind *Conn is ready to be requested
		ldapclient = &ldap.LDAPClient{
			Base:         conf.EZBSTA.StaLdap.Base,
			Host:         conf.EZBSTA.StaLdap.Host,
			Port:         conf.EZBSTA.StaLdap.Port,
			UseSSL:       conf.EZBSTA.StaLdap.UseSSL,
			SkipTLS:      conf.EZBSTA.StaLdap.SkipTLS,
			BindDN:       conf.EZBSTA.StaLdap.BindDN,
			BindPassword: conf.EZBSTA.StaLdap.BindPassword,
			UserFilter:   "(cn=%s)",
			GroupFilter:  "(&(objectClass=group)(member=%s))",
			Attributes:   []string{"ou", "ntaccount", "samaccountname", "description", "displayname", "emailaddress", "givenname", "distinguishedName"},
		}
		staservice = mainService{STAldapauth: ldapclient}
	}
}

func main() {
	//All hardcoded path MUST be ONLY in main.go, it's bad enough.
	defer ldapclient.Close()

	if conferr == nil {
		IsWindowsService, err := svc.IsWindowsService()
		if err != nil {
			log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
		}
		logmanager.SetLogLevel(conf.Logger.LogLevel, exePath, LOGFILE, conf.Logger.MaxSize, conf.Logger.MaxBackups, conf.Logger.MaxAge, IsWindowsService)
		if IsWindowsService {
			servicemanager.RunService(SERVICENAME, false, staservice)
			return
		}
	}

	app := cli.NewApp()
	app.Name = SERVICENAME
	app.Version = VERSION
	app.Usage = SERVICEFULLNAME

	app.Commands = ez_cli.EZCli(SERVICENAME, SERVICEFULLNAME, exePath, confPath, staservice)
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
