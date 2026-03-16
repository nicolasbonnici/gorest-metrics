package metrics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/processor"
)

const MaxFilterValuesPerField = 50

type MetricResource struct {
	processor processor.Processor[Metric, MetricCreateDTO, MetricUpdateDTO, MetricResponseDTO]
}

func RegisterMetricRoutes(app *fiber.App, db database.Database, config *Config) {
	metricCRUD := crud.New[Metric](db)
	hooks := NewMetricHooks(config)
	converter := &MetricConverter{}

	fieldMapping := map[string]string{
		"id":         "id",
		"resource":   "resource",
		"resourceId": "resource_id",
		"name":       "name",
		"value":      "value",
		"createdAt":  "created_at",
	}

	proc := processor.New(processor.ProcessorConfig[Metric, MetricCreateDTO, MetricUpdateDTO, MetricResponseDTO]{
		DB:                 db,
		CRUD:               metricCRUD,
		Converter:          converter,
		PaginationLimit:    config.PaginationLimit,
		PaginationMaxLimit: config.MaxPaginationLimit,
		FieldMap:           fieldMapping,
		AllowedFields:      []string{"id", "resource", "resourceId", "name", "value", "createdAt"},
	}).
		WithCreateHook(hooks.CreateHook).
		WithUpdateHook(hooks.UpdateHook).
		WithGetAllHook(hooks.GetAllHook)

	res := &MetricResource{
		processor: proc,
	}

	app.Get("/metrics", res.GetAll)
	app.Get("/metrics/:id", res.GetByID)
	app.Post("/metrics", res.Create)
	app.Put("/metrics/:id", res.Update)
	app.Delete("/metrics/:id", res.Delete)
}

func (r *MetricResource) Create(c *fiber.Ctx) error {
	return r.processor.Create(c)
}

func (r *MetricResource) GetByID(c *fiber.Ctx) error {
	return r.processor.GetByID(c)
}

func (r *MetricResource) GetAll(c *fiber.Ctx) error {
	return r.processor.GetAll(c)
}

func (r *MetricResource) Update(c *fiber.Ctx) error {
	return r.processor.Update(c)
}

func (r *MetricResource) Delete(c *fiber.Ctx) error {
	return r.processor.Delete(c)
}
