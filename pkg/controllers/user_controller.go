package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/middlewares"
	"github.com/gqvz/mvc/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct{}

func CreateUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users/", uc.CreateUserHandler).Methods("POST")

	editUserHandler := middlewares.Authorize(models.Any)(http.HandlerFunc(uc.EditUserHandler))
	router.Handle("/users/{id}", editUserHandler).Methods("PATCH")

	getUserHandler := middlewares.Authorize(models.Any)(http.HandlerFunc(uc.GetUserHandler))
	router.Handle("/users/{id}", getUserHandler).Methods("GET")

	getUsersHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(uc.GetUsersHandler))
	router.Handle("/users", getUsersHandler).Methods("GET")
}

type CreateUserRequest struct {
	Name     string `json:"name" example:"real"`
	Password string `json:"password" example:"realpassword"`
	Email    string `json:"email" example:"real@real.com"`
}

type CreateUserResponse struct {
	Message string `json:"message" example:"User created successfully"`
}

// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User information"
// @Success 201 {object} CreateUserResponse
// @Failure 400 {object} ErrorResponse "Bad request, missing or invalid parameters"
// @Failure 409 {object} ErrorResponse "Conflict, user with the same email or username already exists"
// @Failure 500 {object} ErrorResponse "Internal server error, failed to create user"
// @Router /users/ [post]
func (uc *UserController) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "Missing required fields: name, password, or email", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByEmailOrUsername(req.Email, req.Name)
	if err != nil {
		http.Error(w, "Error checking for existing user", http.StatusInternalServerError)
		return
	}

	if user != nil {
		http.Error(w, "User with the same email or username already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user, err = models.CreateUser(req.Name, req.Email, string(hashedPassword), models.Customer)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(CreateUserResponse{Message: "User created successfully"})
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

type EditUserRequest struct {
	Name     string `json:"name" example:"real"`
	Email    string `json:"email" example:"real@real.com"`
	Password string `json:"password" example:"realpassword"`
}

type EditUserResponse struct {
	Message string `json:"message" example:"User edited successfully"`
}

// @Summary Edit an existing user
// @Description Edit an existing user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body EditUserRequest true "Updated user information"
// @Security jwt
// @Security cookie
// @Success 200 {object} EditUserResponse
// @Failure 400 {object} ErrorResponse "Bad request, missing or invalid parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized, invalid token"
// @Failure 403 {object} ErrorResponse "Forbidden, you are not allowed to edit this user"
// @Failure 404 {object} ErrorResponse "Not Found, user with the specified ID does not exist"
// @Failure 409 {object} ErrorResponse "Conflict, user with the same email or username already exists"
// @Failure 500 {object} ErrorResponse "Internal server error, failed to edit user"
// @Router /users/{id} [patch]
func (uc *UserController) EditUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req EditUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" {
		http.Error(w, "Missing required fields: name or email", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByEmailOrUsername(req.Email, req.Name)
	if err != nil {
		http.Error(w, "Error checking for existing user", http.StatusInternalServerError)
		return
	}

	if user != nil && user.ID != id {
		http.Error(w, "User with the same email or username already exists", http.StatusConflict)
		return
	}

	user, err = models.GetUserById(id)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User with the specified ID does not exist", http.StatusNotFound)
		return
	}

	userID := r.Context().Value("userid")
	if userID != nil && user.ID != userID.(int64) && !user.Role.HasFlag(models.Admin) {
		http.Error(w, "You are not allowed to edit this user", http.StatusForbidden)
		return
	}

	var hashedPassword string
	if req.Password != "" {
		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		hashedPassword = string(hashedPasswordBytes)
	} else {
		hashedPassword = user.PasswordHash
	}

	err = models.EditUser(id, req.Name, req.Email, hashedPassword)
	if err != nil {
		http.Error(w, "Error editing user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(EditUserResponse{Message: "User edited successfully"})
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

type GetUserResponse struct {
	ID    int64       `json:"id" example:"1"`
	Name  string      `json:"name" example:"real"`
	Email string      `json:"email" example:"real@real.com"`
	Role  models.Role `json:"role" example:"1"`
}

// @Summary Get user by ID
// @Description Get user information by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int false "User ID"
// @Security jwt
// @Security cookie
// @Success 200 {object} GetUserResponse
// @Failure 400 {object} ErrorResponse "Bad request, invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized, invalid token"
// @Failure 403 {object} ErrorResponse "Forbidden, you are not allowed to view this user"
// @Failure 404 {object} ErrorResponse "Not Found, user with the specified ID does not exist"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/{id} [get]
func (uc *UserController) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	userID, err := strconv.ParseInt(idStr, 10, 64)
	role := models.Role(r.Context().Value("role").(byte))
	if userID != r.Context().Value("userid").(int64) && !role.HasFlag(models.Admin) {
		http.Error(w, "You are not allowed to get this user", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserById(userID)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User with the specified ID does not exist", http.StatusNotFound)
		return
	}

	response := GetUserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

// @Summary Get users
// @Description Get user using name, email, role
// @Tags users
// @Accept json
// @Produce json
// @Param search query string false "Filter users by name"
// @Param role query string false "Filter users by role"
// @Param limit query int false "Limit the number of users returned"
// @Param offset query int false "Offset for pagination"
// @Security jwt
// @Security cookie
// @Success 200 {object} []GetUserResponse "List of users"
// @Failure 400 {object} ErrorResponse "Bad request, invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized, invalid token"
// @Failure 403 {object} ErrorResponse "Forbidden, you are not allowed to view fetch users"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users [get]
func (uc *UserController) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	roleB, err := strconv.ParseInt(r.URL.Query().Get("role"), 10, 8)
	if err != nil {
		roleB = 0
	}
	role := models.Role(roleB)
	limitS := r.URL.Query().Get("limit")
	offsetS := r.URL.Query().Get("offset")

	limit, err := strconv.ParseInt(limitS, 10, 32)
	if err != nil || limit < 0 || limit > 20 {
		limit = 10
	}

	offset, err := strconv.ParseInt(offsetS, 10, 32)
	if err != nil || offset < 0 {
		offset = 0
	}

	userRole := models.Role(r.Context().Value("role").(byte))
	if !userRole.HasFlag(models.Admin) {
		http.Error(w, "You are not allowed to get all users", http.StatusForbidden)
		return
	}

	users, err := models.GetUsers(search, role, int(limit), int(offset))

	if err != nil {
		log.Printf("Error retrieving users: %v", err)
		http.Error(w, "Error retrieving users", http.StatusInternalServerError)
		return
	}

	userResponses := make([]GetUserResponse, len(users))

	for i, user := range users {
		userResponses[i] = GetUserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(userResponses)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

type ErrorResponse struct {
	Error   string `json:"error" example:"Error message"`
	Message string `json:"message" example:"Error message"`
}
