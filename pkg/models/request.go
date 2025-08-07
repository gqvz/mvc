package models

import "fmt"

type UserSeenStatus string

const (
	Seen   UserSeenStatus = "seen"
	Unseen UserSeenStatus = "unseen" // whytf did i name this unseen
)

type RequestStatus string

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
}

func CreateRequest(userID int64, role Role) (*Request, error) {
	res, err := DB.Exec("INSERT INTO Requests (user_id, role, status, user_status) VALUES (?, ?, ?, ?)", userID, role, Pending, Unseen)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Request{
		ID:         id,
		UserID:     userID,
		Role:       role,
		Status:     Pending,
		UserStatus: Unseen,
	}, nil
}

func EditRequestStatus(id int64, status RequestStatus, userId int64) error {
	res, err := DB.Exec("UPDATE Requests SET status = ?, granted_by = ? WHERE id = ?", status, userId, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("request not found")
	}
	return nil
}

func EditRequestUserStatus(id int64, userId int64, userStatus UserSeenStatus) error {
	res, err := DB.Exec("UPDATE Requests SET user_status = ? WHERE id = ? AND user_id = ?", userStatus, id, userId)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("request not found")
	}
	return nil
}

func GetRequests(userID int64, role Role, status RequestStatus, seenStatus UserSeenStatus, limit int, offset int) ([]Request, error) {
	query := "SELECT id, user_id, role, status, user_status FROM Requests WHERE 1=1"
	var args []any
	if userID != 0 {
		query += " AND user_id = ?"
		args = append(args, userID)
	}

	if role != 0 {
		query += " AND role = ?"
		args = append(args, role)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	if seenStatus != "" {
		query += " AND user_status = ?"
		args = append(args, seenStatus)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		var request Request
		if err := rows.Scan(&request.ID, &request.UserID, &request.Role, &request.Status, &request.UserStatus); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}
