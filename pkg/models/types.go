package models

import "time"

type Role byte // @name Role

const (
	Any      Role              = iota // @name Any
	Customer                          // @name Customer
	Chef                              // @name Chef
	Admin    = Customer | Chef        // @name Admin
)

func (r Role) HasFlag(flag Role) bool {
	return r&flag == flag
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
} // @name Tag

type UserSeenStatus string // @name UserSeenStatus

const (
	Seen   UserSeenStatus = "seen"
	Unseen UserSeenStatus = "unseen"
)

type RequestStatus string // @name RequestStatus

const (
	Pending  RequestStatus = "pending"
	Granted  RequestStatus = "granted"
	Rejected RequestStatus = "rejected"
)

type Request struct {
	ID         int64          `json:"id"`
	UserID     int64          `json:"user_id"`
	Role       Role           `json:"role"`
	Status     RequestStatus  `json:"status"`
	UserStatus UserSeenStatus `json:"user_status"`
} // @name Request

type PaymentStatus string // @name PaymentStatus

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
} // @name Payment

type ItemStatus string // @name ItemStatus

const (
	Preparing ItemStatus = "preparing"
	Completed ItemStatus = "completed"

	ItemPending ItemStatus = "pending"
) // @name ItemStatus

type OrderItem struct {
	ID                 int64      `json:"id"`
	OrderID            int64      `json:"order_id"`
	ItemID             int64      `json:"item_id"`
	Quantity           int        `json:"quantity"`
	CustomInstructions string     `json:"custom_instructions"`
	Status             ItemStatus `json:"status"`
} // @name OrderItem

type OrderStatus string // @name OrderStatus

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
} // @name Order

type Item struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Tags        []Tag   `json:"tags"`
	ImageURL    string  `json:"image_url"`
	Available   bool    `json:"available"`
} // @name Item
