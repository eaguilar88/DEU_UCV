package entities

import "fmt"

type ContractType string

const (
	ContractTypeInitial  ContractType = "inicial"
	ContractTypeAddendum ContractType = "adenda"

	// File purposes for provider contracts
	ProviderFileTypeAddendum = "adenda"
)

func (ct ContractType) String() string {
	return string(ct)
}

// ProviderContract represents one legal-document submission event for a
// provider: either the initial contract (carta_intencion + carta_compromiso)
// or an addendum (adenda), each of which covers a set of the provider's
// approved courses.
type ProviderContract struct {
	ID             string
	ProviderID     string
	Type           ContractType
	CoveredCourses []string
	Files          []*File
	CreatedAt      string
}

// DocumentID identifies the contract the way the diplomados UI expects it:
// INTENCION-<id> for the initial contract and ADENDA-<id> for an addendum.
// A nil contract (course not covered) yields an empty string.
func (c *ProviderContract) DocumentID() string {
	if c == nil {
		return ""
	}
	if c.Type == ContractTypeAddendum {
		return fmt.Sprintf("ADENDA-%s", c.ID)
	}
	return fmt.Sprintf("INTENCION-%s", c.ID)
}
