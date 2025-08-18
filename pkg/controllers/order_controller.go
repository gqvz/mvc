package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type OrderController struct {
}

func CreateOrderController() *OrderController {
	return &OrderController{}
}

type CreateOrderRequest struct {
	TableNumber int `json:"table_number"`
} // @name CreateOrderRequest

type CreateOrderResponse struct {
	OrderID int64 `json:"order_id"`
} // @name CreateOrderResponse

// @Summary Create a new order
// @ID createOrder
// @Description Create a new order
// @Tags orders
// @Accept json
// @Produce json
// @Security jwt
// @Param request body CreateOrderRequest true "Create Order Request"
// @Success 200 {object} CreateOrderResponse
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 409 {object} string "Conflict"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders [post]
func (c *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TableNumber <= 0 || req.TableNumber > 100 {
		http.Error(w, "Table number must be greater than 0 and less than 100", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userid").(int64)
	order, err := models.CreateOrder(userId, req.TableNumber)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to create order", http.StatusInternalServerError)
		}
		return
	}

	response := CreateOrderResponse{
		OrderID: order.ID,
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Close an order
// @ID closeOrderById
// @Description Close an order by ID
// @Tags orders
// @Security jwt
// @Param id path int true "Order ID"
// @Success 204 "No Content"
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Not Found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders/{id}/close [post]
func (c *OrderController) CloseOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIdS, ok := vars["id"]
	if !ok {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	orderId, err := strconv.ParseInt(orderIdS, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userid").(int64)
	role := models.Role(r.Context().Value("role").(byte))
	if role.HasFlag(models.Admin) {
		userId = 0
	}
	err = models.EditOrderStatus(orderId, userId, models.Closed)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "forbidden") {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			http.Error(w, "Failed to close order", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type GetOrderResponse = models.Order // @name GetOrderResponse

// @Summary Get order by ID
// @ID getOrderById
// @Description Get an order by its ID
// @Tags orders
// @Security jwt
// @Param id path int true "Order ID"
// @Success 200 {object} GetOrderResponse "Order details"
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Not Found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders/{id} [get]
func (c *OrderController) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIdS, ok := vars["id"]
	if !ok {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	orderId, err := strconv.ParseInt(orderIdS, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userid").(int64)
	role := models.Role(r.Context().Value("role").(byte))
	if role.HasFlag(models.Admin) {
		userId = 0
	}
	order, err := models.GetOrderById(orderId, userId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve order", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Get orders
// @ID getOrders
// @Description Get order filtered by table number, date, user, status
// @Tags orders
// @Security jwt
// @Param table_number query int false "Table number"
// @Param date query string false "Date in format YYYY-MM-DD"
// @Param user_id query int false "User ID"
// @Param status query string false "Order status (open, closed)"
// @Param limit query int false "Limit for number of orders"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} GetOrderResponse "List of orders"
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders [get]
func (c *OrderController) GetOrders(w http.ResponseWriter, r *http.Request) {
	tableNumberStr := r.URL.Query().Get("table_number")
	dateStr := r.URL.Query().Get("date")
	userIdStr := r.URL.Query().Get("user_id")
	status := models.OrderStatus(r.URL.Query().Get("status"))

	var tableNumber int
	if tableNumberStr != "" {
		var err error
		tableNumber, err = strconv.Atoi(tableNumberStr)
		if err != nil || tableNumber <= 0 {
			http.Error(w, "Invalid table number", http.StatusBadRequest)
			return
		}
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil && dateStr != "" {
		http.Error(w, "Invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	var userId int64
	if userIdStr != "" {
		var err error
		userId, err = strconv.ParseInt(userIdStr, 10, 64)
		if err != nil || userId <= 0 {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
	}

	role := models.Role(r.Context().Value("role").(byte))
	currentUserId := r.Context().Value("userid").(int64)
	if !role.HasFlag(models.Admin) || userIdStr == "" {
		userId = currentUserId
	}

	limitStr := r.URL.Query().Get("limit")
	var limit int
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 20 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
	} else {
		limit = 10
	}

	offsetStr := r.URL.Query().Get("offset")
	var offset int
	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
	} else {
		offset = 0
	}

	orders, err := models.GetOrders(userId, status, tableNumber, date, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve orders", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("[]"))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
