package models

import (
	"database/sql"
	"errors"
)

type Role byte

const (
	Any Role = iota
	Customer
	Chef
	Admin = Customer | Chef
)

func (r Role) HasFlag(flag Role) bool {
	return r&flag == flag
}

type User struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Role         Role   `json:"role"`
}

func CreateUser(name string, email string, passwordHash string, role Role) (*User, error) {
	result, err := DB.Exec("INSERT INTO Users (name, email, password_hash, role) VALUES (?, ?, ?, ?)", name, email, passwordHash, role)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           id,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}, nil
}

func GetUserByEmailOrUsername(email string, name string) (*User, error) {
	var user User
	err := DB.QueryRow("SELECT id, name, email, password_hash, role FROM Users WHERE email = ? OR name = ? LIMIT 1;", email, name).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func GetUserById(id int64) (*User, error) {
	var user User
	err := DB.QueryRow("SELECT id, name, email, password_hash, role FROM Users WHERE id = ? LIMIT 1;", id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func EditUser(id int64, name string, email string, password string) error {
	_, err := DB.Exec("UPDATE Users SET name = ?, email = ?, password_hash = ? WHERE id = ?", name, email, password, id)
	return err
}

func GetUsers(search string, role Role, limit int, offset int) ([]User, error) {
	if search == "" && role == Any {
		rows, err := DB.Query("SELECT id, name, email, password_hash, role FROM Users LIMIT ? OFFSET ?;", limit, offset)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role); err != nil {
				return nil, err
			}
			users = append(users, user)
		}
		return users, nil
	}

	query := "SELECT id, name, email, password_hash, role FROM Users WHERE 1=1"
	var args []any
	if search != "" {
		query += " AND (name LIKE ? OR email LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if role != Any {
		query += " AND role & ?"
		args = append(args, role)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
