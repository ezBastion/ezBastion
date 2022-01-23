package middleware

import (
	"crypto/sha256"
	"crypto/tls"
	db "ezBastion/cmd/ezb_db/models"
	"ezBastion/cmd/ezb_sta/models"
	"ezBastion/pkg/confmanager"
	"ezBastion/pkg/logmanager"
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
	// This is the main entry to handle LDAP connection. First check the configuration to see the settings
	if ldapclient.SkipTLS {
		if ldapclient.UseSSL {
			// If port is set to 389, switch to 636, otherwise use the port defined
			if ldapclient.Port == 389 {
				_ = logmanager.Debug("LDAP client port set to 389 but also using SSL, so switched to 636")
				ldapclient.Port = 636
			}
			return ldapSTDconnect(ldapclient)
		} else {
			// If port set to a value different from 389, switched to 389, standard non SSL port
			if ldapclient.Port != 389 {
				_ = logmanager.Debug(fmt.Sprintf("LDAP client port set to %d but not using SSLL, so switched to 389", ldapclient.Port))
				ldapclient.Port = 389
			}
			return ldapSTDconnect(ldapclient)
		}
	} else {
		// Will use the TLS to connect the LDAP, bypass use SSL and port check
		return ldapTLSconnect(ldapclient)
	}
}

func ldapSTDconnect(ldapclient *models.Ldapinfo) (*ldap.Conn, error) {
	// Proceed to a test...
	ldapurl := fmt.Sprintf("%s:%d", ldapclient.ServerName, ldapclient.Port)
	l, err := ldap.Dial("tcp", ldapurl)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect. %s", err)
	}

	if err := l.Bind(ldapclient.BindDN, ldapclient.BindPassword); err != nil {
		return nil, fmt.Errorf("Failed to bind. %s", err)
	}

	return l, nil
}

func ldapTLSconnect(ldapclient *models.Ldapinfo) (*ldap.Conn, error) {
	// Proceed to a test...
	exePath, err := setupmanager.ExePath()
	ldapurl := fmt.Sprintf("%s:%d", ldapclient.ServerName, ldapclient.Port)
	//tlsconf := &tls.Config{InsecureSkipVerify: true}

	// Load cer & key files into a pair of []byte
	cert, err := tls.LoadX509KeyPair(path.Join(exePath, ldapclient.LDAPcrt), path.Join(exePath, ldapclient.LDAPpk))
	tlsconf := &tls.Config{ServerName: ldapclient.ServerName, Certificates: []tls.Certificate{cert}}
	l, err := ldap.DialTLS("tcp", ldapurl, tlsconf)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect. %s", err)
	}

	if err := l.Bind(ldapclient.BindDN, ldapclient.BindPassword); err != nil {
		return nil, fmt.Errorf("Failed to bind. %s", err)
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
