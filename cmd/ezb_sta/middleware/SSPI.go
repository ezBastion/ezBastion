package middleware

import (
	"errors"
	"ezBastion/cmd/ezb_sta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/quasoft/websspi"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"strings"
	"unsafe"
)

var (
	auth               *websspi.Authenticator
	config             *websspi.Config
	contextKeyUserInfo = contextKey("UserInfo")
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

func EzbAuthSSPI(c *gin.Context) {

	authHead := c.GetHeader("Authorization")
	username := c.GetHeader("X-Authenticated-user")
	logg := log.WithFields(log.Fields{"middleware": "sspi"})

	if authHead != "" {
		nego := strings.Split(authHead, " ")
		if len(nego) != 2 {
			logg.Error("bad Negotiation #SSPI0001: " + authHead)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-SSPI0001"))
			return
		}
		if username != "" && strings.ToLower(nego[0]) == "negotiate" {

			// user is computed from specific module
			stauser := models.StaUser{}
			stauser.User = username
			stauser.UserGroups = ""

			// TODO compute SID and groups
			c.Set("connection", stauser)
			c.Set("aud", "ad")
			//testhash := fmt.Sprintf("%x", md5.Sum([]byte(username)))
			//c.Set("sign_key", testhash)
		}
	}
}

func SspiHandler() gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	config = websspi.NewConfig()
	auth, _ = websspi.New(config)
	auth.Config.AuthUserKey = "X-Authenticated-user"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Currently no needs to do something, websspi will try to set the userinfo
	})
	// try to use the handler to do the sspi
	h := auth.WithAuth(handler)

	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(c)
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func getContextInternals(ctx interface{}, ctxkey string) interface{} {
	contextValues := reflect.ValueOf(ctx).Elem()
	contextKeys := reflect.TypeOf(ctx).Elem()

	found := false
	if contextKeys.Kind() == reflect.Struct {
		for i := 0; i < contextValues.NumField(); i++ {
			reflectValue := contextValues.Field(i)
			reflectValue = reflect.NewAt(reflectValue.Type(), unsafe.Pointer(reflectValue.UnsafeAddr())).Elem()

			reflectField := contextKeys.Field(i)

			if reflectField.Name == "Context" {
				getContextInternals(reflectValue.Interface(), ctxkey)
			} else {
				fmt.Printf("field name: %+v\n", reflectField.Name)
				fmt.Printf("value: %+v\n", reflectValue.Interface())
				if strings.ToLower(reflectField.Name) == "key" {
					if strings.ToLower(ctxkey) == strings.ToLower(fmt.Sprintf("%+v", reflectValue.Interface())) {
						// the context value is a key and this key matches the parameter. Next try should be a "val"
						found = true
					}
				} else {
					if found && strings.ToLower(reflectField.Name) == "val" {
						// the new try is effectively a "val", we will return th interface hosted by the val
						return reflectValue.Interface()
					} else {
						found = false
					}
				}
			}
		}
	}
	return nil
}
