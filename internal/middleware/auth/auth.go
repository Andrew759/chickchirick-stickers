package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const UserUUIDKey = "auth_user_uuid"

type validateResponse struct {
	Payload struct {
		Valid    bool   `json:"valid"`
		UserUUID string `json:"user_uuid"`
	} `json:"payload"`
	Error string `json:"error"`
}

// RequireUser validates the same access_token cookie used by other Chickchirick services.
func RequireUser(authURL string, client *http.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("access_token")
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}
		if strings.TrimSpace(authURL) == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "auth service is not configured"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(authURL, "/")+"/auth/validate", nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to build auth request"})
			return
		}
		req.AddCookie(cookie)
		resp, err := client.Do(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "auth service unavailable"})
			return
		}
		defer resp.Body.Close()
		var result validateResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "failed to parse auth response"})
			return
		}
		if resp.StatusCode != http.StatusOK || !result.Payload.Valid || result.Payload.UserUUID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(UserUUIDKey, result.Payload.UserUUID)
		c.Next()
	}
}
