package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/middlewares"
	"github.com/gqvz/mvc/pkg/models"
	"net/http"
	"strconv"
	"strings"
)

type RequestController struct {
}

func CreateRequestController() *RequestController {
	return &RequestController{}
}

func (c *RequestController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/requests", c.CreateRequestHandler).Methods("POST")

	router.HandleFunc("/requests", c.GetRequestsHandler).Methods("GET")

	grantRequestHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.GrantRequestHandler))
	router.Handle("/requests/{id:[0-9]+}/grant", grantRequestHandler).Methods("POST")

	rejectRequestHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.RejectRequestHandler))
	router.Handle("/requests/{id:[0-9]+}/reject", rejectRequestHandler).Methods("POST")

	router.HandleFunc("/requests/{id:[0-9]+}/seen", c.MarkRequestSeenHandler).Methods("POST")
}

type CreateRequestRequest struct {
	Role models.Role `json:"role" example:"3"`
}

// @Summary Create request
// @Description Create a new request for a role
// @Tags requests
// @Accept json
// @Param request body CreateRequestRequest true "Request data"
// @Security jwt
// @Security cookie
// @Success 201 {object} string "Request created successfully"
// @Failure 400 {object} string "Bad request, invalid request data"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to create requests
// @Failure 500 {object} string "Internal server error"
// @Router /requests [post]
func (c *RequestController) CreateRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == 0 {
		http.Error(w, "Role is required", http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value("userid").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := models.CreateRequest(userId, req.Role); err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type GetRequestResponse = models.Request

// @Summary Get requests
// @Description Get requests filtered by user, role, status
// @Tags requests
// @Accept json
// @Produce json
// @Param user query int false "Filter by user ID"
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Param limit query int false "Number of requests to return"
// @Param offset query int false "Offset for pagination"
// @Security jwt
// @Security cookie
// @Success 200 {array} GetRequestResponse "List of requests"
// @Success 204 {object} string "No content, no requests found"
// @Failure 400 {object} string "Bad request, invalid query parameters"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to view requests
// @Failure 500 {object} string "Internal server error"
// @Router /requests [get]
func (c *RequestController) GetRequestsHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userid").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := strconv.ParseInt(r.URL.Query().Get("user"), 10, 64)
	if err != nil && r.URL.Query().Get("user") != "" {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	userRole := models.Role(r.Context().Value("role").(byte))
	if (user == 0 || user != userId) && !userRole.HasFlag(models.Admin) {
		http.Error(w, "Forbidden, you are not allowed to view requests for other users", http.StatusForbidden)
		return
	}
	role, err := strconv.ParseInt(r.URL.Query().Get("role"), 10, 8)
	if err != nil && r.URL.Query().Get("role") != "" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}
	status := r.URL.Query().Get("status")

	var seenStatus models.UserSeenStatus
	if userRole.HasFlag(models.Admin) {
		seenStatus = ""
	} else {
		seenStatus = models.Unseen
	}

	limitS := r.URL.Query().Get("limit")
	limit := 10
	if limitS != "" {
		var err error
		limit, err = strconv.Atoi(limitS)
		if err != nil || limit < 0 || limit > 20 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
	}

	offsetS := r.URL.Query().Get("offset")
	offset := 0
	if offsetS != "" {
		var err error
		offset, err = strconv.Atoi(offsetS)
		if err != nil || offset < 0 {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
	}

	requests, err := models.GetRequests(user, models.Role(byte(role)), models.RequestStatus(status), seenStatus, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve requests", http.StatusInternalServerError)
		return
	}

	if len(requests) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	responses := make([]GetRequestResponse, len(requests))
	for i, req := range requests {
		responses[i] = GetRequestResponse{
			ID:         req.ID,
			UserID:     req.UserID,
			Role:       req.Role,
			Status:     req.Status,
			UserStatus: req.UserStatus,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Grant Request
// @Description Grant a request for a role
// @Tags requests
// @Accept json
// @Param id path int true "Request ID"
// @Security jwt
// @Security cookie
// @Success 200 {object} string "Request granted successfully"
// @Failure 400 {object} string "Bad request, invalid request ID"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to grant requests
// @Failure 404 {object} string "Not found, request does not exist"
// @Failure 500 {object} string "Internal server error"
// @Router /requests/{id}/grant [post]
func (c *RequestController) GrantRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value("userid").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := models.EditRequestStatus(id, models.Granted, userId); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Not found, request does not exist", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to grant request", http.StatusInternalServerError)
		return
	}
	// TODO give the role to the user

	w.WriteHeader(http.StatusOK)
}

// @Summary Reject Request
// @Description Reject a request for a role
// @Tags requests
// @Accept json
// @Param id path int true "Request ID"
// @Security jwt
// @Security cookie
// @Success 200 {object} string "Request rejected successfully"
// @Failure 400 {object} string "Bad request, invalid request ID"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to reject requests
// @Failure 404 {object} string "Not found, request does not exist"
// @Failure 500 {object} string "Internal server error"
// @Router /requests/{id}/reject [post]
func (c *RequestController) RejectRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value("userid").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := models.EditRequestStatus(id, models.Rejected, userId); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Not found, request does not exist", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to reject request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Mark Request as Seen
// @Description Mark a request as seen by the user
// @Tags requests
// @Accept json
// @Param id path int true "Request ID"
// @Security jwt
// @Security cookie
// @Success 200 {object} string "Request marked as seen successfully"
// @Failure 400 {object} string "Bad request, invalid request ID"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 404 {object} string "Not found, request does not exist"
// @Failure 500 {object} string "Internal server error"
// @Router /requests/{id}/seen [post]
func (c *RequestController) MarkRequestSeenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value("userid").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := models.EditRequestUserStatus(id, userId, models.Seen); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Not found, request does not exist", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to mark request as seen", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
