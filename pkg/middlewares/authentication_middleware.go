package middlewares

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
)

type Claims struct {
	UserId int64 `json:"user_id"`
	Role   byte  `json:"role"`
	jwt.RegisteredClaims
}

func CreateAuthenticationMiddleware(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string
			authHeader := r.Header.Get("Authorization")
			if strings.Trim(authHeader, " \n") != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = authHeader[7:]
			}

			if tokenString == "" {
				cookie, err := r.Cookie("jwt")
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}
				tokenString = cookie.Value
			}

			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				next.ServeHTTP(w, r)
				return
			}

			if claims, ok := token.Claims.(*Claims); ok {
				ctx := r.Context()
				ctx = context.WithValue(ctx, "userid", claims.UserId)
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
