package entities

type ProviderRequest struct {
	ID         int64
	ProviderID int64
	Status     RequestStatus
	ReviewerID string
	Comments   string
	ReviewedAt string
	CreatedAt  string
	UpdatedAt  string
	Provider   *Provider
}
