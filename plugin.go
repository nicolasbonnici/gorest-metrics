package metrics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest-metrics/migrations"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/plugin"
)

type MetricsPlugin struct {
	config Config
	db     database.Database
}

func NewPlugin() plugin.Plugin {
	return &MetricsPlugin{}
}

func (p *MetricsPlugin) Name() string {
	return "metrics"
}

func (p *MetricsPlugin) Dependencies() []string {
	return []string{}
}

func (p *MetricsPlugin) Initialize(config map[string]interface{}) error {
	p.config = DefaultConfig()

	if db, ok := config["database"].(database.Database); ok {
		p.db = db
		p.config.Database = db
	}

	if allowedTypes, ok := config["allowed_types"].([]interface{}); ok {
		types := make([]string, 0, len(allowedTypes))
		for _, t := range allowedTypes {
			if str, ok := t.(string); ok {
				types = append(types, str)
			}
		}
		if len(types) > 0 {
			p.config.AllowedTypes = types
		}
	}

	if maxKeyLength, ok := config["max_key_length"].(int); ok {
		p.config.MaxKeyLength = maxKeyLength
	}

	if onlyPositiveValues, ok := config["only_positive_values"].(bool); ok {
		p.config.OnlyPositiveValues = onlyPositiveValues
	}

	if paginationLimit, ok := config["pagination_limit"].(int); ok {
		p.config.PaginationLimit = paginationLimit
	}

	if maxPaginationLimit, ok := config["max_pagination_limit"].(int); ok {
		p.config.MaxPaginationLimit = maxPaginationLimit
	}

	return p.config.Validate()
}

func (p *MetricsPlugin) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

func (p *MetricsPlugin) SetupEndpoints(app *fiber.App) error {
	if p.db == nil {
		return nil
	}

	RegisterRoutes(app, p.db, &p.config)
	return nil
}

func (p *MetricsPlugin) MigrationSource() interface{} {
	return migrations.GetMigrations()
}

func (p *MetricsPlugin) MigrationDependencies() []string {
	return []string{}
}

func (p *MetricsPlugin) GetOpenAPIResources() []plugin.OpenAPIResource {
	return []plugin.OpenAPIResource{{
		Name:          "metric",
		PluralName:    "metrics",
		BasePath:      "/metrics",
		Tags:          []string{"Metrics"},
		ResponseModel: Metric{},
		CreateModel:   CreateMetricRequest{},
		UpdateModel:   UpdateMetricRequest{},
		Description:   "Integer metrics tracking for polymorphic resources",
	}}
}
