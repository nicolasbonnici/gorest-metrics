package metrics

import (
	"time"
)

type Metric struct {
	Id         string     `json:"id,omitempty" db:"id"`
	Resource   string     `json:"resource" db:"resource"`
	ResourceId string     `json:"resourceId" db:"resource_id"`
	Key        string     `json:"key" db:"name"`
	Value      int        `json:"value" db:"value"`
	CreatedAt  *time.Time `json:"createdAt,omitempty" db:"created_at"`
}

func (Metric) TableName() string {
	return "metrics"
}
