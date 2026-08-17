package activities

import "github.com/eaguilar88/deu/internal/entities"

func activityToResponse(a entities.Activity, includeParticipantList bool) GetActivityResponse {
	var coverURL string
	if a.CoverImage != nil {
		coverURL = a.CoverImage.URL
	}

	var participantListURL string
	if includeParticipantList && a.ParticipantList != nil {
		participantListURL = a.ParticipantList.URL
	}

	knowledgeArea := a.KnowledgeArea
	if knowledgeArea == nil {
		knowledgeArea = make([]string, 0)
	}

	return GetActivityResponse{
		ID:                    a.ID,
		GroupID:               a.GroupID,
		GroupName:             a.GroupName,
		Name:                  a.Name,
		Description:           a.Description,
		DateStart:             a.DateStart,
		DateEnd:               a.DateEnd,
		Location:              a.Location,
		KnowledgeArea:         knowledgeArea,
		Allies:                a.Allies,
		GroupParticipants:     a.GroupParticipants,
		EstimatedParticipants: a.EstimatedParticipants,
		ActualParticipants:    a.ActualParticipants,
		Financing:             a.Financing,
		Comments:              a.Comments,
		CoverImage:            coverURL,
		ParticipantList:       participantListURL,
		GalleryURL:            a.GalleryURL,
		ReportChecked:         a.ReportChecked,
		IsFeatured:            a.IsFeatured,
		CreatedAt:             a.CreatedAt,
		UpdatedAt:             a.UpdatedAt,
	}
}

func activitiesToResponse(list []entities.Activity) []GetActivityResponse {
	res := make([]GetActivityResponse, 0, len(list))
	for _, a := range list {
		res = append(res, activityToResponse(a, false))
	}
	return res
}

func createActivityEntityFromRequest(req CreateActivityRequest) entities.Activity {
	knowledgeArea := req.KnowledgeArea
	if knowledgeArea == nil {
		knowledgeArea = make([]string, 0)
	}

	return entities.Activity{
		GroupID:               req.GroupID,
		Name:                  req.Name,
		Description:           req.Description,
		DateStart:             req.DateStart,
		DateEnd:               req.DateEnd,
		Location:              req.Location,
		KnowledgeArea:         knowledgeArea,
		Allies:                req.Allies,
		GroupParticipants:     req.GroupParticipants,
		EstimatedParticipants: req.EstimatedParticipants,
		ActualParticipants:    req.ActualParticipants,
		Financing:             req.Financing,
		Comments:              req.Comments,
		GalleryURL:            req.GalleryURL,
	}
}

func updateActivityEntityFromRequest(req UpdateActivityRequest) entities.Activity {
	knowledgeArea := req.KnowledgeArea
	if knowledgeArea == nil {
		knowledgeArea = make([]string, 0)
	}

	return entities.Activity{
		ID:                    req.ID,
		Name:                  req.Name,
		Description:           req.Description,
		DateStart:             req.DateStart,
		DateEnd:               req.DateEnd,
		Location:              req.Location,
		KnowledgeArea:         knowledgeArea,
		Allies:                req.Allies,
		GroupParticipants:     req.GroupParticipants,
		EstimatedParticipants: req.EstimatedParticipants,
		ActualParticipants:    req.ActualParticipants,
		Financing:             req.Financing,
		Comments:              req.Comments,
		GalleryURL:            req.GalleryURL,
	}
}
