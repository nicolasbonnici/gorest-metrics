package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxFilterValuesPerField(t *testing.T) {
	assert.Equal(t, 50, MaxFilterValuesPerField, "MaxFilterValuesPerField should be 50")
}

func TestMetricResource_Initialization(t *testing.T) {
	resource := &MetricResource{}
	assert.NotNil(t, resource)
}
