package mappers

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

func AnnouncementModelToEntity(model models.Announcement) entities.Announcement {
	return entities.Announcement{
		ID:        model.ID,
		Title:     model.Title,
		Content:   model.Content,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		DeletedAt: model.DeletedAt.String,
	}
}

func AnnouncementModelsToEntities(models []models.Announcement) []entities.Announcement {
	entities := make([]entities.Announcement, 0, len(models))
	for _, model := range models {
		entities = append(entities, AnnouncementModelToEntity(model))
	}
	return entities
}
