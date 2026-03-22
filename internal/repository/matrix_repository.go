package repository

import (
	"choice-matrix-backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MatrixRepository struct {
	db *gorm.DB
}

func NewMatrixRepository(db *gorm.DB) *MatrixRepository {
	return &MatrixRepository{db: db}
}

// Rows
func (r *MatrixRepository) CreateRow(row *models.Row) error {
	return r.db.Create(row).Error
}

func (r *MatrixRepository) GetRowsByProjectID(projectID uint) ([]models.Row, error) {
	var rows []models.Row
	err := r.db.Where("project_id = ?", projectID).Order("sort_order asc").Find(&rows).Error
	return rows, err
}

func (r *MatrixRepository) UpdateRowOrder(projectID uint, rowIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range rowIDs {
			if err := tx.Model(&models.Row{}).Where("id = ? AND project_id = ?", id, projectID).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *MatrixRepository) DeleteRow(id uint, projectID uint) error {
	return r.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&models.Row{}).Error
}

// Columns
func (r *MatrixRepository) CreateColumn(col *models.Column) error {
	return r.db.Create(col).Error
}

func (r *MatrixRepository) GetColumnsByProjectID(projectID uint) ([]models.Column, error) {
	var cols []models.Column
	err := r.db.Where("project_id = ?", projectID).Order("sort_order asc").Find(&cols).Error
	return cols, err
}

func (r *MatrixRepository) UpdateColumnOrder(projectID uint, colIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range colIDs {
			if err := tx.Model(&models.Column{}).Where("id = ? AND project_id = ?", id, projectID).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *MatrixRepository) DeleteColumn(id uint, projectID uint) error {
	return r.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&models.Column{}).Error
}

// Cells
func (r *MatrixRepository) UpsertCell(cell *models.Cell) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "row_id"}, {Name: "column_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"text_content", "numeric_value", "score_value", "updated_at"}),
	}).Create(cell).Error
}

func (r *MatrixRepository) GetCellsByProjectID(projectID uint) ([]models.Cell, error) {
	var cells []models.Cell
	err := r.db.Where("project_id = ?", projectID).Find(&cells).Error
	return cells, err
}
