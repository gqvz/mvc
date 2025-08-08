package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/models"
	"net/http"
	"strconv"
	"strings"
)

type PaymentController struct {
}

func CreatePaymentController() *PaymentController {
	return &PaymentController{}
}

type CreatePaymentRequest struct {
	OrderID   int64   `json:"order_id"`
	Tip       float64 `json:"tip"`
	CashierID int64   `json:"cashier_id"`
}

type CreatePaymentResponse struct {
	PaymentID int64 `json:"payment_id"`
}

// @Summary Create a new payment
// @Description Create a new payment
// @Tags payments
// @Accept json
// @Produce json
// @Security jwt
// @Security cookie
// @Param request body CreatePaymentRequest true "Create Payment Request"
// @Success 201 {object} CreatePaymentResponse
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 500 {object} string "Internal Server Error"
// @Router /payments [post]
func (c *PaymentController) CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Tip < 0 {
		http.Error(w, "Tip cannot be negative", http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value("userid").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orderItems, err := models.GetItemsByOrderId(req.OrderID, userId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve order items: ", http.StatusInternalServerError)
		return
	}

	if orderItems == nil || len(*orderItems) == 0 {
		http.Error(w, "No items found for the order", http.StatusBadRequest)
		return
	}

	subtotal := 0.0

	itemIds := make([]int64, 0, len(*orderItems))
	for _, item := range *orderItems {
		itemIds = append(itemIds, item.ItemID)
	}
	items, err := models.GetItemByIdBulk(itemIds)
	if err != nil {
		http.Error(w, "Failed to retrieve items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for i, item := range *orderItems {
		itemDetails := (*items)[i]
		subtotal += itemDetails.Price * float64(item.Quantity)
	}
	payment, err := models.CreatePayment(req.OrderID, subtotal, req.Tip, req.CashierID, userId)

	if err != nil {
		http.Error(w, "Failed to create payment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := CreatePaymentResponse{
		PaymentID: payment.ID,
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type GetPaymentResponse = models.Payment

// @Summary Get a payment by ID
// @Description Get a payment by ID
// @Tags payments
// @Accept json
// @Produce json
// @Security jwt
// @Security cookie
// @Param id path int true "Payment ID"
// @Success 200 {object} GetPaymentResponse
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Not Found"
// @Router /payments/{id} [get]
func (c *PaymentController) GetPaymentHandler(w http.ResponseWriter, r *http.Request) {
	paymentIdStr := mux.Vars(r)["id"]
	paymentId, err := strconv.ParseInt(paymentIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value("userid").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	payment, err := models.GetPaymentByID(paymentId, userId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Payment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve payment", http.StatusInternalServerError)
		return
	}

	response := GetPaymentResponse{
		ID:        payment.ID,
		OrderID:   payment.OrderID,
		Subtotal:  payment.Subtotal,
		Tip:       payment.Tip,
		Total:     payment.Total,
		Status:    payment.Status,
		CashierID: payment.CashierID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Get payments
// @Description Get all payments with filters
// @Tags payments
// @Accept json
// @Produce json
// @Security jwt
// @Security cookie
// @Param status query string false "Payment status"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param user_id query int false "User ID"
// @Success 200 {array} GetPaymentResponse
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 500 {object} string "Internal Server Error"
// @Router /payments [get]
func (c *PaymentController) GetPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	status := models.PaymentStatus(r.URL.Query().Get("status"))
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	userIdStr := r.URL.Query().Get("user_id")

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil && userIdStr != "" {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	limit := 10
	offset := 0

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 20 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
	}

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
	}

	currentUserId, ok := r.Context().Value("userid").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role := models.Role(r.Context().Value("role").(byte))
	if !role.HasFlag(models.Admin) {
		userId = currentUserId
	}
	payments, err := models.GetPayments(userId, status, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve payments", http.StatusInternalServerError)
		return
	}

	if payments == nil || len(payments) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payments); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type EditPaymentStatusRequest struct {
	Status models.PaymentStatus `json:"status"`
}

// @Summary Edit payment status
// @Description Edit payment status
// @Tags payments
// @Accept json
// @Produce json
// @Security jwt
// @Security cookie
// @Param id path int true "Payment ID"
// @Param request body EditPaymentStatusRequest true "Edit Payment Status Request"
// @Success 200 {object} string "Payment status updated"
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Not Found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /payments/{id} [patch]
func (c *PaymentController) EditPaymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	paymentIdStr := mux.Vars(r)["id"]
	paymentId, err := strconv.ParseInt(paymentIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}

	var req EditPaymentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status != models.Processing && req.Status != models.Accepted {
		http.Error(w, "Invalid payment status", http.StatusBadRequest)
		return
	}

	err = models.UpdatePaymentStatus(paymentId, req.Status)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Payment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update payment status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode("Payment status updated"); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
