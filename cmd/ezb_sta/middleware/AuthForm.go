package middleware

import (
	"crypto/sha256"
	"crypto/tls"
	"errors"
	db "ezBastion/cmd/ezb_db/models"
	"ezBastion/cmd/ezb_sta/models"
	"ezBastion/pkg/confmanager"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"net/http"
	"path"
	"strconv"
)

func EzbAuthForm(c *gin.Context) {

	var mp models.EzbFormAuth
	err := c.ShouldBindJSON(&mp)
	if err != nil {
		// Error in binding, this is not an basic
		return
	}

	if mp.Grant_type == "password" {

		config, _ := c.Keys["configuration"].(confmanager.Configuration)
		expath := c.GetString("exPath")
		username := mp.Username
		password := mp.Password

		target := "https://" + config.EZBDB.NetworkPKI.FQDN + ":" + strconv.Itoa(config.EZBDB.NetworkPKI.Port) + "/accounts/" + username
		client := resty.New()
		cert, err := tls.LoadX509KeyPair(path.Join(expath, config.TLS.PublicCert), path.Join(expath, config.TLS.PrivateKey))
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		dbaccount := db.EzbAccounts{}
		client.SetRootCertificate(path.Join(expath, config.EZBPKI.CaCert))
		client.SetCertificates(cert)
		resp, e := client.R().
			EnableTrace().
			SetHeader("Accept", "application/json").
			SetHeader("Authorization", c.GetHeader("Authorization")).
			SetResult(&dbaccount).
			Get(target)
		if e != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if resp.StatusCode() != 200 {
			c.AbortWithError(http.StatusInternalServerError, errors.New("#A0001 EZB_DB return error : "+err.Error()))
			return
		}

		testhash := fmt.Sprintf("%x", sha256.Sum256([]byte(password+dbaccount.Salt)))
		if testhash == dbaccount.Password {
			// user is computed from specific module
			stauser := models.StaUser{}
			stauser.User = dbaccount.Name
			// TODO compute SID and groups
			c.Set("connection", stauser)
			c.Set("aud", "internal")
		}
	}
}
