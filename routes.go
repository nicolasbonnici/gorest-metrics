package metrics

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nicolasbonnici/gorest/database"
)

func RegisterRoutes(router fiber.Router, db database.Database, config *Config) {
	RegisterMetricRoutes(router, db, config)
}
