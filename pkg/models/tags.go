package models

import "fmt"

type Tag struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func CreateTag(name string) (*Tag, error) {
	rows, err := DB.Query("SELECT id, name FROM Tags WHERE name = ? LIMIT 1;", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tag Tag
	if rows.Next() {
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("tag with name '%s' already exists", name)
	} else {
		res, err := DB.Exec("INSERT INTO Tags (name) VALUES (?);", name)
		if err != nil {
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		tag.Id = id
		tag.Name = name
		return &tag, nil
	}
}

func GetTagById(id int64) (*Tag, error) {
	rows, err := DB.Query("SELECT id, name FROM Tags WHERE id = ? LIMIT 1;", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tag Tag
	if rows.Next() {
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, err
		}
		return &tag, nil
	} else {
		return nil, fmt.Errorf("tag with id '%d' not found", id)
	}
}

func EditTag(id int64, name string) (*Tag, error) {
	rows, err := DB.Query("SELECT id, name FROM Tags WHERE id = ? LIMIT 1;", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tag Tag
	if rows.Next() {
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("tag with id '%d' not found", id)
	}

	if tag.Name == name {
		return &tag, nil
	}

	res, err := DB.Exec("UPDATE Tags SET name = ? WHERE id = ?;", name, id)
	if err != nil {
		return nil, err
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return nil, fmt.Errorf("no rows updated for tag with id '%d'", id)
	}

	tag.Name = name
	return &tag, nil
}

func GetTags() ([]Tag, error) {
	rows, err := DB.Query("SELECT id, name FROM Tags;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}
