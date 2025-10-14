package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "online-disk-server/internal/auth"
)

const CtxUserID = "userID"

func AuthRequired(jwtm *auth.JWTManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        authz := c.GetHeader("Authorization")
        if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
            return
        }
        token := strings.TrimPrefix(authz, "Bearer ")
        claims, err := jwtm.Parse(token)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        // sub stored as float64 by default in MapClaims
        if sub, ok := claims["sub"].(float64); ok {
            c.Set(CtxUserID, uint(sub))
        }
        c.Next()
    }
}
