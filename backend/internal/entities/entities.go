package entities

import "strconv"

const (
	DefaultPerPage = 10
	DefaultPage    = 1
)

type PageScope struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
	Count   int `json:"count,omitempty"`
}

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
