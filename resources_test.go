package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxFilterValuesPerField(t *testing.T) {
	assert.Equal(t, 50, MaxFilterValuesPerField, "MaxFilterValuesPerField should be 50")
}

func TestMetricResource_Initialization(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"post", "user"},
		PaginationLimit:    50,
		MaxPaginationLimit: 200,
	}

	resource := &MetricResource{
		Config:             config,
		PaginationLimit:    config.PaginationLimit,
		PaginationMaxLimit: config.MaxPaginationLimit,
	}

	assert.NotNil(t, resource)
	assert.Equal(t, config, resource.Config)
	assert.Equal(t, 50, resource.PaginationLimit)
	assert.Equal(t, 200, resource.PaginationMaxLimit)
}
