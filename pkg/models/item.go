package models

import (
	"context"
	"encoding/json"
	"fmt"
)

type Item struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Tags        []Tag   `json:"tags"`
	ImageURL    string  `json:"image_url"`
	Available   bool    `json:"available"`
}

func CreateItem(ctx context.Context, name string, description string, price float64, tags []Tag, imageURL string, available bool) (*Item, error) {
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	res, err := tx.Exec("INSERT INTO Items (name, description, price, is_available, image_url) SELECT ?, ?, ?, ?, ? WHERE NOT EXISTS (SELECT 1 FROM Items WHERE name = ?)", name, description, price, available, imageURL, name)
	if err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			return nil, fmt.Errorf("%v %v", err1, err)
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			return nil, fmt.Errorf("%v %v", err1, err)
		}
		return nil, err
	}

	if id == 0 {
		return nil, fmt.Errorf("item with name '%s' already exists", name)
	}

	item := &Item{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		ImageURL:    imageURL,
		Available:   available,
		Tags:        tags,
	}

	query := "INSERT INTO ItemTags (item_id, tag_id) VALUES "
	args := make([]any, 0, len(tags)*2)
	for _, tag := range tags {
		query += "(?, ?),"
		args = append(args, id, tag.Id)
	}
	if len(tags) > 0 {
		query = query[:len(query)-1]
		query += ";"
		_, err = tx.Exec(query, args...)
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				return nil, fmt.Errorf("%v %v", err1, err)
			}
			return nil, err
		}
		err := tx.Commit()
		if err != nil {
			return nil, err
		}
	}

	return item, nil
}

func EditItem(ctx context.Context, id int64, name string, description string, price float64, tags []Tag, imageURL string, available bool) (*Item, error) {
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	res, err := tx.Exec("UPDATE Items SET name = ?, description = ?, price = ?, is_available = ?, image_url = ? WHERE id = ? AND NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM Items WHERE name = ? AND id != ?) AS temp_table);", name, description, price, available, imageURL, id, name, id)
	if err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			return nil, fmt.Errorf("%v %v", err1, err)
		}
		return nil, err
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		_ = tx.Rollback()
		return nil, fmt.Errorf("item not found with id '%d'", id)
	}

	_, err = tx.Exec("DELETE FROM ItemTags WHERE item_id = ?", id)
	if err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			return nil, fmt.Errorf("%v %v", err1, err)
		}
		return nil, err
	}

	query := "INSERT INTO ItemTags (item_id, tag_id) VALUES "
	args := make([]any, 0, len(tags)*2)
	for _, tag := range tags {
		query += "(?, ?),"
		args = append(args, id, tag.Id)
	}
	if len(tags) > 0 {
		query = query[:len(query)-1]
		query += ";"
		_, err = tx.Exec(query, args...)
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				return nil, fmt.Errorf("%v %v", err1, err)
			}
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	item := &Item{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		ImageURL:    imageURL,
		Available:   available,
		Tags:        tags,
	}

	return item, nil
}

func GetItemById(id int64) (*Item, error) {
	rows, err := DB.Query(`
								SELECT Items.id, Items.name, description, price, is_available, image_url,
									CONCAT('[', 
										GROUP_CONCAT(
         									JSON_OBJECT('id', Tags.id, 'name', Tags.name)
	       								),
									']') as tags
									FROM Items 
									JOIN ItemTags ON ItemTags.item_id = Items.id 
									JOIN Tags ON ItemTags.tag_id = Tags.id
									WHERE Items.id = ?
									GROUP BY Items.id
									LIMIT 1;`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var item Item
	if rows.Next() {
		var tagsJSON string
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Available, &item.ImageURL, &tagsJSON); err != nil {
			return nil, err
		}

		if tagsJSON != "[]" {
			if err := json.Unmarshal([]byte(tagsJSON), &item.Tags); err != nil {
				return nil, fmt.Errorf("failed to read tags: %v", err)
			}
		}
		return &item, nil
	} else {
		return nil, fmt.Errorf("item with id '%d' not found", id)
	}
}

func GetItems(tags []Tag, search string, available bool, limit int, offset int) ([]Item, error) {
	query := `
				SELECT Items.id, Items.name, description, price, is_available, image_url,
					CONCAT('[', 
						GROUP_CONCAT(
							JSON_OBJECT('id', Tags.id, 'name', Tags.name)
	       				),
					']') as tags
					FROM Items
					JOIN ItemTags ON ItemTags.item_id = Items.id 
					JOIN Tags ON ItemTags.tag_id = Tags.id `
	var args []any

	if search != "" {
		query += " AND (Items.name LIKE ? OR Items.description LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if !available {
		query += " AND is_available = false"
	}
	if len(tags) > 0 {
		query += "WHERE Items.id IN (SELECT item_id FROM ItemTags WHERE tag_id IN ("
		for i := range tags {
			query += "?,"
			args = append(args, tags[i].Id)
		}
		query = query[:len(query)-1] + ")) "
	} else {
		query += "WHERE 1=1 "
	}
	query += " GROUP BY Items.id LIMIT ? OFFSET ?;"
	args = append(args, limit, offset)
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var tagsJSON string
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Available, &item.ImageURL, &tagsJSON); err != nil {
			return nil, err
		}

		if tagsJSON != "[]" {
			if err := json.Unmarshal([]byte(tagsJSON), &item.Tags); err != nil {
				return nil, fmt.Errorf("failed to read tags: %v", err)
			}
		}
		items = append(items, item)
	}

	return items, nil
}

func GetItemByIdBulk(ids []int64) (*[]Item, error) {

	query := "SELECT Items.id, Items.name, description, price, is_available, image_url, CONCAT('[', GROUP_CONCAT(JSON_OBJECT('id', Tags.id, 'name', Tags.name)), ']') as tags FROM Items JOIN ItemTags ON ItemTags.item_id = Items.id JOIN Tags ON ItemTags.tag_id = Tags.id WHERE Items.id IN ("

	args := make([]any, len(ids))
	for i, id := range ids {
		query += "?,"
		args[i] = id
	}
	query = query[:len(query)-1] + ") GROUP BY Items.id;"

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var tagsJSON string
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Available, &item.ImageURL, &tagsJSON); err != nil {
			return nil, err
		}

		if tagsJSON != "[]" {
			if err := json.Unmarshal([]byte(tagsJSON), &item.Tags); err != nil {
				return nil, fmt.Errorf("failed to read tags: %v", err)
			}
		}
		items = append(items, item)
	}

	return &items, nil
}
