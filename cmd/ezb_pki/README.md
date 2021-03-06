# ezBastion internal PKI microservice.

**ezb_pki** is a *Public Key infrastructure* microservice. It will used by ezBastion nodes to interact together.


## SETUP

The PKI (Public Key Infrastructure) is the first node to be installed. It will be in charge to create and deploy the ECDSA pair key, used by all ezBastion's node to communicate between them.
The certificates are used to sign JWT too.


### 1. Download ezb_pki.zip from [GitHub](<https://github.com/ezBastion/ezBastion/releases/latest>)

### 2. Unzip it, then with an admin command like CMD or Powershell.

### 3. Run ezb_pki.exe with **init** option.

```powershell
    PS E:\ezbastion> ezb_pki.exe init
```

```toml
[ezb_pki]

# Root PKI public cert, must be copied on all ezBastion nodes
cacert = "cert/ca.crt"

# Root PKI private key, do not share it. Must ne ONLY on PKI node
cakey = "cert/ca.key"

[ezb_pki.listener]
fqdn = "chavers-desk"
port = 5010
[logger]

# log output filter [debug | info | warning | error | critical]
loglevel = "debug"
maxage = 180
maxbackups = 10
maxsize = 5

[tls]
san = ["chavers-desk"]
```

- **servicename**: This is the name used as Windows service and as certificates root name.
- **servicefullname**: The Windows service description.
- **listen**: The TCP/IP port used by ezb_pki to respond at nodes request. This port MUST BE reachable by all ezBastion's node.
- **loglevel**: Choose log level in debug,info,warning,error,critical.
- **maxsize**: is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes.
- **maxbackups**: MaxBackups is the maximum number of old log files to retain.
- **maxage**: MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.


### 4. Install Windows service and start it.

```powershell
    ezb_pki install
    ezb_pki start
```


## security consideration

- ezb_pki is an auto-enrolment system, if you do not add nodes, stop the service or don't install it and use debug mode instead.
- Protect cert folder.
- Backup the private/public key.


## Copyright

Copyleft (C) 2018 Renaud DEVERS info@ezbastion.com
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