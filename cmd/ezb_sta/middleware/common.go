package middleware

import (
	"crypto/sha256"
	"crypto/tls"
	db "ezBastion/cmd/ezb_db/models"
	"ezBastion/cmd/ezb_sta/models"
	"ezBastion/pkg/confmanager"
	"ezBastion/pkg/setupmanager"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"gopkg.in/ldap.v2"
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

func F_GetADproperties(username string, lc *models.Ldapinfo) (iu *models.IntrospectUser, err error) {

	searchRequest := ldap.NewSearchRequest(
		lc.Base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(lc.UserFilter, username),
		[]string{"ou", "ntaccount", "samaccountname", "description", "displayname", "emailaddress", "givenname", "distinguishedName"},
		nil,
	)
	sr, err := lc.LConn.Search(searchRequest)
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
		//TODO groups
		//iu.Groups, _ = lc.GetGroupsOfUser(firstentry.DN)
	}

	return iu, nil
}
func LDAPconnect(ldapclient *models.Ldapinfo) (*ldap.Conn, error) {
	// Proceed to a test...
	ldapurl := fmt.Sprintf("%s:389", ldapclient.ServerName)
	l, err := ldap.Dial("tcp", ldapurl)
	if err != nil {
		fmt.Errorf("Failed to connect. %s", err)
		return nil, err
	}

	if err := l.Bind(ldapclient.BindDN, ldapclient.BindPassword); err != nil {
		fmt.Errorf("Failed to bind. %s", err)
		return nil, err
	}

	return l, nil
}

func LDAPTLSconnect(ldapclient *models.Ldapinfo, certfile string, pkfile string, srvname string) (*ldap.Conn, error) {
	// Proceed to a test...
	exePath, err := setupmanager.ExePath()
	ldapurl := fmt.Sprintf("%s:636", ldapclient.ServerName)
	//tlsconf := &tls.Config{InsecureSkipVerify: true}

	// Load cer & key files into a pair of []byte
	cert, err := tls.LoadX509KeyPair(path.Join(exePath, certfile), path.Join(exePath, pkfile))
	tlsconf := &tls.Config{ServerName: srvname, Certificates: []tls.Certificate{cert}}
	l, err := ldap.DialTLS("tcp", ldapurl, tlsconf)
	if err != nil {
		fmt.Errorf("Failed to connect. %s", err)
		return nil, err
	}

	if err := l.Bind(ldapclient.BindDN, ldapclient.BindPassword); err != nil {
		fmt.Errorf("Failed to bind. %s", err)
		return nil, err
	}

	return l, nil
}

func LDAPauth(ldapclient *models.Ldapinfo, user string, pass string) (bool, []string, error) {
	result, err := ldapclient.LConn.Search(ldap.NewSearchRequest(
		ldapclient.Base,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf(ldapclient.UserFilter, user),
		[]string{"dn"},
		nil,
	))

	if err != nil {
		return false, nil, fmt.Errorf("Failed to find user. %s", err)
	}

	if len(result.Entries) < 1 {
		return false, nil, fmt.Errorf("User does not exist")
	}

	if len(result.Entries) > 1 {
		return false, nil, fmt.Errorf("Too many entries returned")
	}

	if err := ldapclient.LConn.Bind(result.Entries[0].DN, pass); err != nil {
		fmt.Printf("Failed to auth. %s", err)
	} else {
		fmt.Printf("Authenticated successfuly!")
	}

	return true, nil, nil
}
