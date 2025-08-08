package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/models"
	"net/http"
	"strconv"
	"strings"
)

type OrderItemController struct {
}

func CreateOrderItemController() *OrderItemController {
	return &OrderItemController{}
}

type CreateOrderItemRequest struct {
	ItemID             int64  `json:"item_id"`
	Quantity           int    `json:"quantity"`
	CustomInstructions string `json:"custom_instructions"`
}

type CreateOrderItemResponse struct {
	OrderItemID int64 `json:"order_item_id"`
}

// @Summary Create a new order item
// @Description Create a new order item
// @Tags order_items
// @Accept json
// @Produce json
// @Security jwt
// @Security cookie
// @Param id path int true "Order ID"
// @Param request body CreateOrderItemRequest true "Create Order Item Request"
// @Success 200 {object} CreateOrderItemResponse
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 409 {object} string "Conflict"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders/{id}/items [post]
func (c *OrderItemController) CreateOrderItem(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ItemID <= 0 || req.Quantity <= 0 {
		http.Error(w, "Item ID and quantity must be greater than zero", http.StatusBadRequest)
		return
	}

	orderIdStr := mux.Vars(r)["id"]
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userid").(int64)

	orderItem, err := models.CreateOrderItem(orderId, userId, req.ItemID, req.Quantity, req.CustomInstructions)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to create order item: ", http.StatusInternalServerError)
		}
		return
	}

	response := CreateOrderItemResponse{
		OrderItemID: orderItem.ID,
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type EditOrderItemStatusRequest struct {
	Status models.ItemStatus `json:"status"`
}

// @Summary Edit an order item status
// @Description Edit the status of an order item
// @Tags order_items
// @Accept json
// @Produce json
// @Security jwt
// @Security cookie
// @Param id path int true "Order Item ID"
// @Param status body EditOrderItemStatusRequest true "New status"
// @Success 200 {object} string "Order item status updated"
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Not Found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders/items/{id}/ [patch]
func (c *OrderItemController) EditOrderItemStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderItemIdS, ok := vars["id"]
	if !ok {
		http.Error(w, "Order item ID is required", http.StatusBadRequest)
		return
	}

	orderItemId, err := strconv.ParseInt(orderItemIdS, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order item ID", http.StatusBadRequest)
		return
	}

	var req EditOrderItemStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = models.EditOrderItemStatus(orderItemId, req.Status)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Order item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update order item status: ", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode("Order item status updated"); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type GetOrderItemResponse = models.OrderItem

// @Summary Get order items
// @Description Get all items in an order
// @Tags order_items
// @Security jwt
// @Security cookie
// @Param id path int true "Order ID"
// @Success 200 {array} GetOrderItemResponse "List of order items"
// @Success 204 {object} string "No Content"
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Not Found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders/{id}/items [get]
func (c *OrderItemController) GetOrderItems(w http.ResponseWriter, r *http.Request) {
	orderIdStr := mux.Vars(r)["id"]
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userid").(int64)
	role := models.Role(r.Context().Value("role").(byte))
	if role.HasFlag(models.Admin) {
		userId = 0
	}

	orderItems, err := models.GetItemsByOrderId(orderId, userId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve order items", http.StatusInternalServerError)
		}
		return
	}

	if orderItems == nil {
		http.Error(w, "No items found for this order", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orderItems); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Get order items by status
// @Description Get all items in an order by status
// @Tags order_items
// @Security jwt
// @Security cookie
// @Param limit query int false "Limit the number of items returned"
// @Param offset query int false "Offset for pagination"
// @Param status query string false "Filter by item status (preparing, completed)"
// @Success 200 {array} GetOrderItemResponse "List of order items"
// @Success 204 {object} string "No Content"
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Not Found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders/items [get]
func (c *OrderItemController) GetOrderItemsByStatus(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	status := r.URL.Query().Get("status")

	if status == "" {
		status = string(models.Preparing)
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 20 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	orderItems, err := models.GetOrderItems(models.ItemStatus(status), limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve order items", http.StatusInternalServerError)
		return
	}

	if len(orderItems) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orderItems); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
