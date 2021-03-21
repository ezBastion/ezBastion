#  ezBastion API frontend (ezb_srv)

This node, called Bastion, is the front part of ezBastion architecture. It receive all clients requests and serves as api gateway.


## SETUP

*prerequisite: ezb_pki*

### 1. Download ezb_srv.zip from [GitHub](<https://github.com/ezBastion/ezBastion/releases/latest>)

### 2. Unzip it, then with an admin command like CMD or Powershell.

### 3. Run ezb_srv.exe with **init** option.

```powershell
    PS E:\ezbastion> ezb_srv.exe init
```

this command will create folders and default config Toml file. If config file exist, the *init* process pass to next step (certificate generation). 


```toml

[ezb_db]

# only [sqlite] available
db = "sqlite"

[ezb_db.listener]

# DNS name for this service
fqdn = "ezbdb1.ezbastion.com"

# tcp port to listen to.
port = 5011

[ezb_db.sqlite]

# Relative path to ezb_db.exe file
dbpath = "db/ezbastion.db"

[ezb_pki]

# Root PKI public cert, must be copied on all ezBastion nodes
cacert = "./cert/ca.crt"

# Root PKI private key, do not share it. Must ne ONLY on PKI node
cakey = "cert/ca.key"

[ezb_pki.listener]

# DNS name for this service
fqdn = "ezbpki.ezbastion.com"

# tcp port to listen to.
port = 5010

[ezb_srv]

# Cache memory duration in second. RAM cache for high performance, used to unload database.
# Longer value for less DB request but increase waiting time to apply modification coming from admin console.
cacheL1 = 600

[ezb_srv.listener]

# DNS name for this service
fqdn = "api.ezbastion.com"

# tcp port to listen to.
port = 5000

[ezb_wks]
jobpath = ""
limitmax = 0
limitwarning = 0
scriptpath = ""

[ezb_wks.listener]

# DNS name for this service
fqdn = "worker1.ezbastion.com"

# tcp port to listen to.
port = 5100

[logger]

# log output filter [debug | info | warning | error | critical]
loglevel = "debug"

# MaxAge is the maximum number of days to retain old log files based on the
# timestamp encoded in their filename.  Note that a day is defined as 24 hours and may not
# exactly correspond to calendar days due to daylight savings, leap seconds, etc.
# The default is not to remove old log files based on age.
maxage = 180

# Whenever a new logfile gets created, old log files may be deleted.
# The most recent files according to the encoded timestamp will be retained, up to a number equal
# to MaxBackups (or all of them if MaxBackups is 0).
maxbackups = 10

# MaxSize is the maximum size in megabytes of the log file before it gets
# rotated. It defaults to 100 megabytes.
maxsize = 5

[tls]

# Private ECDA key, used to communicate with other ezBastion microservice and sign JWT tokens. Generate but not used by ezb_pki.
privatekey = "cert/ezbastion.key"

# Public  ECDA certificate, used to communicate with other ezBastion microservice. Generate but not used by ezb_pki.
publiccert = "cert/ezbastion.crt"

# SAN (Subject Alternative Name) is a list of name that allows identities to be bound
# to the subject of the certificate. It can be a DNS name, an IP address or a NBIOS name.
san = ["worker1.ezbastion.com","api.ezbastion.com","ezbpki.ezbastion.com","ezbdb1.ezbastion.com"]

```
Edit config file then generate certificate, by restart init process.

```powershell
    PS E:\ezbastion> ezb_srv.exe init
```

### 4. Install Windows service and start it.

```powershell
    PS E:\ezbastion\ezb_srv> ezb_srv.exe install
    PS E:\ezbastion\ezb_srv> ezb_srv.exe start
```

Or run it interactively with *debug* option without installation.
```powershell
    PS E:\ezbastion\ezb_srv> ezb_srv.exe debug
```


## Copyright

Copyleft (C) 2021 info@ezbastion.com
<p align="center">
<a href="LICENSE"><img src="https://img.shields.io/badge/license-AGPL%20v3-blueviolet.svg?style=for-the-badge&logo=gnu" alt="License"></a></p>


Used library:

Name      | Copyright | version | url
----------|-----------|--------:|----------------------------
gin       | MIT       | 1.2     | github.com/gin-gonic/gin
cli       | MIT       | 1.20.0  | github.com/urfave/cli
gorm      | MIT       | 1.9.2   | github.com/jinzhu/gorm
logrus    | MIT       | 1.0.4   | github.com/sirupsen/logrus
go-fqdn   | Apache v2 | 0       | github.com/ShowMax/go-fqdn
jwt-go    | MIT       | 3.2.0   | github.com/dgrijalva/jwt-go
gopsutil  | BSD       | 2.15.01 | github.com/shirou/gopsutil
lumberjack| MIT       | 2.1     | github.com/natefinch/lumberjack
go-toml   | MIT       | 1.8.1   | github.com/pelletier/go-toml