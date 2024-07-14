package entities

import "strconv"

const (
	DefaultPerPage = 10
	DefaultPage    = 1
)

type PageScope struct {
	Page    int
	PerPage int
}

func (p PageScope) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p *PageScope) GetPageFromVars(raw string) error {
	p.Page = DefaultPage
	if raw == "" {
		return nil
	}
	var err error
	p.Page, err = strconv.Atoi(raw)
	return err // if not a number (helps debugging)
}

func (p *PageScope) GetPerPageFromVars(raw string) error {
	if raw == "" {
		p.PerPage = DefaultPerPage
		return nil
	}
	var err error
	p.PerPage, err = strconv.Atoi(raw)
	if p.PerPage <= 0 {
		p.PerPage = DefaultPerPage
	}
	return err // if not a number (helps debugging)
}
