package group_analytics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnowledgeAreaYear_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		year KnowledgeAreaYear
		want map[string]interface{}
	}{
		{
			name: "flattens areas onto the top level alongside lugar",
			year: KnowledgeAreaYear{
				Lugar: "2025",
				Areas: map[string]int{"Innovación": 3, "Otros": 1},
			},
			want: map[string]interface{}{
				"lugar":      "2025",
				"Innovación": float64(3),
				"Otros":      float64(1),
			},
		},
		{
			name: "empty areas still produces lugar",
			year: KnowledgeAreaYear{
				Lugar: "2026",
				Areas: map[string]int{},
			},
			want: map[string]interface{}{
				"lugar": "2026",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.year)
			assert.NoError(t, err)

			var got map[string]interface{}
			assert.NoError(t, json.Unmarshal(b, &got))
			assert.Equal(t, tt.want, got)
		})
	}
}
