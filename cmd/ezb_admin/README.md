#  Admin console (ezb_admin)

The ezBastion console is dedicated to API administrator. With this web console you can 
declare account and api then link them together. A dashboard provide ezBastion statistic. 
ezb_admin is a pure javascript application, running 100% on the administrator browser 
(HTML5 mandatyory).

## Http server

### Corporate server
On any OS, any http server without module. Just prepare a server for static file.
Unzip ezb_admin and add config.json file as:
- ezb_db: The ezb_db address with JWT/token port.
- ezb_srv: The ezb_srv front address and port.
```json
    {
        "ezb_db":"https://ezb_db.fqdn:5012/",
        "ezb_srv":"http://esb_srv.fqdn:5000/"
    }
```
### ezBastion self hosted
Configure Toml file the use ezb_admin with the classic "init / install / start"
```toml
[ezb_srv]
  # ezBastion URL used by API clients. From front of DNS alias, VIP, load balancing... Like https://myserviceapi.corporate.com/
  externalurl = "http://esb_srv.fqdn:5000/"
[ezb_db]
# Use only by ezb_admin with JWT authentication
[ezb_db.listener-external]
# DNS name for this service
fqdn = "ezb_db.fqdn"
# tcp port to listen to.
port = 5012
[ezb_adm]
[ezb_adm.listener]
# DNS name for this service
fqdn = "ezb_db.fqdn"
# tcp port to listen to.
port = 8080

```


## Copyright

Copyright (C) 2021 info@ezbastion.com
<p align="center">
<a href="LICENSE"><img src="https://img.shields.io/badge/license-AGPL%20v3-blueviolet.svg?style=for-the-badge&logo=gnu" alt="License"></a></p>


Used library:

Name       | Copyright | version | url
-----------|-----------|--------:|----------------------------
gin        | MIT       | 1.2     | github.com/gin-gonic/gin
cli        | MIT       | 1.20.0  | github.com/urfave/cli
gorm       | MIT       | 1.9.2   | github.com/jinzhu/gorm
logrus     | MIT       | 1.0.4   | github.com/sirupsen/logrus
go-fqdn    | Apache v2 | 0       | github.com/ShowMax/go-fqdn
jwt-go     | MIT       | 3.2.0   | github.com/dgrijalva/jwt-go
gopsutil   | BSD       | 2.15.01 | github.com/shirou/gopsutil
lumberjack | MIT       | 2.1     | github.com/natefinch/lumberjack
go-sqlite3 | MIT       | 1.10.0  | github.com/mattn/go-sqlite3