package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
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
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	_ "github.com/sizzlorox/go-service-boilerplate/internal/docs"

	"github.com/sizzlorox/go-service-boilerplate/internal/datastore"
	"github.com/sizzlorox/go-service-boilerplate/internal/v1/router"
)

type Config struct {
	SERVICE_ENV  string
	SERVICE_NAME string
	SERVICE_PORT int
	DB_URI       string
	DB_PWD       string
	LOGGING      bool
	CACHE        bool
	PREFORK      bool
}

var config Config

// @title Go Service Boilerplate
// @version 1.0
// @description This is a go service boilerplate
// @termsOfService http://swagger.io/terms/
// @host localhost:8080
// @BasePath /
func main() {
	ENV := os.Getenv("SERVICE_ENV")

	// Load dotenv path
	var dotEnvPath string
	if ENV == "production" {
		dotEnvPath = "../../config/.env"
	} else {
		dotEnvPath = "../../config/.env-dev"
	}

	// Load Config
	err := godotenv.Load(dotEnvPath)
	if err != nil {
		log.Panic(err)
	}

	// Parse Environment Variables
	port, err := strconv.Atoi(os.Getenv("SERVICE_PORT"))
	if err != nil {
		log.Panic(err)
	}
	logEnabled, err := strconv.ParseBool(os.Getenv("LOGGING"))
	if err != nil {
		log.Panic(err)
	}
	cachEnabled, err := strconv.ParseBool(os.Getenv("CACHE"))
	if err != nil {
		log.Panic(err)
	}
	preforkEnabled, err := strconv.ParseBool(os.Getenv("PREFORK"))
	if err != nil {
		log.Panic(err)
	}

	config = Config{
		SERVICE_ENV:  os.Getenv("SERVICE_ENV"),
		SERVICE_NAME: os.Getenv("SERVICE_NAME"),
		SERVICE_PORT: port,
		DB_URI:       os.Getenv("DB_URI"),
		DB_PWD:       os.Getenv("DB_PWD"),
		LOGGING:      logEnabled,
		CACHE:        cachEnabled,
		PREFORK:      preforkEnabled,
	}

	if fiber.IsChild() {
		log.Infof("[%d] Child", os.Getppid())
	} else {
		log.Infof("[%d] Master", os.Getppid())
	}

	// Initialize Datastore
	dsConfig := datastore.Config{
		Uri:          config.DB_URI,
		DatabaseName: config.SERVICE_NAME,
	}
	ds := datastore.NewDatastore(&dsConfig)
	// CHANGE: Update indexes here ????
	ds.EnsureIndexes("models", []string{"email"})

	// Initialize Fiber App
	app := initializeApp()

	// Load Routes
	api := app.Group("/api")
	router.LoadRoutes(api, ds)

	// Load Middlewares
	loadMiddlewares(app)

	// Start Server
	go func() {
		err := app.Listen(fmt.Sprintf(":%d", config.SERVICE_PORT))
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

func initializeApp() *fiber.App {
	// New fiber instance
	app := fiber.New(fiber.Config{
		Prefork:      config.PREFORK,
		ServerHeader: config.SERVICE_NAME,
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

	// Exposes swagger docs in /swagger/index.html
	app.Get("/swagger/*", swagger.Handler)

	return app
}

func loadMiddlewares(app *fiber.App) {
	app.Use(cors.New())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(etag.New())
	app.Use(requestid.New())
	if config.SERVICE_ENV == "productionb" {
		app.Use(pprof.New())
	}
	if config.CACHE {
		app.Use(cache.New(cache.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.Query("refresh") == "true"
			},
			Expiration:   1 * time.Minute,
			CacheControl: true,
		}))

	}
	if config.LOGGING {
		app.Use(logger.New(logger.Config{
			Next:         nil,
			Format:       "[${time}] ${status} - ${latency} ${method} ${path}\n",
			TimeFormat:   "15:04:05",
			TimeZone:     "Local",
			TimeInterval: 500 * time.Millisecond,
			Output:       log.StandardLogger().Out,
		}))
	}
	app.Use(compress.New(compress.Config{
		Next: func(c *fiber.Ctx) bool {
			re := regexp.MustCompile(`swagger`)
			return c.Path() == "/dont_compress" || re.Match([]byte(c.Path()))
		},
		Level: compress.LevelBestSpeed,
	}))
}
