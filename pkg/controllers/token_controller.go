package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/config"
	"github.com/gqvz/mvc/pkg/middlewares"
	"github.com/gqvz/mvc/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type TokenController struct{}

func CreateTokenController() *TokenController {
	return &TokenController{}
}

func (ac *TokenController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/token", ac.CreateTokenHandler).Methods("POST")
}

type CreateTokenRequest struct {
	Username string `json:"username" example:"real"`
	Password string `json:"password" example:"realpassword"`
}

type CreateTokenResponse struct {
	Token string `json:"token"`
}

// @Summary Create a new JWT token
// @Description Create a new JWT token for user authentication
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body CreateTokenRequest true "User credentials"
// @Success 201 {object} CreateTokenResponse
// @Failure 400 {object} ErrorResponse "Bad request, missing or invalid parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized, invalid username or password"
// @Failure 500 {object} ErrorResponse "Internal server error, failed to create token"
// @Router /token [post]
func (ac *TokenController) CreateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByEmailOrUsername("", req.Username)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error comparing passwords", http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, middlewares.Claims{
		UserId: user.ID,
		Role:   byte(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "mvc",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}, nil)

	ss, err := token.SignedString([]byte(config.Config.JwtSecret))
	if err != nil {
		http.Error(w, "Error signing token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    ss,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(2 * time.Hour),
	})

	response := CreateTokenResponse{
		Token: ss,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}
