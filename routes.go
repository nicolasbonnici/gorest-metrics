package metrics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/database"
)

func RegisterRoutes(app *fiber.App, db database.Database, config *Config) {
	RegisterMetricRoutes(app, db, config)
}
