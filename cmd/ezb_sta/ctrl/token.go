package ctrl

import (
	"errors"
	"ezBastion/cmd/ezb_sta/models"
	"ezBastion/pkg/confmanager"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ExePath string
	Conf    confmanager.Configuration
)

func Renewtoken(c *gin.Context) {
	expath := c.GetString("exPath")
	conf, err := c.Keys["configuration"].(confmanager.Configuration)
	if err == false {
		c.AbortWithError(http.StatusInternalServerError, errors.New("#STA0001 - Context does not contain configuration struct"))
		return
	}

	cert, tErr := ioutil.ReadFile(expath + "/cert/" + conf.EZBSTA.JWT.Issuer + ".key")
	if tErr != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("#STA0002 - Cannot read the STA secret key"))
		return
	}
	expirationTime := time.Now().Add(time.Minute)
	stauser, _ := c.MustGet("user").(string)
	j := c.MustGet("jwt").(*jwt.Token)
	claims, _ := j.Claims.(jwt.MapClaims)
	payload := &models.Payload{
		JTI: claims["jti"].(string),
		ISS: conf.EZBSTA.JWT.Issuer,
		SUB: stauser,
		AUD: conf.EZBSTA.JWT.Audience,
		EXP: expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, payload)
	keystruct, _ := jwt.ParseECPrivateKeyFromPEM(cert)
	tokenString, tErr := token.SignedString(keystruct)
	if tErr != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("#STA0003 - Error signing token : "+tErr.Error()))
		return
	}
	b := new(models.Bearer)
	b.TokenType = "bearer"
	b.AccessToken = tokenString
	b.ExpireAt = payload.EXP
	b.ExpireIn = expirationTime.Second()
	c.JSON(http.StatusOK, b)
}

func Createtoken(c *gin.Context) {

	expath := c.GetString("exPath")
	conf := c.Keys["configuration"].(confmanager.Configuration)

	key, tErr := ioutil.ReadFile(expath + "/cert/" + conf.EZBSTA.JWT.Issuer + ".key")
	if tErr != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("#STA0002 - Cannot read the STA secret key"))
		return
	}
	conn, err := c.Get("connection")
	if err == false {
		c.JSON(http.StatusUnauthorized, "")
		return
	}
	stauser := conn.(models.StaUser)
	expirationTime := time.Now().Add(time.Minute)
	payload := &models.Payload{
		JTI: uuid.NewV4().String(),
		ISS: conf.EZBSTA.JWT.Issuer,
		SUB: stauser.User,
		AUD: conf.EZBSTA.JWT.Audience,
		EXP: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, payload)
	keystruct, _ := jwt.ParseECPrivateKeyFromPEM(key)
	tokenString, tErr := token.SignedString(keystruct)
	b := new(models.Bearer)
	if tErr != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("#STA0003 - Error signing token : "+tErr.Error()))
		return
	}
	b.TokenType = "bearer"
	b.AccessToken = tokenString
	b.ExpireAt = payload.EXP
	b.ExpireIn = expirationTime.Second()
	c.JSON(http.StatusOK, b)
}
