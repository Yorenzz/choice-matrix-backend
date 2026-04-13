package repository

import (
	"choice-matrix-backend/internal/models"

	"gorm.io/gorm"
)

type WorkspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// Folders
func (r *WorkspaceRepository) CreateFolder(folder *models.Folder) error {
	return r.db.Create(folder).Error
}

func (r *WorkspaceRepository) GetFoldersByUserID(userID uint) ([]models.Folder, error) {
	var folders []models.Folder
	err := r.db.Where("user_id = ?", userID).Find(&folders).Error
	return folders, err
}

func (r *WorkspaceRepository) DeleteFolder(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Folder{}).Error
}

// Projects
func (r *WorkspaceRepository) CreateProject(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *WorkspaceRepository) GetProjectsByUserID(userID uint) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Where("user_id = ?", userID).Order("updated_at desc").Find(&projects).Error
	return projects, err
}

func (r *WorkspaceRepository) GetProjectByID(id uint, userID uint) (*models.Project, error) {
	var project models.Project
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *WorkspaceRepository) UpdateProjectFields(id uint, userID uint, fields map[string]any) error {
	return r.db.Model(&models.Project{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(fields).
		Error
}

func (r *WorkspaceRepository) DeleteProject(id uint, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", id).Delete(&models.Cell{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&models.Row{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&models.Column{}).Error; err != nil {
			return err
		}

		return tx.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Project{}).Error
	})
}
