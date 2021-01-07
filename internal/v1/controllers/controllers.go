package controllers

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"gitlab.com/PROD/go-service-boilerplate/internal/v1/models"
	"gitlab.com/PROD/go-service-boilerplate/internal/v1/services"
)

type Controller interface {
	Get(ctx *fiber.Ctx) error
	GetById(ctx *fiber.Ctx) error
	Create(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
}

type controller struct {
	s services.Service
}

/*
* CONSTRUCTOR
 */

func NewController(s services.Service) Controller {
	return &controller{s}
}

/*
* PUBLIC
 */

// Get godoc
// @Summary Gets a model
// @Tags Model
// @Produce json
// @Success 200 {object} models.Response
// @Router / [get]
func (c *controller) Get(ctx *fiber.Ctx) error {
	p := ctx.Query("page")
	l := ctx.Query("limit")
	if len(p) == 0 {
		return fiber.NewError(409, "Page is required")
	}

	res, err := c.s.Get(p, l)
	if err != nil {
		log.Error(err)
	}
	return ctx.Status(res.Status).JSON(res)
}

// GetById godoc
// @Summary Gets a model by ID
// @Tags Model
// @Produce json
// @Success 200 {object} models.Response
// @Router /{id} [get]
func (c *controller) GetById(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	res, err := c.s.GetById(id)
	if err != nil {
		log.Error(err)
		return err
	}
	return ctx.Status(res.Status).JSON(res)
}

// Create godoc
// @Summary Creates a model
// @Tags Model
// @Produce json
// @Success 201 {object} models.Response{data=models.CreateResponse}
// @Failure 409 {object} models.ValidationError
// @Router /create [post]
func (c *controller) Create(ctx *fiber.Ctx) error {
	var m models.Model
	if err := ctx.BodyParser(&m); err != nil {
		log.Error(err)
		return err
	}

	errors := m.ValidateStruct()
	if errors != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors)
	}

	res, err := c.s.Create(&m)
	if err != nil {
		log.Error(err)
		return err
	}
	return ctx.Status(res.Status).JSON(res)
}

// Update godoc
// @Summary Updates a model
// @Tags Model
// @Produce json
// @Success 200 {object} models.Response
// @Router /{id}/update [put]
func (c *controller) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var m models.Model
	if err := ctx.BodyParser(&m); err != nil {
		log.Error(err)
		return err
	}

	if m.IsNil() {
		err := fiber.NewError(fiber.StatusBadRequest, "You require at least one field to update")
		log.Error(err)
		return err
	}

	res, err := c.s.Update(id, &m)
	if err != nil {
		log.Error(err)
		return err
	}
	return ctx.Status(res.Status).JSON(res)
}

// Delete godoc
// @Summary Deletes a model
// @Tags Model
// @Produce json
// @Success 200 {object} models.Response
// @Router /{id} [delete]
func (c *controller) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	res, err := c.s.Delete(id)
	if err != nil {
		log.Error(err)
		return err
	}
	return ctx.Status(res.Status).JSON(res)
}
