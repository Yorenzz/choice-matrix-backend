package handlers

import (
	"net/http"
	"strconv"
	"time"

	"choice-matrix-backend/internal/models"
	"choice-matrix-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type WorkspaceHandler struct {
	repo *repository.WorkspaceRepository
}

func NewWorkspaceHandler(repo *repository.WorkspaceRepository) *WorkspaceHandler {
	return &WorkspaceHandler{repo: repo}
}

// Folders
type CreateFolderRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *WorkspaceHandler) CreateFolder(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	folder := &models.Folder{
		UserID: userID,
		Name:   req.Name,
	}

	if err := h.repo.CreateFolder(folder); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to create folder")
		return
	}

	respondSuccess(c, http.StatusCreated, folder, "Folder created")
}

func (h *WorkspaceHandler) GetFolders(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	folders, err := h.repo.GetFoldersByUserID(userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch folders")
		return
	}
	respondSuccess(c, http.StatusOK, folders, "Folders fetched")
}

// Projects
type CreateProjectRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FolderID    *uint  `json:"folder_id"`
	IsFavorite  bool   `json:"is_favorite"`
}

func (h *WorkspaceHandler) CreateProject(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now()

	project := &models.Project{
		UserID:       userID,
		Title:        req.Title,
		Description:  req.Description,
		FolderID:     req.FolderID,
		IsFavorite:   req.IsFavorite,
		LastOpenedAt: &now,
	}

	if err := h.repo.CreateProject(project); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to create project")
		return
	}

	respondSuccess(c, http.StatusCreated, project, "Project created")
}

func (h *WorkspaceHandler) GetProjects(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	projects, err := h.repo.GetProjectsByUserID(userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch projects")
		return
	}
	respondSuccess(c, http.StatusOK, projects, "Projects fetched")
}

type UpdateProjectRequest struct {
	Title        string     `json:"title" binding:"required"`
	Description  string     `json:"description"`
	FolderID     *uint      `json:"folder_id"`
	IsFavorite   bool       `json:"is_favorite"`
	LastOpenedAt *time.Time `json:"last_opened_at"`
	ShareToken   *string    `json:"share_token"`
}

func (h *WorkspaceHandler) UpdateProject(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	fields := map[string]any{
		"title":          req.Title,
		"description":    req.Description,
		"folder_id":      req.FolderID,
		"is_favorite":    req.IsFavorite,
		"last_opened_at": req.LastOpenedAt,
		"share_token":    req.ShareToken,
	}

	if err := h.repo.UpdateProjectFields(uint(id), userID, fields); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to update project")
		return
	}

	project, err := h.repo.GetProjectByID(uint(id), userID)
	if err != nil {
		respondError(c, http.StatusNotFound, "Project not found")
		return
	}

	respondSuccess(c, http.StatusOK, project, "Project updated")
}

func (h *WorkspaceHandler) GetProjectDetails(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	project, err := h.repo.GetProjectByID(uint(id), userID)
	if err != nil {
		respondError(c, http.StatusNotFound, "Project not found")
		return
	}

	respondSuccess(c, http.StatusOK, project, "Project fetched")
}

func (h *WorkspaceHandler) DeleteProject(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.repo.DeleteProject(uint(id), userID); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"id": uint(id)}, "Project deleted")
}
