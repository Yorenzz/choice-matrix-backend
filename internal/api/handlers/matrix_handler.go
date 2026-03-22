package handlers

import (
	"net/http"
	"strconv"

	"choice-matrix-backend/internal/models"
	"choice-matrix-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type MatrixHandler struct {
	repo          *repository.MatrixRepository
	workspaceRepo *repository.WorkspaceRepository
}

func NewMatrixHandler(repo *repository.MatrixRepository, workspaceRepo *repository.WorkspaceRepository) *MatrixHandler {
	return &MatrixHandler{repo: repo, workspaceRepo: workspaceRepo}
}

// Helper to check project ownership
func (h *MatrixHandler) checkProjectAccess(c *gin.Context, projectID uint) bool {
	userID := c.MustGet("userID").(uint)
	_, err := h.workspaceRepo.GetProjectByID(projectID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied or project not found"})
		return false
	}
	return true
}

func parseProjectID(c *gin.Context) (uint, bool) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return 0, false
	}
	return uint(id), true
}

// --- Project Payload ---
func (h *MatrixHandler) GetProjectPayload(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	userID := c.MustGet("userID").(uint)
	project, _ := h.workspaceRepo.GetProjectByID(projectID, userID)
	rows, _ := h.repo.GetRowsByProjectID(projectID)
	cols, _ := h.repo.GetColumnsByProjectID(projectID)
	cells, _ := h.repo.GetCellsByProjectID(projectID)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"project": project,
			"rows":    rows,
			"columns": cols,
			"cells":   cells,
		},
	})
}

// --- Rows ---
type CreateRowRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *MatrixHandler) CreateRow(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	var req CreateRowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := &models.Row{
		ProjectID: projectID,
		Name:      req.Name,
	}

	if err := h.repo.CreateRow(row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create row"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": row})
}

type ReorderRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

func (h *MatrixHandler) ReorderRows(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	var req ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateRowOrder(projectID, req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder rows"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rows reordered"})
}

// --- Columns ---
type CreateColumnRequest struct {
	Title string            `json:"title" binding:"required"`
	Type  models.ColumnType `json:"type" binding:"required"`
}

func (h *MatrixHandler) CreateColumn(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	var req CreateColumnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	col := &models.Column{
		ProjectID: projectID,
		Title:     req.Title,
		Type:      req.Type,
		Weight:    1.0, // default
	}

	if err := h.repo.CreateColumn(col); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create column"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": col})
}

func (h *MatrixHandler) ReorderColumns(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	var req ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateColumnOrder(projectID, req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder columns"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Columns reordered"})
}

// --- Cells ---
type UpsertCellRequest struct {
	RowID        uint    `json:"row_id" binding:"required"`
	ColumnID     uint    `json:"column_id" binding:"required"`
	TextContent  string  `json:"text_content"`
	NumericValue float64 `json:"numeric_value"`
	ScoreValue   float64 `json:"score_value"`
}

func (h *MatrixHandler) UpsertCell(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	var req UpsertCellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cell := &models.Cell{
		ProjectID:    projectID,
		RowID:        req.RowID,
		ColumnID:     req.ColumnID,
		TextContent:  req.TextContent,
		NumericValue: req.NumericValue,
		ScoreValue:   req.ScoreValue,
	}

	if err := h.repo.UpsertCell(cell); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upsert cell"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cell})
}
