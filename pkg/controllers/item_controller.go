package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ItemController struct {
}

func CreateItemController() *ItemController {
	return &ItemController{}
}

type CreateItemRequest struct {
	Name        string   `json:"name" example:"real"`
	Price       float64  `json:"price" example:"69.69"`
	Description string   `json:"description" example:"real"`
	ImageURL    string   `json:"image_url" example:"https://http.cat/404"`
	Tags        []string `json:"tags" example:"real,tag"`
	Available   bool     `json:"available" example:"true"`
}

type CreateItemResponse struct {
	ID int64 `json:"id"`
}

// @Summary Create item
// @Description Create a new item
// @Tags items
// @Accept json
// @Produce json
// @Param item body CreateItemRequest true "Item request"
// @Security jwt
// @Security cookie
// @Success 201 {object} CreateItemResponse "Created item"
// @Failure 400 {object} string "Bad request, invalid item data"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to create items"
// @Failure 409 {object} string "Conflict, item with this name already exists"
// @Failure 500 {object} string "Internal server error"
// @Router /items [post]
func (c *ItemController) CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Price <= 0 {
		http.Error(w, "Name and price are required", http.StatusBadRequest)
		return
	}

	var tags []models.Tag
	allTags, err := models.GetTags()
	if err != nil {
		http.Error(w, "Failed to get tags", http.StatusInternalServerError)
		return
	}
	for _, tagName := range req.Tags {
		found := false
		for _, tag := range allTags {
			if tag.Name == tagName {
				tags = append(tags, tag)
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Tag not found: "+tagName, http.StatusBadRequest)
			return
		}
	}

	item, err := models.CreateItem(r.Context(), req.Name, req.Description, req.Price, tags, req.ImageURL, req.Available)
	if err != nil {
		if strings.Contains(err.Error(), "exists") {
			http.Error(w, "Item with this name already exists", http.StatusConflict)
		}
		http.Error(w, "Failed to create item", http.StatusInternalServerError)
		return
	}

	response := CreateItemResponse{ID: item.ID}
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

type GetItemResponse struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       float64      `json:"price"`
	Tags        []models.Tag `json:"tags"`
	ImageURL    string       `json:"image_url"`
	Available   bool         `json:"available"`
}

// @Summary Get item by ID
// @Description Get an item by its ID
// @Tags items
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Security jwt
// @Security cookie
// @Success 200 {object} GetItemResponse "Item details"
// @Failure 400 {object} string "Bad request, invalid item ID"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to view
// @Failure 404 {object} string "Item not found"
// @Failure 500 {object} string "Internal server error"
// @Router /items/{id} [get]
func (c *ItemController) GetItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idS, ok := vars["id"]
	if !ok {
		http.Error(w, "Item ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idS, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := models.GetItemById(id)
	if err != nil {
		http.Error(w, "Failed to get item", http.StatusInternalServerError)
		log.Printf("Error retrieving item: %v", err)
		return
	}
	if item == nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	response := GetItemResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		Tags:        item.Tags,
		ImageURL:    item.ImageURL,
		Available:   item.Available,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

// @Summary Get items
// @Description Get all items with optional filters
// @Tags items
// @Accept json
// @Produce json
// @Param tags query string false "Filter by tags (comma-separated)"
// @Param search query string false "Search items by name or description"
// @Param available query bool false "Filter by availability"
// @Param limit query int false "Limit number of items returned"
// @Param offset query int false "Offset for pagination"
// @Security jwt
// @Security cookie
// @Success 200 {array} GetItemResponse "List of items"
// @Failure 400 {object} string "Bad request, invalid parameters"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to view
// @Failure 500 {object} string "Internal server error"
// @Router /items [get]
func (c *ItemController) GetItemsHandler(w http.ResponseWriter, r *http.Request) {
	tagsParam := r.URL.Query().Get("tags")
	searchParam := r.URL.Query().Get("search")
	availableParam := r.URL.Query().Get("available")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	var tags []models.Tag
	allTags, err := models.GetTags()
	if err != nil {
		http.Error(w, "Failed to get tags", http.StatusInternalServerError)
		return
	}

	if tagsParam == "" {
		tags = allTags
	} else {
		for _, tagName := range strings.Split(strings.TrimSpace(tagsParam), ",") {
			found := false
			for _, tag := range allTags {
				if tag.Name == tagName {
					tags = append(tags, tag)
					found = true
					break
				}
			}
			if !found {
				http.Error(w, "Tag not found: "+tagName, http.StatusBadRequest)
				return
			}
		}
	}
	available := true
	if availableParam != "" {
		availableBool, err := strconv.ParseBool(availableParam)
		if err != nil {
			http.Error(w, "Invalid value for 'available' parameter", http.StatusBadRequest)
			return
		}
		available = availableBool
	}

	limit := 10
	if limitParam != "" {
		limitInt, err := strconv.Atoi(limitParam)
		if err != nil || limitInt <= 0 || limitInt > 20 {
			http.Error(w, "Invalid value for 'limit' parameter", http.StatusBadRequest)
			return
		}
		limit = limitInt
	}

	offset := 0
	if offsetParam != "" {
		offsetInt, err := strconv.Atoi(offsetParam)
		if err != nil || offsetInt < 0 {
			http.Error(w, "Invalid value for 'offset' parameter", http.StatusBadRequest)
			return
		}
		offset = offsetInt
	}

	items, err := models.GetItems(tags, searchParam, available, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get items", http.StatusInternalServerError)
		log.Printf("Error retrieving items: %v", err)
		return
	}

	responseItems := make([]GetItemResponse, len(items))
	for i, item := range items {
		responseItems[i] = GetItemResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Tags:        item.Tags,
			ImageURL:    item.ImageURL,
			Available:   item.Available,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(responseItems)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

type EditItemRequest = CreateItemRequest

// @Summary Edit item
// @Description Edit an existing item
// @Tags items
// @Accept json
// @Param id path int true "Item ID"
// @Param item body EditItemRequest true "Item request"
// @Security jwt
// @Security cookie
// @Success 200 "Edited item"
// @Failure 400 {object} string "Bad request, invalid item data"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to edit
// @Failure 404 {object} string "Item not found"
// @Failure 500 {object} string "Internal server error"
// @Router /items/{id} [put]
func (c *ItemController) EditItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idS, ok := vars["id"]
	if !ok {
		http.Error(w, "Item ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idS, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var req EditItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Price <= 0 {
		http.Error(w, "Name and price are required", http.StatusBadRequest)
		return
	}

	var tags []models.Tag
	allTags, err := models.GetTags()
	for _, tagName := range req.Tags {
		found := false
		for _, tag := range allTags {
			if tag.Name == tagName {
				tags = append(tags, tag)
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Tag not found: "+tagName, http.StatusBadRequest)
			return
		}
	}
	if err != nil {
		http.Error(w, "Failed to get tags", http.StatusInternalServerError)
		return
	}

	item, err := models.EditItem(r.Context(), id, req.Name, req.Description, req.Price, tags, req.ImageURL, req.Available)
	if err != nil {
		http.Error(w, "Failed to edit item", http.StatusInternalServerError)
		log.Printf("Error editing item: %v", err)
		return
	}
	if item == nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)

}
