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

type Configuration struct {
	Logger Logger `toml:"logger" json:"logger" ini:"logger"`
	TLS    TLS    `toml:"tls" json:"tls" ini:"tls"`
	EZBSRV EZBSRV `toml:"ezb_srv" json:"ezb_srv" ini:"ezb_srv"`
	EZBDB  EZBDB  `toml:"ezb_db" json:"ezb_db" ini:"ezb_db"`
	EZBPKI EZBPKI `toml:"ezb_pki" json:"ezb_pki" ini:"ezb_pki"`
	EZBWKS EZBWKS `toml:"ezb_wks" json:"ezb_wks" ini:"ezb_wks"`
	EZBSTA EZBSTA `toml:"ezb_sta" json:"ezb_sta" ini:"ezb_sta"`
	EZBADM EZBADM `toml:"ezb_adm" json:"ezb_adm" ini:"ezb_adm"`
}

type Logger struct {
	LogLevel   string `ini:"loglevel" json:"loglevel" toml:"loglevel" comment:"log output filter [debug | info | warning | error | critical]"`
	MaxSize    int    `ini:"maxsize" json:"maxsize" toml:"maxsize" comment:"MaxSize is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes."`
	MaxBackups int    `ini:"maxbackups" json:"maxbackups" toml:"maxbackups" comment:"Whenever a new logfile gets created, old log files may be deleted. The most recent files according to the encoded timestamp will be retained, up to a number equal to MaxBackups (or all of them if MaxBackups is 0)."`
	MaxAge     int    `ini:"maxage" json:"maxage" toml:"maxage" comment:"MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.  Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc. The default is not to remove old log files based on age."`
}

type Network struct {
	Port int    `ini:"port" json:"port" toml:"port" comment:"tcp port to listen to."`
	FQDN string `ini:"fqdn" json:"fqdn" toml:"fqdn" comment:"DNS name for this service"`
}

type TLS struct {
	SAN        []string `ini:"san" json:"san" toml:"san" comment:"SAN (Subject Alternative Name) is a list of name that allows identities to be bound to the subject of the certificate. It can be a DNS name, an IP address or a NBIOS name."`
	PrivateKey string   `ini:"privatekey" json:"privatekey" toml:"privatekey" comment:"Private ECDA key, used to communicate with other ezBastion microservice and sign JWT tokens. Generate but not used by ezb_pki.Relative path from ezBastion binaries."`
	PublicCert string   `ini:"publiccert" json:"publiccert" toml:"publiccert" comment:"Public  ECDA certificate, used to communicate with other ezBastion microservice. Generate but not used by ezb_pki.Relative path from ezBastion binaries."`
}

type EZBADM struct {
	Network Network `ini:"ezb_adm.listener" json:"listener" toml:"listener"`
}

type EZBSRV struct {
	Network     Network `ini:"ezb_srv.listener" json:"listener" toml:"listener"`
	ExternalURL string  `ini:"externalurl" json:"externalurl" toml:"externalurl" comment:"ezBastion URL used by API clients. From front of DNS alias, VIP, load balancing... Like https://myserviceapi.corporate.com/"`
	CacheL1     int     `ini:"cacheL1" json:"cacheL1" toml:"cacheL1" comment:"Cache memory duration in second. RAM cache for high performance, used to unload database. Longer value for less DB request but increase waiting time to apply modification coming from admin console."`
	LB          string  `ini:"loadbalancing" json:"loadbalancing" toml:"loadbalancing" comment:"[rand|rrb]Workers load balancing algorithms (random or round robin)"`
}
type EZBSTA struct {
	Network Network `ini:"ezb_sta.listener" json:"listener" toml:"listener"`
	JWT     JWT     `ini:"ezb_sta.jwt" json:"jwt" toml:"jwt"`
}
type JWT struct {
	Issuer   string `ini:"issuer" json:"issuer" toml:"issuer" comment:" sta (Secure Token Authority) name, must be unique and set in ezb_admin"`
	Audience string `ini:"audience" json:"audience" toml:"audience" comment:"by default [ezBastion]"`
	TTL      int    `ini:"ttl" json:"ttl" toml:"ttl" comment:"Time to live of jwt token in second"`
}
type EZBDB struct {
	NetworkPKI Network `ini:"ezb_db.listener-internal" json:"listener-internal" toml:"listener-internal" comment:"Use only by ezb_srv with PKI authentication"`
	NetworkJWT Network `ini:"ezb_db.listener-external" json:"listener-external" toml:"listener-external" comment:"Use only by ezb_admin with JWT authentication"`
	DB         string  `ini:"db" json:"db" toml:"db" comment:"only [sqlite] available for the moment."`
	SQLITE     SQLite  `ini:"ezb_db.sqlite" json:"sqlite" toml:"sqlite"`
	//MSSQL   MSSql   `toml:"mssql"`
	//MYSQL   MYSql   `toml:"mysql"`
}

type EZBPKI struct {
	Network  Network `ini:"ezb_pki.listener" json:"listener" toml:"listener"`
	CaCert   string  `ini:"cacert" json:"cacert" toml:"cacert" comment:"Root PKI public cert, must be copied on all ezBastion nodes"`
	CaKey    string  `ini:"cakey" json:"cakey" toml:"cakey" comment:"Root PKI private key, do not share it. Must ne ONLY on PKI node"`
	Autosign int     `ini:"autosign" json:"autosign" toml:"autosign" comment:"[0|1] 0: Store CSR until manually signature; 1: Accept and sign all CSRs"`
}

type EZBWKS struct {
	Network           Network  `ini:"ezb_wks.listener" json:"listener" toml:"listener"`
	ScriptPath        string   `ini:"scriptpath" json:"scriptpath" toml:"scriptpath" comment:"Full path of root folder where jobs are will be created."`
	JobPath           string   `ini:"jobpath" json:"jobpath" toml:"jobpath" comment:"Full path of root folder where add your scripts."`
	LimitWarning      int      `ini:"limitwarning" json:"limitwarning" toml:"limitwarning" comment:"Goroutine number before add warning in log file, 0 for ni limit"`
	LimitMax          int      `ini:"limitmax" json:"limitmax" toml:"limitmax" comment:"Goroutine number before reject api call with 429 <to many request>, 0 for ni limit"`
	ServiceName       string   `ini:"name" json:"name" toml:"name" comment:"Service name"`
	ScriptInterpreter string   `ini:"interpreter" json:"interpreter" toml:"interpreter" comment:"Binary used to execute the script, like Python, Perl, Ruby, Powershell.exe, pwsh ... Can be the full path or look in the path."`
	InterpreterParams []string `ini:"interpreterparams" json:"interpreterparams" toml:"interpreterparams" comment:"Static interpreter parameters need between binary and script file. Powershell EX: ['-NoLogo', '-NonInteractive', '-File']"`
}

type MSSql struct {
	Host     string `toml:"host"`
	Database string `toml:"database"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Instance string `toml:"instance"`
}

type SQLite struct {
	DBPath string `ini:"dbpath" json:"dbpath" toml:"dbpath" comment:"Relative path to ezb_db.exe file"`
}

type MYSql struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}
