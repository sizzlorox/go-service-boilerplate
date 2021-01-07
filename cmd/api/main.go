package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/helmet/v2"
	log "github.com/sirupsen/logrus"
	_ "github.com/sizzlorox/go-service-boilerplate/internal/docs"
	"github.com/sizzlorox/go-service-boilerplate/internal/utils"

	"github.com/sizzlorox/go-service-boilerplate/internal/datastore"
	"github.com/sizzlorox/go-service-boilerplate/internal/v1/controllers"
	"github.com/sizzlorox/go-service-boilerplate/internal/v1/services"
)

// @title Go Service Boilerplate
// @version 1.0
// @description This is a go service boilerplate
// @termsOfService http://swagger.io/terms/
// @host localhost:8080
// @BasePath /
func main() {
	// TODO: check for prefork
	if fiber.IsChild() {
		log.Infof("[%d] Child", os.Getppid())
	} else {
		log.Infof("[%d] Master", os.Getppid())
	}

	// TODO: Move prefork to config
	app := fiber.New(fiber.Config{
		Prefork:      true,
		ServerHeader: "Service Name",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			err = ctx.Status(code).JSON(err)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}

			return nil
		},
	})

	// Load Middlewares
	loadMiddlewares(app)

	// Initialize Datastore
	dsConfig := datastore.Config{
		Uri:          "mongodb://gs_user:gs_pwd@localhost:27017/gs_service",
		DatabaseName: "gs_service",
	}
	ds := datastore.NewDatastore(&dsConfig)
	ds.EnsureIndexes("models", []string{"email"})

	// Initialize Utils
	u := utils.NewUtils()

	// Initialize Service and Controller
	s := services.NewService(ds, u)
	c := controllers.NewController(s)

	// Register Routes and Handlers
	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.Get("/", c.Get)
	v1.Get("/:id", c.GetById)
	v1.Post("/create", c.Create)
	v1.Put("/:id/update", c.Update)
	v1.Delete("/delete/:id", c.Delete)

	// Exposes swagger docs in /swagger/index.html
	app.Get("/swagger/*", swagger.Handler)

	// Start Server
	go func() {
		err := app.Listen(":3500")
		if err != nil {
			log.Panic(err)
		}
	}()

	// Graceful Shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block main thread until interrupt is received
	_ = <-sig

	log.Info("Gracefully shutting down...")
	_ = app.Shutdown()

	log.Info("Cleaning up modules...")
	ds.Close()
}

func loadMiddlewares(app *fiber.App) {
	app.Use(cors.New())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(etag.New())
	app.Use(requestid.New())
	// TODO: Check for prod env, this can only activate in a dev env
	app.Use(pprof.New())
	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("refresh") == "true"
		},
		Expiration:   1 * time.Minute,
		CacheControl: true,
	}))
	app.Use(logger.New(logger.Config{
		Next:         nil,
		Format:       "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat:   "15:04:05",
		TimeZone:     "Local",
		TimeInterval: 500 * time.Millisecond,
		Output:       log.StandardLogger().Out,
	}))
	app.Use(compress.New(compress.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/dont_compress"
		},
		Level: compress.LevelBestSpeed,
	}))
}
