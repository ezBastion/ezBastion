package ctrl

import (
	"bytes"
	"encoding/gob"
	"errors"
	"ezBastion/cmd/ezb_srv/cache"
	"ezBastion/cmd/ezb_sta/models"
	"ezBastion/pkg/confmanager"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"time"
)

func Renewtoken(storage cache.Storage) gin.HandlerFunc {
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
		stauser, _ := c.MustGet("user").(string)
		j := c.MustGet("jwt").(*jwt.Token)
		claims, _ := j.Claims.(jwt.MapClaims)
		tjti := claims["jti"].(string)
		payload := &models.Payload{
			JTI: tjti,
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
		// Token is created, let's cache it
		StoreToken(storage, b, tjti, b.ExpireIn)
		// then send it
		c.JSON(http.StatusOK, b)
	}
}

func Createtoken(storage cache.Storage) gin.HandlerFunc {
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
		// ttl in nanoseconds
		ttl := conf.EZBSTA.JWT.TTL * 1000000000
		expirationTime := time.Now().Add(time.Duration(ttl))
		tjti := uuid.NewV4().String()
		payload := &models.Payload{
			JTI: tjti,
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
		// Token is created, let's cache it
		StoreToken(storage, b, tjti, b.ExpireIn)
		// then send it
		c.JSON(http.StatusOK, b)
	}
}

func StoreToken(storage cache.Storage, b *models.Bearer, key string, ttk int) {

	var bearer bytes.Buffer
	enc := gob.NewEncoder(&bearer)
	err := enc.Encode(b)
	if err != nil {
		// Error on encoding
	}
	storage.Set(key, bearer.Bytes(), time.Duration(ttk)*time.Second)
}
