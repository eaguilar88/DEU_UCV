package entities

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
