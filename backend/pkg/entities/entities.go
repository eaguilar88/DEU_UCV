package entities

type PageScope struct {
	Page    int
	PerPage int
}

func (p PageScope) Offset() int {
	return (p.Page - 1) * p.PerPage
}

type Pagination struct {
	Page         int
	PerPage      int
	TotalPages   int
	TotalResults int
}
