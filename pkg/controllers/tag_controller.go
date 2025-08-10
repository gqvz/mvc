package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/models"
	"net/http"
	"strconv"
	"strings"
)

type TagController struct{}

func CreateTagController() *TagController {
	return &TagController{}
}

type CreateTagRequest struct {
	Name string `json:"name" example:"real"`
} // @name CreateTagRequest

type CreateTagResponse struct {
	ID int64 `json:"id"`
} // @name CreateTagResponse

// @Summary Create tag
// @ID createTag
// @Description Create a new tag
// @Tags tags
// @Accept json
// @Produce json
// @Param tag body CreateTagRequest true "Tag request"
// @Security jwt
// @Success 201 {object} CreateTagResponse "Created tag"
// @Failure 400 {object} string "Bad request, invalid tag name"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to create tags"
// @Failure 409 {object} string "Conflict, tag with the same name already exists"
// @Failure 500 {object} string "Internal server error"
// @Router /tags [post]
func (c *TagController) CreateTagHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Tag name is required", http.StatusBadRequest)
		return
	}

	tag, err := models.CreateTag(req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			http.Error(w, "Tag with the same name already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create tag", http.StatusInternalServerError)
		return
	}

	response := CreateTagResponse{
		ID: tag.ID,
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println("Error encoding response:", err)
	}
}

type GetTagResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
} // @name GetTagResponse

// @Summary Get tag by ID
// @ID getTagById
// @Description Get a tag by its ID
// @Tags tags
// @Accept json
// @Produce json
// @Param id path int true "Tag ID"
// @Security jwt
// @Success 200 {object} GetTagResponse "Tag details"
// @Failure 400 {object} string "Bad request, invalid tag ID"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to view
// @Failure 404 {object} string "Tag not found"
// @Failure 500 {object} string "Internal server error"
// @Router /tags/{id} [get]
func (c *TagController) GetTagHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	tag, err := models.GetTagById(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve tag", http.StatusInternalServerError)
		return
	}

	response := GetTagResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println("Error encoding response:", err)
		return
	}
}

// @Summary Get tags
// @ID getTags
// @Description Get all tags
// @Tags tags
// @Accept json
// @Produce json
// @Security jwt
// @Success 200 {array} GetTagResponse "List of tags"
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to view tags"
// @Failure 500 {object} string "Internal server error"
// @Router /tags [get]
func (c *TagController) GetTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := models.GetTags()
	if err != nil {
		http.Error(w, "Failed to retrieve tags", http.StatusInternalServerError)
		return
	}

	var tagResponses []GetTagResponse
	for _, tag := range tags {
		tagResponses = append(tagResponses, GetTagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tagResponses)
	if err != nil {
		fmt.Println("Error encoding response:", err)
		return
	}
}

type EditTagRequest = CreateTagRequest // @name EditTagRequest
type EditTagResponse = GetTagResponse  // @name EditTagResponse

// @Summary Edit tag
// @ID editTag
// @Description Edit an existing tag
// @Tags tags
// @Accept json
// @Produce json
// @Param id path int true "Tag ID"
// @Param tag body EditTagRequest true "Tag request"
// @Security jwt
// @Success 200 {object} EditTagResponse "Updated tag"
// @Failure 400 {object} string "Bad request, invalid tag name or ID
// @Failure 401 {object} string "Unauthorized, invalid token"
// @Failure 403 {object} string "Forbidden, you are not allowed to edit
// @Failure 404 {object} string "Tag not found"
// @Failure 409 {object} string "Conflict, tag with the same name already exists"
// @Failure 500 {object} string "Internal server error"
// @Router /tags/{id} [put]
func (c *TagController) EditTagHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	var req EditTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Tag name is required", http.StatusBadRequest)
		return
	}

	tag, err := models.EditTag(id, req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		} else if strings.Contains(err.Error(), "Duplicate") {
			http.Error(w, "Tag with the same name already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to edit tag", http.StatusInternalServerError)
		return
	}

	response := EditTagResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println("Error encoding response:", err)
		return
	}
}
