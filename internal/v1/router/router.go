package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sizzlorox/go-service-boilerplate/internal/datastore"
	"github.com/sizzlorox/go-service-boilerplate/internal/utils"
	"github.com/sizzlorox/go-service-boilerplate/internal/v1/controllers"
	"github.com/sizzlorox/go-service-boilerplate/internal/v1/services"
)

func LoadRoutes(api fiber.Router, ds datastore.Repository) {
	// Initialize Utils
	u := utils.NewUtils()

	// Initialize Service and Controller
	s := services.NewService(ds, u)
	c := controllers.NewController(s)

	// Register Routes and Handlers
	v1 := api.Group("/v1")
	v1.Get("/", c.Get)
	v1.Get("/:id", c.GetById)
	v1.Put("/create", c.Create)
	v1.Post("/:id/update", c.Update)
	v1.Delete("/:id/delete", c.Delete)
}
