package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

type CreateGroupRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type UpdateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type GroupResponse struct {
	ID          string `json:"id"`
	CompanyID   string `json:"company_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	TaskCount   int    `json:"task_count"`
	CreatedAt   string `json:"created_at"`
}

func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := requestValidator.Struct(req); err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	companyID := GetCompanyID(r.Context())
	if companyID == nil {
		http.Error(w, "no company", http.StatusForbidden)
		return
	}

	group, err := h.groupService.CreateGroup(r.Context(), *companyID, req.Name, req.Description, req.Color)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapGroupResp(group))
}

func (h *GroupHandler) List(w http.ResponseWriter, r *http.Request) {
	companyID := GetCompanyID(r.Context())
	if companyID == nil {
		http.Error(w, "no company", http.StatusForbidden)
		return
	}

	groups, err := h.groupService.GetGroupsByCompany(r.Context(), *companyID)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	resp := make([]GroupResponse, len(groups))
	for i, g := range groups {
		resp[i] = mapGroupResp(g)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *GroupHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	group, err := h.groupService.GetGroup(r.Context(), id)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapGroupResp(group))
}

func (h *GroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	group, err := h.groupService.UpdateGroup(r.Context(), id, req.Name, req.Description, req.Color)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapGroupResp(group))
}

func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.groupService.DeleteGroup(r.Context(), id); err != nil {
		handleAuthError(err, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapGroupResp(g *domain.ProjectGroup) GroupResponse {
	return GroupResponse{
		ID:          g.ID.String(),
		CompanyID:   g.CompanyID.String(),
		Name:        g.Name,
		Description: g.Description,
		Color:       g.Color,
		TaskCount:   g.TaskCount,
		CreatedAt:   g.CreatedAt.Format(time.RFC3339),
	}
}
