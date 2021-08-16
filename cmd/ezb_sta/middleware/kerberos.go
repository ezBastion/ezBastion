package middleware

import (
	"crypto/tls"
	"errors"
	db "ezBastion/cmd/ezb_db/models"
	"ezBastion/pkg/confmanager"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/quasoft/websspi"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strconv"
	"strings"
)

var (
	auth   *websspi.Authenticator
	config *websspi.Config
)

func EzbAuthKerberos(c *gin.Context) {

	logg := log.WithFields(log.Fields{"Middleware": "kerberos"})
	authHead := c.GetHeader("Authorization")
	if authHead != "" {

		nego := strings.Split(authHead, " ")
		if len(nego) != 2 {
			logg.Error("bad Authorization #J0001: " + authHead)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-KRB0001"))
			return
		}
		if strings.Compare(strings.ToLower(nego[0]), "negotiate") != 0 {
			logg.Error("bad Authorization #J0002: " + authHead)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-KRB0002"))
			return
		}

		config, _ := c.Keys["configuration"].(confmanager.Configuration)
		expath := c.GetString("exPath")
		username := ""
		// no need to give the password

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

		/*testhash := fmt.Sprintf("%x", sha256.Sum256([]byte(password+dbaccount.Salt)))
		if testhash == dbaccount.Password {
			// user is computed from specific module
			stauser := models.StaUser{}
			stauser.User = dbaccount.Name
			// TODO compute SID and groups
			c.Set("connection", stauser)
			c.Set("aud", "internal")
		}
		*/

	}
}

func SspiHandler() gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	config = websspi.NewConfig()
	auth, _ = websspi.New(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		user := ctx.Value("UserInfo")
		fmt.Sprint(user)
	})
	// try to use the handler to do the sspi
	h := auth.WithAuth(handler)

	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(c)
		h.ServeHTTP(c.Writer, c.Request)
	}
}
