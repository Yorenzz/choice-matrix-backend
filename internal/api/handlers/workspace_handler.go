package handlers

import (
	"net/http"
	"strconv"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder := &models.Folder{
		UserID: userID,
		Name:   req.Name,
	}

	if err := h.repo.CreateFolder(folder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": folder})
}

func (h *WorkspaceHandler) GetFolders(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	folders, err := h.repo.GetFoldersByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch folders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": folders})
}

// Projects
type CreateProjectRequest struct {
	Title    string `json:"title" binding:"required"`
	FolderID *uint  `json:"folder_id"`
}

func (h *WorkspaceHandler) CreateProject(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := &models.Project{
		UserID:   userID,
		Title:    req.Title,
		FolderID: req.FolderID,
	}

	if err := h.repo.CreateProject(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": project})
}

func (h *WorkspaceHandler) GetProjects(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	projects, err := h.repo.GetProjectsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": projects})
}

func (h *WorkspaceHandler) GetProjectDetails(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	project, err := h.repo.GetProjectByID(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// TODO: Fetch matrix data (rows, columns, cells) for this project

	c.JSON(http.StatusOK, gin.H{"data": project})
}
