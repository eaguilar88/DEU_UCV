package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderContractDocumentID(t *testing.T) {
	tests := []struct {
		name     string
		contract *ProviderContract
		want     string
	}{
		{name: "initial contract", contract: &ProviderContract{ID: "7", Type: ContractTypeInitial}, want: "INTENCION-7"},
		{name: "addendum", contract: &ProviderContract{ID: "9", Type: ContractTypeAddendum}, want: "ADENDA-9"},
		{name: "not covered", contract: nil, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.contract.DocumentID())
		})
	}
}
