# ezBastion microservices

- All microservices are store in **cmd** directory. 
- All Library (aka ezb_lib) are store in **pkg** directory.
- To build **ezb_db** you need **GCC** (for sqlIt driver). 
  For windows [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) do it nicely.



### Build.

- update version & build all => powershell Makefile.ps1 build
- update version & build one => powershell Makefile.ps1 build ezb_srv
- update all binary version  => powershell Makefile.ps1 generate
- update one binary version  => powershell Makefile.ps1 generate ezb_srv 
- make zip                   => powershell Makefile.ps1 compress

```powershell
    PS E:\ezbastion\ezb_wks> ezb_wks install
    PS E:\ezbastion\ezb_wks> ezb_wks start
```




## Copyright

Copyleft (C) 2021 ezBastion info@ezbastion.com
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
