package metrics

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/query"
)

type MetricHooks struct {
	config *Config
}

func NewMetricHooks(config *Config) *MetricHooks {
	return &MetricHooks{
		config: config,
	}
}

func (h *MetricHooks) CreateHook(c *fiber.Ctx, dto MetricCreateDTO, model *Metric) error {
	if !h.config.IsAllowedType(dto.Resource) {
		return fiber.NewError(400, "resource type is not allowed")
	}

	if _, err := uuid.Parse(dto.ResourceId); err != nil {
		return fiber.NewError(400, "resourceId must be a valid UUID")
	}

	key := strings.TrimSpace(dto.Key)
	if key == "" {
		return fiber.NewError(400, "key cannot be empty")
	}

	if len(key) > h.config.MaxKeyLength {
		return fiber.NewError(400, "key exceeds maximum length")
	}

	if h.config.OnlyPositiveValues && dto.Value < 0 {
		return fiber.NewError(400, "value must be positive")
	}

	model.Key = key

	return nil
}

func (h *MetricHooks) UpdateHook(c *fiber.Ctx, dto MetricUpdateDTO, model *Metric) error {
	if h.config.OnlyPositiveValues && dto.Value < 0 {
		return fiber.NewError(400, "value must be positive")
	}

	return nil
}

func (h *MetricHooks) GetAllHook(c *fiber.Ctx, conditions *[]query.Condition, orderBy *[]crud.OrderByClause) error {
	return nil
}
