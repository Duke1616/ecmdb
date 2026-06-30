package plugin

import (
	"testing"

	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

func TestCompleteRelationSpec(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		spec     pluginx.ResourceSpec
		expected string
	}{
		{
			name: "out to target",
			base: "host",
			spec: pluginx.ResourceSpec{
				ModelUID:     "gateway",
				RelationType: "default",
				Direction:    pluginx.DirectionToTarget,
			},
			expected: "host_default_gateway",
		},
		{
			name: "in from source",
			base: "host",
			spec: pluginx.ResourceSpec{
				ModelUID:     "AuthGateway",
				RelationType: "default",
				Direction:    pluginx.DirectionToSource,
			},
			expected: "AuthGateway_default_host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildRelationName(tt.base, tt.spec)
			if err != nil {
				t.Fatalf("buildRelationName() error = %v", err)
			}
			if got != tt.expected {
				t.Fatalf("RelationName = %s, want %s", got, tt.expected)
			}
		})
	}
}
