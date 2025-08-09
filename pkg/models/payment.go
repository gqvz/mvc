package models

import (
	"fmt"
	"strings"
)

type PaymentStatus string

const (
	Processing PaymentStatus = "processing"
	Accepted   PaymentStatus = "accepted"
)

type Payment struct {
	ID        int64         `json:"id"`
	OrderID   int64         `json:"order_id"`
	Subtotal  float64       `json:"subtotal"`
	Tip       float64       `json:"tip"`
	Total     float64       `json:"total"`
	Status    PaymentStatus `json:"status"`
	CashierID int64         `json:"cashier_id"`
}

func CreatePayment(orderId int64, subtotal float64, tip float64, cashierId int64, userId int64) (*Payment, error) {
	res, err := DB.Exec("INSERT INTO Payments (order_id, user_id, order_subtotal, tip, status, cashier_id) VALUES (?, ?, ?, ?, ?, ?)", orderId, userId, subtotal, tip, Processing, cashierId)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Payment{
		ID:        id,
		OrderID:   orderId,
		Subtotal:  subtotal,
		Tip:       tip,
		Total:     subtotal + tip,
		Status:    Processing,
		CashierID: cashierId,
	}, nil
}

func GetPaymentByID(paymentId int64, userId int64) (*Payment, error) {
	row := DB.QueryRow("SELECT id, order_id, order_subtotal, tip, status, cashier_id FROM Payments WHERE id = ? AND (user_id = ? OR ? = 0)", paymentId, userId, userId)
	payment := &Payment{}
	err := row.Scan(&payment.ID, &payment.OrderID, &payment.Subtotal, &payment.Tip, &payment.Status, &payment.CashierID)
	if err != nil {
		return nil, err
	}
	payment.Total = payment.Subtotal + payment.Tip
	return payment, nil
}

func GetPayments(userId int64, status PaymentStatus, limit int, offset int) ([]*Payment, error) {
	query := "SELECT id, order_id, order_subtotal, tip, status, cashier_id FROM Payments WHERE 1=1"
	var args []any
	if userId > 0 {
		query += " AND user_id = ?"
		args = append(args, userId)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		payment := &Payment{}
		err := rows.Scan(&payment.ID, &payment.OrderID, &payment.Subtotal, &payment.Tip, &payment.Status, &payment.CashierID)
		if err != nil {
			return nil, err
		}
		payment.Total = payment.Subtotal + payment.Tip
		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

func UpdatePaymentStatus(paymentId int64, status PaymentStatus) error {
	res, err := DB.Exec("UPDATE Payments SET status = ? WHERE id = ?", status, paymentId)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return fmt.Errorf("payment not found")
		}
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}
