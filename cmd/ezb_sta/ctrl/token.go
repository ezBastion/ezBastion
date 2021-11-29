package ctrl

import (
	"encoding/gob"
	"errors"
	"ezBastion/cmd/ezb_sta/models"
	"ezBastion/pkg/confmanager"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"time"
)

func init() {
	gob.Register(models.StaUser{})
}

func Renewtoken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ExePath := c.GetString("exPath")
		conf, err := c.Keys["configuration"].(confmanager.Configuration)
		if err == false {
			c.AbortWithError(http.StatusInternalServerError, errors.New("#STA0001 - Context does not contain configuration struct"))
			return
		}

		cert, tErr := ioutil.ReadFile(ExePath + "/cert/" + conf.EZBSTA.JWT.Issuer + ".key")
		if tErr != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("#STA0002 - Cannot read the STA secret key"))
			return
		}
		ttl := conf.EZBSTA.JWT.TTL * 1000000000
		expirationTime := time.Now().Add(time.Duration(ttl))
		stauser := new(models.StaUser)
		stauser.User = c.MustGet("user").(string)
		// TODO Getting user properties from the cache...

		j := c.MustGet("jwt").(*jwt.Token)
		claims, _ := j.Claims.(jwt.MapClaims)
		tjti := claims["jti"].(string)
		payload := &models.Payload{
			JTI: tjti,
			ISS: conf.EZBSTA.JWT.Issuer,
			SUB: stauser.User,
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
		b.ExpireIn = conf.EZBSTA.JWT.TTL
		c.JSON(http.StatusOK, b)
	}
}

func Createtoken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ExePath := c.GetString("exPath")
		conf, err := c.Keys["configuration"].(confmanager.Configuration)
		if err == false {
			c.AbortWithError(http.StatusInternalServerError, errors.New("#STA0001 - Context does not contain configuration struct"))
			return
		}

		key, tErr := ioutil.ReadFile(ExePath + "/cert/" + conf.EZBSTA.JWT.Issuer + ".key")
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
		// TODO Getting user properties from the cache...

		// ttl in nanoseconds
		ttl := conf.EZBSTA.JWT.TTL * 1000000000
		expirationTime := time.Now().Add(time.Duration(ttl))
		payload := &models.Payload{
			JTI: uuid.NewV4().String(),
			ISS: conf.EZBSTA.JWT.Issuer,
			SUB: stauser.User,
			AUD: conf.EZBSTA.JWT.Audience,
			EXP: expirationTime.Unix(),
		}
		fmt.Println("JTI : " + payload.JTI)
		token := jwt.NewWithClaims(jwt.SigningMethodES256, payload)
		keystruct, _ := jwt.ParseECPrivateKeyFromPEM(key)
		tokenString, tErr := token.SignedString(keystruct)

		// send the token
		b := new(models.Bearer)
		if tErr != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("#STA0003 - Error signing token : "+tErr.Error()))
			return
		}
		b.TokenType = "bearer"
		b.AccessToken = tokenString
		b.ExpireAt = payload.EXP
		b.ExpireIn = conf.EZBSTA.JWT.TTL
		c.JSON(http.StatusOK, b)
	}
}
