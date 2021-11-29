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
	genldap "gopkg.in/ldap.v2"
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

func F_GetADproperties(username string, lc *ldap.LDAPClient) (iu *models.IntrospectUser, err error) {

	searchRequest := genldap.NewSearchRequest(
		lc.Base,
		genldap.ScopeWholeSubtree, genldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(lc.UserFilter, username),
		[]string{"ou", "ntaccount", "samaccountname", "description", "displayname", "emailaddress", "givenname", "distinguishedName"},
		nil,
	)
	sr, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(sr.Entries) > 0 {
		iu = new(models.IntrospectUser)
		// We will take obnly the first entry (the first user matching username given
		firstentry := sr.Entries[0]
		iu.Distinguishedname = firstentry.DN
		iu.Displayname = firstentry.GetAttributeValue("displayname")
		iu.Description = firstentry.GetAttributeValue("description")
		iu.Emailaddress = firstentry.GetAttributeValue("emailaddress")
		iu.Givenname = firstentry.GetAttributeValue("givenname")
		iu.Ntaccount = firstentry.GetAttributeValue("ntaccount")
		iu.Ou = firstentry.GetAttributeValue("ou")
		iu.Samaccountname = firstentry.GetAttributeValue("samaccountname")
		iu.Groups, _ = lc.GetGroupsOfUser(firstentry.DN)
	}

	return iu, nil
}
