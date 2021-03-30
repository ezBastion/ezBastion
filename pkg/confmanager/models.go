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
	Logger Logger `toml:"logger"`
	TLS    TLS    `toml:"tls"`
	EZBSRV EZBSRV `toml:"ezb_srv"`
	EZBDB  EZBDB  `toml:"ezb_db"`
	EZBPKI EZBPKI `toml:"ezb_pki"`
	EZBWKS EZBWKS `toml:"ezb_wks"`
	EZBSTA EZBSTA `toml:"ezb_sta"`
	EZBADM EZBADM `toml:"ezb_adm"`
}

type Logger struct {
	LogLevel   string `toml:"loglevel" comment:"log output filter [debug | info | warning | error | critical]"`
	MaxSize    int    `toml:"maxsize" comment:"MaxSize is the maximum size in megabytes of the log file before it gets\n rotated. It defaults to 100 megabytes."`
	MaxBackups int    `toml:"maxbackups" comment:"Whenever a new logfile gets created, old log files may be deleted.\n The most recent files according to the encoded timestamp will be retained, up to a number equal\n to MaxBackups (or all of them if MaxBackups is 0)."`
	MaxAge     int    `toml:"maxage" comment:"MaxAge is the maximum number of days to retain old log files based on the\n timestamp encoded in their filename.  Note that a day is defined as 24 hours and may not\n exactly correspond to calendar days due to daylight savings, leap seconds, etc.\n The default is not to remove old log files based on age."`
}

type Network struct {
	Port int    `toml:"port" comment:"tcp port to listen to."`
	FQDN string `toml:"fqdn" comment:"DNS name for this service"`
}

type TLS struct {
	SAN        []string `toml:"san" comment:"SAN (Subject Alternative Name) is a list of name that allows identities to be bound\n to the subject of the certificate. It can be a DNS name, an IP address or a NBIOS name."`
	PrivateKey string   `toml:"privatekey" comment:"Private ECDA key, used to communicate with other ezBastion microservice and sign JWT tokens. Generate but not used by ezb_pki.\nRelative path from ezBastion binaries."`
	PublicCert string   `toml:"publiccert" comment:"Public  ECDA certificate, used to communicate with other ezBastion microservice. Generate but not used by ezb_pki.\nRelative path from ezBastion binaries."`
}

type EZBADM struct {
	Network Network `toml:"listener"`

}

type EZBSRV struct {
	Network Network `toml:"listener"`
	ExternalURL string `toml:"externalurl" comment:"ezBastion URL used by API clients. From front of DNS alias, VIP, load balancing... Like https://myserviceapi.corporate.com/"`
	CacheL1 int     `toml:"cacheL1" comment:"Cache memory duration in second. RAM cache for high performance, used to unload database.\n Longer value for less DB request but increase waiting time to apply modification coming from admin console."`
}
type EZBSTA struct {
	Network Network `toml:"listener"`
	JWT     JWT     `toml:"jwt"`
}
type JWT struct {
	Issuer   string `toml:"issuer" comment:" sta (Secure Token Authority) name, must be unique and set in ezb_admin"`
	Audience string `toml:"audience" comment:"by default [ezBastion]"`
	TTL      int    `toml:"ttl" comment:"Time to live of jwt token in second"`
}
type EZBDB struct {
	NetworkPKI Network `toml:"listener-internal" comment:"Use only by ezb_srv with PKI authentication"`
	NetworkJWT Network `toml:"listener-external" comment:"Use only by ezb_admin with JWT authentication"`
	DB         string  `toml:"db" comment:"only [sqlite] available for the moment."`
	SQLITE     SQLite  `toml:"sqlite"`
	//MSSQL   MSSql   `toml:"mssql"`
	//MYSQL   MYSql   `toml:"mysql"`
}

type EZBPKI struct {
	Network Network `toml:"listener"`
	CaCert  string  `toml:"cacert" comment:"Root PKI public cert, must be copied on all ezBastion nodes"`
	CaKey   string  `toml:"cakey" comment:"Root PKI private key, do not share it. Must ne ONLY on PKI node"`
}

type EZBWKS struct {
	Network      Network `toml:"listener"`
	ScriptPath   string  `toml:"scriptpath" comment:"Full path of root folder where jobs are will be created."`
	JobPath      string  `toml:"jobpath" comment:"Full path of root folder where add your scripts."`
	LimitWarning int     `toml:"limitwarning" comment:"Goroutine number before add warning in log file, 0 for ni limit"`
	LimitMax     int     `toml:"limitmax" comment:"Goroutine number before reject api call with 429 <to many request>, 0 for ni limit"`
	ServiceName  string  `toml:"name" comment:"Service name"`
}

type MSSql struct {
	Host     string `toml:"host"`
	Database string `toml:"database"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Instance string `toml:"instance"`
}

type SQLite struct {
	DBPath string `toml:"dbpath" comment:"Relative path to ezb_db.exe file"`
}

type MYSql struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}
