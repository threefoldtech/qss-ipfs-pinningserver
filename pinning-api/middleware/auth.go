package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/database"
	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
)

func ApiKeyMiddleware() gin.HandlerFunc {
	apiKeyHeader := config.CFG.Auth.ApiKeyHeader

	return func(c *gin.Context) {
		apiKey, err := bearerToken(c.Request, apiKeyHeader)
		if err != nil {
			fmt.Print("API key authentication failed", "error", err)
			return
		}

		if user_id, ok := apiKeyIsValid(c, apiKey); !ok {
			hostIP, _, err := net.SplitHostPort(c.Request.RemoteAddr)
			if err != nil {
				fmt.Print("failed to parse remote address", "error", err)
				hostIP = c.Request.RemoteAddr
			}
			fmt.Print("no matching API key found", "remoteIP", hostIP)
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.Failure{
				Error: models.FailureError{
					Reason:  "UNAUTHORIZED",
					Details: "Access token is missing or invalid",
				},
			})
			return
		} else {
			c.Set("userID", user_id)
		}
	}

}

// apiKeyIsValid checks if the given API key is valid and returns the user id if it is.
func apiKeyIsValid(c *gin.Context, rawKey string) (uint, bool) {
	// TODO: use tf-pinning-service/auth package. not implemented yet.
	// hash := sha256.Sum256([]byte(rawKey))
	users := database.NewUsersRepository()
	hash := sha256.Sum256([]byte(rawKey))
	hex_string := hex.EncodeToString(hash[:])
	user, err := users.FindByTokenHash(c, hex_string)
	if err != nil {
		return 0, false
	}
	return user.ID, true
}

// bearerToken extracts the content from the header, striping the Bearer prefix
func bearerToken(r *http.Request, header string) (string, error) {
	rawToken := r.Header.Get(header)
	pieces := strings.SplitN(rawToken, " ", 2)
	if len(pieces) < 2 {
		return "", errors.New("token with incorrect bearer format")
	}

	token := strings.TrimSpace(pieces[1])

	return token, nil
}