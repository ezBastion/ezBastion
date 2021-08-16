// This file is part of ezBastion.

//     ezBastion is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     ezBastion is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.

package middleware

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"ezBastion/cmd/ezb_db/Middleware"
	"fmt"
	"io/ioutil"

	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func EzbAuthJWT(c *gin.Context) {

	logg := log.WithFields(log.Fields{"Middleware": "jwt"})

	authHead := c.GetHeader("Authorization")
	if authHead != "" {
		var err error
		bearer := strings.Split(authHead, " ")
		if len(bearer) != 2 {
			logg.Error("bad Authorization #J0001: " + authHead)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-JWT0001"))
			return
		}
		if strings.Compare(strings.ToLower(bearer[0]), "bearer") != 0 {
			if strings.Compare(strings.ToLower(bearer[0]), "negotiate") == 0 {
				return
			}
			logg.Error("bad Authorization #J0002: " + authHead)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-JWT0002"))
			return
		}
		tokenString := bearer[1]
		ex, _ := os.Executable()
		exPath := filepath.Dir(ex)
		parts := strings.Split(tokenString, ".")
		p, err := base64.RawStdEncoding.DecodeString(parts[1])
		if err != nil {
			logg.Error("Unable to decode payload: ", err)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-JWT0009"))
			return
		}
		var payload Middleware.Payload
		err = json.Unmarshal(p, &payload)
		if err != nil {
			logg.Error("Unable to parse payload: ", err)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-JWT0011"))
			return
		}
		jwtkeyfile := fmt.Sprintf("%s.crt", payload.ISS)
		jwtpubkey := path.Join(exPath, "cert", jwtkeyfile)

		if _, err := os.Stat(jwtpubkey); os.IsNotExist(err) {
			logg.Error("Unable to load sta public certificat: ", err)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-JWT0010"))
			return
		}

		key, _ := ioutil.ReadFile(jwtpubkey)
		var ecdsaKey *ecdsa.PublicKey
		if ecdsaKey, err = jwt.ParseECPublicKeyFromPEM(key); err != nil {
			logg.Error("Unable to parse ECDSA public key: ", err)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-JWT0003"))
		}
		methode := jwt.GetSigningMethod("ES256")
		// parts := strings.Split(tokenString, ".")
		err = methode.Verify(strings.Join(parts[0:2], "."), parts[2], ecdsaKey)
		if err != nil {
			logg.Error("Error while verifying key: ", err)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-JWT0004"))
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return jwt.ParseECPublicKeyFromPEM(key)
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("jwt", token)
			c.Set("user", claims["sub"])
		} else {
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-JWT0005"))
			logg.Error(err)
			return
		}
	}
}
