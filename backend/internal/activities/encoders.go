package activities

import "github.com/eaguilar88/deu/internal/entities"

func activityToResponse(a entities.Activity) GetActivityResponse {
	return GetActivityResponse{
		ID:                    a.ID,
		GroupID:               a.GroupID,
		Name:                  a.Name,
		Description:           a.Description,
		Date:                  a.Date,
		KnowledgeArea:         a.KnowledgeArea,
		Allies:                a.Allies,
		EstimatedParticipants: a.EstimatedParticipants,
		ActualParticipants:    a.ActualParticipants,
		Financing:             a.Financing,
		Comments:              a.Comments,
		Files:                 filesToResponse(a.Files.GetAllFiles()),
		CreatedAt:             a.CreatedAt,
		UpdatedAt:             a.UpdatedAt,
	}
}

func activitiesToResponse(list []entities.Activity) []GetActivityResponse {
	res := make([]GetActivityResponse, 0, len(list))
	for _, a := range list {
		res = append(res, activityToResponse(a))
	}
	return res
}

func filesToResponse(files []*entities.File) []FileResponse {
	if len(files) == 0 {
		return nil
	}
	res := make([]FileResponse, 0, len(files))
	for _, f := range files {
		res = append(res, FileResponse{
			ID:      f.ID,
			Name:    f.Name,
			URL:     f.URL,
			Purpose: f.Purpose,
		})
	}
	return res
}

func createActivityEntityFromRequest(req CreateActivityRequest) entities.Activity {
	return entities.Activity{
		GroupID:               req.GroupID,
		Name:                  req.Name,
		Description:           req.Description,
		Date:                  req.Date,
		KnowledgeArea:         req.KnowledgeArea,
		Allies:                req.Allies,
		EstimatedParticipants: req.EstimatedParticipants,
		ActualParticipants:    req.ActualParticipants,
		Financing:             req.Financing,
		Comments:              req.Comments,
	}
}

func updateActivityEntityFromRequest(req UpdateActivityRequest) entities.Activity {
	return entities.Activity{
		ID:                    req.ID,
		Name:                  req.Name,
		Description:           req.Description,
		Date:                  req.Date,
		KnowledgeArea:         req.KnowledgeArea,
		Allies:                req.Allies,
		EstimatedParticipants: req.EstimatedParticipants,
		ActualParticipants:    req.ActualParticipants,
		Financing:             req.Financing,
		Comments:              req.Comments,
	}
}
