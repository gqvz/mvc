package middlewares

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gqvz/mvc/pkg/services"
	"log"
	"net/http"
	"strings"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	Role   byte  `json:"role"`
	jwt.RegisteredClaims
}

func CreateAuthenticationMiddleware(jwtSecret string) func(next http.Handler) http.Handler {
	tokenCache := services.GetTokenCache()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string
			authHeader := r.Header.Get("Authorization")
			if strings.Trim(authHeader, " \n") != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = authHeader[7:]
			}

			if tokenString == "" {
				next.ServeHTTP(w, r)
				return
			}

			cachedToken, exists := tokenCache.GetToken(tokenString)
			if exists {
				ctx := r.Context()
				ctx = context.WithValue(ctx, "userid", cachedToken.UserID)
				ctx = context.WithValue(ctx, "role", cachedToken.Role)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}

			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(*Claims); ok {
				tokenCache.AddToken(tokenString, claims.UserID, claims.Role, claims.ExpiresAt.Time)

				ctx := r.Context()
				ctx = context.WithValue(ctx, "userid", claims.UserID)
				ctx = context.WithValue(ctx, "role", claims.Role)

				r = r.WithContext(ctx)

				next.ServeHTTP(w, r)
				return
			}

			log.Printf("Error processing token claims")
			http.Error(w, "Failed to process token claims", http.StatusInternalServerError)
		})
	}
}
