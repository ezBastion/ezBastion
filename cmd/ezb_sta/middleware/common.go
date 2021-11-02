package middleware

import (
	"crypto/sha256"
	"crypto/tls"
	db "ezBastion/cmd/ezb_db/models"
	"ezBastion/cmd/ezb_sta/models"
	"ezBastion/pkg/confmanager"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/jtblin/go-ldap-client"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strconv"
)

const (
	ERR_QUERYLDAP = "LDAP:0001"
)

func checkDBUser(c *gin.Context, username string, password string) (errorcode int) {
	config, _ := c.Keys["configuration"].(confmanager.Configuration)
	expath := c.GetString("exPath")
	errorcode = -1

	target := "https://" + config.EZBDB.NetworkPKI.FQDN + ":" + strconv.Itoa(config.EZBDB.NetworkPKI.Port) + "/accounts/" + username
	client := resty.New()
	cert, err := tls.LoadX509KeyPair(path.Join(expath, config.TLS.PublicCert), path.Join(expath, config.TLS.PrivateKey))
	if err != nil {
		errorcode = http.StatusInternalServerError
		return errorcode
	}
	dbaccount := db.EzbAccounts{}
	client.SetRootCertificate(path.Join(expath, config.EZBPKI.CaCert))
	client.SetCertificates(cert)
	resp, err := client.R().
		EnableTrace().
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", c.GetHeader("Authorization")).
		SetResult(&dbaccount).
		Get(target)
	if err != nil {
		errorcode = http.StatusInternalServerError
		return errorcode
	}
	if resp.StatusCode() != 200 {
		errorcode = http.StatusInternalServerError
		return errorcode
	}

	testhash := fmt.Sprintf("%x", sha256.Sum256([]byte(password+dbaccount.Salt)))
	if testhash == dbaccount.Password {
		errorcode = 0
	}
	return errorcode
}

func CheckUserandSet(c *gin.Context, username string, password string, ldapclient *ldap.LDAPClient) (errorcode string) {
	if ldapclient == nil {
		err := checkDBUser(c, username, password)
		if err == 0 {
			// user is computed from specific module
			stauser := models.StaUser{}
			stauser.User = username
			c.Set("connection", stauser)
			c.Set("aud", "internal")
		}
	} else {
		// TODO use also groups and err and not _,_
		ok, _, err := ldapclient.Authenticate(username, password)
		if ok {
			// user is computed from specific module
			stauser := models.StaUser{}
			stauser.User = username
			// TODO compute SID and groups
			c.Set("connection", stauser)
			c.Set("aud", "ad")
		} else {
			if err != nil {
				log.Errorf("Error authenticating user %s: %+v", "username", err)
				errorcode = ERR_QUERYLDAP
			}
		}
	}
	return errorcode
}
