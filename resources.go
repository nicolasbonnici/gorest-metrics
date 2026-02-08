package metrics

import (
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/filter"
	"github.com/nicolasbonnici/gorest/pagination"
	"github.com/nicolasbonnici/gorest/response"
)

const MaxFilterValuesPerField = 50

type MetricResource struct {
	DB                 database.Database
	CRUD               *crud.CRUD[Metric]
	Config             *Config
	PaginationLimit    int
	PaginationMaxLimit int
}

func RegisterMetricRoutes(app *fiber.App, db database.Database, config *Config) {
	res := &MetricResource{
		DB:                 db,
		CRUD:               crud.New[Metric](db),
		Config:             config,
		PaginationLimit:    config.PaginationLimit,
		PaginationMaxLimit: config.MaxPaginationLimit,
	}

	app.Get("/metrics", res.List)
	app.Get("/metrics/:id", res.Get)
	app.Post("/metrics", res.Create)
	app.Put("/metrics/:id", res.Update)
	app.Delete("/metrics/:id", res.Delete)
}

func (r *MetricResource) validateResourceFilter(filters *filter.FilterSet) error {
	for _, f := range filters.Filters {
		if f.Field == "resource" {
			for _, val := range f.Values {
				if !r.Config.IsAllowedType(val) {
					return fmt.Errorf("invalid resource type '%s' (allowed: %v)", val, r.Config.AllowedTypes)
				}
			}
		}
	}
	return nil
}

func (r *MetricResource) validateFilterLimits(filters *filter.FilterSet) error {
	for _, f := range filters.Filters {
		if len(f.Values) > MaxFilterValuesPerField {
			return fmt.Errorf("too many filter values for field '%s' (max: %d, got: %d)",
				f.Field, MaxFilterValuesPerField, len(f.Values))
		}
	}
	return nil
}

func (r *MetricResource) List(c *fiber.Ctx) error {
	limit := pagination.ParseIntQuery(c, "limit", r.PaginationLimit, r.PaginationMaxLimit)
	page := pagination.ParseIntQuery(c, "page", 1, 10000)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	includeCount := c.Query("count", "true") != "false"

	fieldMapping := map[string]string{
		"id":         "id",
		"resource":   "resource",
		"resourceId": "resource_id",
		"key":        "key",
		"value":      "value",
		"createdAt":  "created_at",
	}

	queryParams := make(url.Values)
	for key, value := range c.Context().QueryArgs().All() {
		queryParams.Add(string(key), string(value))
	}

	filters := filter.NewFilterSetWithMapping(fieldMapping, r.DB.Dialect())
	if err := filters.ParseFromQuery(queryParams); err != nil {
		return pagination.SendPaginatedError(c, 400, err.Error())
	}

	if err := r.validateResourceFilter(filters); err != nil {
		return pagination.SendPaginatedError(c, 400, err.Error())
	}

	if err := r.validateFilterLimits(filters); err != nil {
		return pagination.SendPaginatedError(c, 400, err.Error())
	}

	ordering := filter.NewOrderSetWithMapping(fieldMapping)
	if err := ordering.ParseFromQuery(queryParams); err != nil {
		return pagination.SendPaginatedError(c, 400, err.Error())
	}

	filterOrderClauses := ordering.OrderClauses()
	orderByClauses := make([]crud.OrderByClause, len(filterOrderClauses))
	for i, oc := range filterOrderClauses {
		orderByClauses[i] = crud.OrderByClause{
			Column:    oc.Column,
			Direction: oc.Direction,
		}
	}

	result, err := r.CRUD.GetAllPaginated(c.Context(), crud.PaginationOptions{
		Limit:        limit,
		Offset:       offset,
		IncludeCount: includeCount,
		Conditions:   filters.Conditions(),
		OrderBy:      orderByClauses,
	})
	if err != nil {
		return pagination.SendPaginatedError(c, 500, err.Error())
	}

	return pagination.SendHydraCollection(c, result.Items, result.Total, limit, page, r.PaginationLimit)
}

func (r *MetricResource) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := r.CRUD.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	return response.SendFormatted(c, 200, item)
}

func (r *MetricResource) Create(c *fiber.Ctx) error {
	var req CreateMetricRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := req.Validate(r.Config); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var item Metric
	item.Id = uuid.New().String()
	item.Resource = req.Resource
	item.ResourceId = req.ResourceId
	item.Key = req.Key
	item.Value = req.Value

	ctx := c.Context()
	if err := r.CRUD.Create(ctx, item); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	created, err := r.CRUD.GetByID(ctx, item.Id)
	if err != nil {
		return response.SendFormatted(c, 201, item)
	}

	return response.SendFormatted(c, 201, created)
}

func (r *MetricResource) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateMetricRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := req.Validate(r.Config); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	existing, err := r.CRUD.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	existing.Value = req.Value

	if err := r.CRUD.Update(c.Context(), id, *existing); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return response.SendFormatted(c, 200, existing)
}

func (r *MetricResource) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := r.CRUD.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}
