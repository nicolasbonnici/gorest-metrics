package metrics

import (
	"time"
)

type MetricCreateDTO struct {
	Resource   string `json:"resource"`
	ResourceId string `json:"resourceId"`
	Key        string `json:"key"`
	Value      int    `json:"value"`
}

type MetricUpdateDTO struct {
	Value int `json:"value"`
}

type MetricResponseDTO struct {
	ID         string     `json:"id"`
	Resource   string     `json:"resource"`
	ResourceID string     `json:"resourceId"`
	Key        string     `json:"key"`
	Value      int        `json:"value"`
	CreatedAt  *time.Time `json:"createdAt,omitempty"`
}
