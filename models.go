package metrics

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
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

type CreateMetricRequest struct {
	Resource   string `json:"resource" validate:"required"`
	ResourceId string `json:"resourceId" validate:"required,uuid"`
	Key        string `json:"key" validate:"required"`
	Value      int    `json:"value" validate:"required"`
}

type UpdateMetricRequest struct {
	Value int `json:"value" validate:"required"`
}

func (r *CreateMetricRequest) Validate(config *Config) error {
	if !config.IsAllowedType(r.Resource) {
		return errors.New("resource type is not allowed")
	}

	if _, err := uuid.Parse(r.ResourceId); err != nil {
		return errors.New("resourceId must be a valid UUID")
	}

	r.Key = strings.TrimSpace(r.Key)
	if r.Key == "" {
		return errors.New("key cannot be empty")
	}

	if len(r.Key) > config.MaxKeyLength {
		return errors.New("key exceeds maximum length")
	}

	if config.OnlyPositiveValues && r.Value < 0 {
		return errors.New("value must be positive")
	}

	return nil
}

func (r *UpdateMetricRequest) Validate(config *Config) error {
	if config.OnlyPositiveValues && r.Value < 0 {
		return errors.New("value must be positive")
	}

	return nil
}
