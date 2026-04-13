package handlers

import (
	"encoding/json"
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
		respondError(c, http.StatusForbidden, "Access denied or project not found")
		return false
	}
	return true
}

func parseUintParam(c *gin.Context, name string, label string) (uint, bool) {
	idParam := c.Param(name)
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "Invalid "+label+" format")
		return 0, false
	}
	return uint(id), true
}

func parseProjectID(c *gin.Context) (uint, bool) {
	return parseUintParam(c, "id", "project ID")
}

func encodeOptions(options []string) string {
	if len(options) == 0 {
		return "[]"
	}

	encoded, err := json.Marshal(options)
	if err != nil {
		return "[]"
	}

	return string(encoded)
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

	respondSuccess(c, http.StatusOK, gin.H{
		"project": project,
		"rows":    rows,
		"columns": cols,
		"cells":   cells,
	}, "Project payload fetched")
}

// --- Rows ---
type CreateRowRequest struct {
	Name     string `json:"name" binding:"required"`
	Subtitle string `json:"subtitle"`
}

func (h *MatrixHandler) CreateRow(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	var req CreateRowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	row := &models.Row{
		ProjectID: projectID,
		Name:      req.Name,
		Subtitle:  req.Subtitle,
	}

	if err := h.repo.CreateRow(row); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to create row")
		return
	}

	respondSuccess(c, http.StatusCreated, row, "Row created")
}

type UpdateRowRequest struct {
	Name     string `json:"name"`
	Subtitle string `json:"subtitle"`
}

func (h *MatrixHandler) UpdateRow(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	rowID, ok := parseUintParam(c, "rowId", "row ID")
	if !ok {
		return
	}

	var req UpdateRowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.repo.UpdateRow(rowID, projectID, map[string]any{
		"name":     req.Name,
		"subtitle": req.Subtitle,
	}); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to update row")
		return
	}

	rows, err := h.repo.GetRowsByProjectID(projectID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch row")
		return
	}

	for _, row := range rows {
		if row.ID == rowID {
			respondSuccess(c, http.StatusOK, row, "Row updated")
			return
		}
	}

	respondError(c, http.StatusNotFound, "Row not found")
}

func (h *MatrixHandler) DeleteRow(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	rowID, ok := parseUintParam(c, "rowId", "row ID")
	if !ok {
		return
	}

	if err := h.repo.DeleteRow(rowID, projectID); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to delete row")
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"id": rowID}, "Row deleted")
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
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.repo.UpdateRowOrder(projectID, req.IDs); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to reorder rows")
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"ids": req.IDs}, "Rows reordered")
}

// --- Columns ---
type CreateColumnRequest struct {
	Title   string            `json:"title" binding:"required"`
	Type    models.ColumnType `json:"type" binding:"required"`
	Weight  float64           `json:"weight"`
	Unit    string            `json:"unit"`
	Options []string          `json:"options"`
}

func (h *MatrixHandler) CreateColumn(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	var req CreateColumnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	weight := req.Weight
	if weight == 0 {
		weight = 1
	}

	col := &models.Column{
		ProjectID: projectID,
		Title:     req.Title,
		Type:      req.Type,
		Weight:    weight,
		Unit:      req.Unit,
		Options:   encodeOptions(req.Options),
	}

	if err := h.repo.CreateColumn(col); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to create column")
		return
	}

	respondSuccess(c, http.StatusCreated, col, "Column created")
}

type UpdateColumnRequest struct {
	Title   string            `json:"title"`
	Type    models.ColumnType `json:"type"`
	Weight  float64           `json:"weight"`
	Unit    string            `json:"unit"`
	Options []string          `json:"options"`
}

func (h *MatrixHandler) UpdateColumn(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	columnID, ok := parseUintParam(c, "columnId", "column ID")
	if !ok {
		return
	}

	var req UpdateColumnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.Weight == 0 {
		req.Weight = 1
	}

	if err := h.repo.UpdateColumn(columnID, projectID, map[string]any{
		"title":   req.Title,
		"type":    req.Type,
		"weight":  req.Weight,
		"unit":    req.Unit,
		"options": encodeOptions(req.Options),
	}); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to update column")
		return
	}

	columns, err := h.repo.GetColumnsByProjectID(projectID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch column")
		return
	}

	for _, column := range columns {
		if column.ID == columnID {
			respondSuccess(c, http.StatusOK, column, "Column updated")
			return
		}
	}

	respondError(c, http.StatusNotFound, "Column not found")
}

func (h *MatrixHandler) DeleteColumn(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	columnID, ok := parseUintParam(c, "columnId", "column ID")
	if !ok {
		return
	}

	if err := h.repo.DeleteColumn(columnID, projectID); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to delete column")
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"id": columnID}, "Column deleted")
}

func (h *MatrixHandler) ReorderColumns(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	var req ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.repo.UpdateColumnOrder(projectID, req.IDs); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to reorder columns")
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"ids": req.IDs}, "Columns reordered")
}

// --- Cells ---
type UpsertCellRequest struct {
	RowID        uint     `json:"row_id" binding:"required"`
	ColumnID     uint     `json:"column_id" binding:"required"`
	TextContent  string   `json:"text_content"`
	Note         string   `json:"note"`
	NumericValue *float64 `json:"numeric_value"`
	ScoreValue   *float64 `json:"score_value"`
	SelectValue  *string  `json:"select_value"`
}

func (h *MatrixHandler) UpsertCell(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok || !h.checkProjectAccess(c, projectID) {
		return
	}

	var req UpsertCellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	cell := &models.Cell{
		ProjectID:    projectID,
		RowID:        req.RowID,
		ColumnID:     req.ColumnID,
		TextContent:  req.TextContent,
		Note:         req.Note,
		NumericValue: req.NumericValue,
		ScoreValue:   req.ScoreValue,
		SelectValue:  req.SelectValue,
	}

	if err := h.repo.UpsertCell(cell); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to upsert cell")
		return
	}

	respondSuccess(c, http.StatusOK, cell, "Cell saved")
}
