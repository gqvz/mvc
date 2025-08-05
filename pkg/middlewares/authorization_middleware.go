package middlewares

import (
	"github.com/gqvz/mvc/pkg/models"
	"net/http"
)

func Authorize(requiredRole models.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleB, ok := r.Context().Value("role").(byte)
			role := models.Role(roleB)
			if !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			hasPerms := role.HasFlag(requiredRole)

			if !hasPerms {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
