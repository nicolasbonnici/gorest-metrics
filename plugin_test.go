package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPlugin(t *testing.T) {
	plugin := NewPlugin()

	assert.NotNil(t, plugin, "NewPlugin() should not return nil")
	assert.Equal(t, "metrics", plugin.Name(), "Plugin name should be 'metrics'")
}

func TestMetricsPlugin_Name(t *testing.T) {
	plugin := &MetricsPlugin{}
	assert.Equal(t, "metrics", plugin.Name())
}

func TestMetricsPlugin_Dependencies(t *testing.T) {
	plugin := &MetricsPlugin{}
	deps := plugin.Dependencies()
	assert.Empty(t, deps, "MetricsPlugin should have no dependencies")
}

func TestMetricsPlugin_MigrationDependencies(t *testing.T) {
	plugin := &MetricsPlugin{}
	deps := plugin.MigrationDependencies()
	assert.Empty(t, deps, "MetricsPlugin should have no migration dependencies")
}

func TestMetricsPlugin_Initialize(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "empty config uses defaults",
			config:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "valid config",
			config: map[string]interface{}{
				"allowed_types":        []interface{}{"post", "user"},
				"max_key_length":       255,
				"only_positive_values": false,
				"pagination_limit":     50,
				"max_pagination_limit": 200,
			},
			wantErr: false,
		},
		{
			name: "partial config",
			config: map[string]interface{}{
				"allowed_types":    []interface{}{"post", "product"},
				"max_key_length":   100,
				"pagination_limit": 25,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &MetricsPlugin{}
			err := plugin.Initialize(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsPlugin_GetOpenAPIResources(t *testing.T) {
	plugin := &MetricsPlugin{}
	resources := plugin.GetOpenAPIResources()

	assert.Len(t, resources, 1, "Should return exactly one OpenAPI resource")

	resource := resources[0]
	assert.Equal(t, "metric", resource.Name)
	assert.Equal(t, "metrics", resource.PluralName)
	assert.Equal(t, "/metrics", resource.BasePath)
	assert.Contains(t, resource.Tags, "Metrics")
	assert.NotEmpty(t, resource.Description)
}

func TestMetricsPlugin_MigrationSource(t *testing.T) {
	plugin := &MetricsPlugin{}
	source := plugin.MigrationSource()
	assert.NotNil(t, source, "MigrationSource should not return nil")
}
