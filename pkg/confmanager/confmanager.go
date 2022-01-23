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

package confmanager

import (
	"fmt"
	"github.com/Showmax/go-fqdn"
	toml "github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"path"
)

func CheckConfig(confPath string, exePath string) (conf Configuration, err error) {
	raw, readerror := ioutil.ReadFile(confPath)
	if readerror != nil {
		currentfqdn, err := fqdn.FqdnHostname()
		if err != nil {
			currentfqdn, _ = os.Hostname()
		}
		conf.EZBPKI.Network.FQDN = currentfqdn
		conf.EZBPKI.Network.Port = 5010
		conf.TLS.SAN = []string{currentfqdn}
		conf.EZBWKS.ServiceName = "ezb_wks"
		conf.EZBWKS.ScriptInterpreter = "powershell.exe"
		conf.EZBWKS.InterpreterParams = []string{"-NoLogo", "-NonInteractive", "-File"}
		conf.EZBPKI.CaCert = "cert/ca.crt"
		conf.EZBPKI.CaKey = "cert/ca.key"
		conf.EZBPKI.Autosign = 1
		conf.TLS.PrivateKey = "cert/ezbastion.key"
		conf.TLS.PublicCert = "cert/ezbastion.crt"
		conf.Logger.LogLevel = "info"
		conf.Logger.MaxAge = 180
		conf.Logger.MaxBackups = 10
		conf.Logger.MaxSize = 5
		conf.EZBDB.DB = "sqlite"
		conf.EZBDB.SQLITE.DBPath = "db/ezbastion.db"
		conf.EZBDB.NetworkPKI.FQDN = currentfqdn
		conf.EZBDB.NetworkPKI.Port = 5011
		conf.EZBDB.NetworkJWT.FQDN = currentfqdn
		conf.EZBDB.NetworkJWT.Port = 5012
		conf.EZBSRV.Network.FQDN = currentfqdn
		conf.EZBSRV.Network.Port = 5000
		conf.EZBSRV.CacheL1 = 600
		conf.EZBSRV.ExternalURL = fmt.Sprintf("http://%s:5000/", currentfqdn)
		conf.EZBSRV.LB = "rand"
		conf.EZBWKS.Network.FQDN = currentfqdn
		conf.EZBWKS.Network.Port = 5100
		conf.EZBSTA.Network.FQDN = currentfqdn
		conf.EZBSTA.Network.Port = 1443
		conf.EZBSTA.JWT.TTL = 1200
		conf.EZBSTA.JWT.Audience = "ezBastion"
		conf.EZBSTA.JWT.Issuer = "ezbastion"
		conf.EZBSTA.StaLdap.Base = ""
		conf.EZBSTA.StaLdap.Port = 389
		conf.EZBSTA.StaLdap.UseSSL = false
		conf.EZBSTA.StaLdap.SkipTLS = false
		conf.EZBSTA.StaLdap.BindDN = ""
		conf.EZBSTA.StaLdap.BindUser = ""
		conf.EZBSTA.StaLdap.BindPassword = ""
		conf.EZBSTA.StaLdap.ServerName = ""
		conf.EZBSTA.StaLdap.LDAPcrt = "cert/ldap.cer"
		conf.EZBSTA.StaLdap.LDAPpk = "cert/ldap.key"
		conf.EZBSTA.StaLdap.Userfilter = ""
		conf.EZBSTA.StaLdap.Groupfilter = ""
		conf.EZBWKS.ScriptPath = path.Join(exePath, "script")
		conf.EZBWKS.JobPath = path.Join(exePath, "job")
		conf.EZBWKS.LimitMax = 0
		conf.EZBWKS.LimitWarning = 0
		conf.EZBADM.Network.FQDN = currentfqdn
		conf.EZBADM.Network.Port = 8080
		return conf, readerror
	}
	toml.Unmarshal(raw, &conf)
	return conf, nil
}
