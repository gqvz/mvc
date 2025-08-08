package models

import (
	"fmt"
	"time"
)

type OrderStatus string

const (
	Open   OrderStatus = "open"
	Closed OrderStatus = "closed"
)

type Order struct {
	ID          int64       `json:"id"`
	CustomerID  int64       `json:"customer_id"`
	Status      OrderStatus `json:"status"`
	TableNumber int         `json:"table_number"`
	OrderedAt   time.Time   `json:"ordered_at"`
}

func CreateOrder(userId int64, tableNumber int) (*Order, error) {
	result, err := DB.Exec("INSERT INTO Orders (customer_id, status, table_number, ordered_at) SELECT ?, 'open', ?, ? WHERE NOT EXISTS (SELECT 1 FROM Orders WHERE table_number = ? AND status = 'open')", userId, tableNumber, time.Now(), tableNumber)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if id == 0 {
		return nil, fmt.Errorf("order for table %d already exists", tableNumber)
	}

	return &Order{
		ID:          id,
		CustomerID:  userId,
		Status:      Open,
		TableNumber: tableNumber,
		OrderedAt:   time.Now(),
	}, nil
}

func GetOrderById(id int64, userId int64) (*Order, error) {
	var order Order
	err := DB.QueryRow("SELECT id, customer_id, status, table_number, ordered_at FROM Orders WHERE id = ? AND (customer_id = ? or ? = 0) LIMIT 1;", id, userId, userId).Scan(
		&order.ID, &order.CustomerID, &order.Status, &order.TableNumber, &order.OrderedAt)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func EditOrderStatus(id int64, userId int64, status OrderStatus) error {
	result, err := DB.Exec("UPDATE Orders SET status = ? WHERE id = ? AND (customer_id = ? OR ? = 0)", status, id, userId, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func GetOrders(userId int64, status OrderStatus, tableNumber int, date time.Time, limit int, offset int) ([]*Order, error) {
	query := "SELECT id, customer_id, status, table_number, ordered_at FROM Orders WHERE 1=1"
	var args []any

	if userId > 0 {
		query += " AND customer_id = ?"
		args = append(args, userId)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	if tableNumber > 0 {
		query += " AND table_number = ?"
		args = append(args, tableNumber)
	}

	if !date.IsZero() {
		query += " AND DATE(ordered_at) = ?"
		args = append(args, date.Format("2006-01-02"))
	}

	query += " ORDER BY ordered_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.CustomerID, &order.Status, &order.TableNumber, &order.OrderedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}
