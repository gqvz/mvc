package models

import (
	"database/sql"
	"fmt"
)

type ItemStatus string // @name ItemStatus

const (
	Preparing ItemStatus = "preparing"
	Completed ItemStatus = "completed"
) // @name ItemStatus

type OrderItem struct {
	ID                 int64      `json:"id"`
	OrderID            int64      `json:"order_id"`
	ItemID             int64      `json:"item_id"`
	Quantity           int        `json:"quantity"`
	CustomInstructions string     `json:"custom_instructions"`
	Status             ItemStatus `json:"status"`
} // @name OrderItem

func CreateOrderItem(orderId int64, userId int64, itemId int64, quantity int, customInstructions string) (*OrderItem, error) {
	res, err := DB.Exec("INSERT INTO OrderItems (order_id, item_id, count, status, custom_instructions) SELECT ?, ?, ?, ?, ? FROM Orders WHERE id = ? AND customer_id = ? AND status = 'open'", orderId, itemId, quantity, Preparing, customInstructions, orderId, userId)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	if id == 0 {
		return nil, fmt.Errorf("order not found")
	}

	return &OrderItem{
		ID:                 id,
		OrderID:            orderId,
		ItemID:             itemId,
		Quantity:           quantity,
		CustomInstructions: customInstructions,
		Status:             Preparing,
	}, nil
}

func EditOrderItemStatus(orderItemId int64, status ItemStatus) error {
	res, err := DB.Exec("UPDATE OrderItems SET status = ? WHERE id = ?", status, orderItemId)
	if err != nil {
		return fmt.Errorf("failed to update order item status: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("order item not found")
	}
	return nil
}

func GetItemsByOrderId(orderId int64, userId int64) (*[]OrderItem, error) {
	rows, err := DB.Query("SELECT id, order_id, item_id, count, custom_instructions, status FROM OrderItems WHERE order_id = ? AND EXISTS (SELECT 1 FROM Orders WHERE id = ? AND (customer_id = ? OR ? = 0))", orderId, orderId, userId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order items: %w", err)
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		if err := scanOrderItem(rows, &item); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return &items, nil
}

func GetOrderItems(status ItemStatus, limit int, offset int) ([]OrderItem, error) {
	rows, err := DB.Query("SELECT id, order_id, item_id, count, custom_instructions, status FROM OrderItems WHERE status = ? LIMIT ? OFFSET ?", status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		if err := scanOrderItem(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func scanOrderItem(rows *sql.Rows, item *OrderItem) error {
	if err := rows.Scan(&item.ID, &item.OrderID, &item.ItemID, &item.Quantity, &item.CustomInstructions, &item.Status); err != nil {
		return fmt.Errorf("failed to scan order item: %w", err)
	}
	return nil
}
