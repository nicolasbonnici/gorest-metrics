package metrics

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/processor"
	"github.com/nicolasbonnici/gorest/response"
)

const MaxFilterValuesPerField = 50

type MetricResource struct {
	processor    processor.Processor[Metric, MetricCreateDTO, MetricUpdateDTO, MetricResponseDTO]
	converter    *MetricConverter
	hooks        *MetricHooks
	writer       *batchWriter
	errorHandler processor.ErrorHandler
}

func RegisterMetricRoutes(router fiber.Router, db database.Database, config *Config, writer *batchWriter) {
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
		processor:    proc,
		converter:    converter,
		hooks:        hooks,
		writer:       writer,
		errorHandler: &processor.DefaultErrorHandler{},
	}

	router.Get("/metrics", res.GetAll)
	router.Get("/metrics/:id", res.GetByID)
	router.Post("/metrics", res.Create)
	router.Put("/metrics/:id", res.Update)
	router.Delete("/metrics/:id", res.Delete)
}

// Create records a metric without blocking on the database: it validates the
// request synchronously (preserving the 400s the sync path returned) and hands
// the row to the async batch writer, then answers 201 from the in-memory model.
func (r *MetricResource) Create(c fiber.Ctx) error {
	var dto MetricCreateDTO
	if err := c.Bind().Body(&dto); err != nil {
		return r.errorHandler.HandleError(c, err, "parse")
	}

	model := r.converter.CreateDTOToModel(dto)

	if err := r.hooks.CreateHook(c, dto, &model); err != nil {
		return r.errorHandler.HandleError(c, err, "hook")
	}

	// The DB used to stamp created_at (TIMESTAMP(0)); mirror that precision so
	// the response body stays shape-compatible with the synchronous path.
	now := time.Now().UTC().Truncate(time.Second)
	model.CreatedAt = &now

	r.writer.enqueue(model)

	dtoOut := r.converter.ModelToResponseDTO(model)
	return response.SendFormatted(c, fiber.StatusCreated, dtoOut)
}

func (r *MetricResource) GetByID(c fiber.Ctx) error {
	return r.processor.GetByID(c)
}

func (r *MetricResource) GetAll(c fiber.Ctx) error {
	return r.processor.GetAll(c)
}

func (r *MetricResource) Update(c fiber.Ctx) error {
	return r.processor.Update(c)
}

func (r *MetricResource) Delete(c fiber.Ctx) error {
	return r.processor.Delete(c)
}
