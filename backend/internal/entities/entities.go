package entities

import "strconv"

type PageScope struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
	Count   int `json:"count,omitempty"`
}

type RequestStatus string

type RequestType string

const (
	// Request statuses
	RequestStatus_CREATED      RequestStatus = "created"
	RequestStatus_APPROVED     RequestStatus = "approved"
	RequestStatus_REJECTED     RequestStatus = "rejected"
	RequestStatus_UNDER_REVIEW RequestStatus = "under_review"
	// Request types
	RequestType_GROUP  RequestType = "grupo de extensión"
	RequestType_COURSE RequestType = "diplomado"
	// Pagination defaults
	DefaultPerPage = 10
	DefaultPage    = 1
)

func (p PageScope) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p *PageScope) GetPageFromVars(raw string) error {
	p.Page = DefaultPage
	if raw == "" {
		return nil
	}

	page, err := strconv.Atoi(raw)
	if err != nil || p.Page <= 0 {
		return err
	}

	p.Page = page

	return nil
}

func (p *PageScope) GetPerPageFromVars(raw string) error {
	p.PerPage = DefaultPerPage
	if raw == "" {
		return nil
	}

	perPage, err := strconv.Atoi(raw)
	if err != nil || p.PerPage <= 0 {
		return err
	}

	p.PerPage = perPage

	return nil
}
