package app

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/aslon1213/go-pos-erp/pkg/configs"
	"github.com/aslon1213/go-pos-erp/platform/cache"
	"github.com/aslon1213/go-pos-erp/platform/database"
	"github.com/aslon1213/go-pos-erp/platform/logger"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	_ "github.com/aslon1213/go-pos-erp/docs"
	"github.com/go-playground/validator/v10"

	"go.mongodb.org/mongo-driver/v2/mongo"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// @title Magazin ERP/POS API
// @version 1.0
// @description This is a ERP/POS API for Magazin.
// @contact.name API Support
// @contact.url https://github.com/aslon1213/go-pos-erp
// @contact.email hamidovaslon13@gmail.com
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type App struct {
	Logger *zerolog.Logger
	Cache  *cache.Cache
	DB     *mongo.Client
	Config *configs.Config
	Router *fiber.App
}

var tracer = otel.Tracer("fiber-server")

type (
	ErrorResponse struct {
		Error       bool
		FailedField string
		Tag         string
		Value       interface{}
	}

	XValidator struct {
		validator *validator.Validate
	}

	GlobalErrorHandlerResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
)

var validate = validator.New()

func (v XValidator) Validate(data interface{}) []ErrorResponse {
	validationErrors := []ErrorResponse{}

	errs := validate.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			// In this case data object is actually holding the User struct
			var elem ErrorResponse

			elem.FailedField = err.Field() // Export struct field name
			elem.Tag = err.Tag()           // Export struct tag
			elem.Value = err.Value()       // Export field value
			elem.Error = true

			validationErrors = append(validationErrors, elem)
		}
	}

	return validationErrors
}

func NewFiberApp() *fiber.App {
	config, _ := configs.LoadConfig(".")

	app := fiber.New(
	// fiber.Config{
	// 	ErrorHandler: func(c *fiber.Ctx, err error) error {
	// 		return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
	// 			Success: false,
	// 			Message: err.Error(),
	// 		})
	// 	},
	// },
	)
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Provide a minimal config
	// app.Use(basicauth.New(basicauth.Config{
	// 	Users: map[string]string{
	// 		"john":  "doe",
	// 		"admin": "123456",
	// 	},

	// }))

	app.Use(otelfiber.Middleware())
	app.Use(cors.New())
	app.Use(logger.CustomZerologMiddleware)

	app.Group(
		"/docs",
		basicauth.New(basicauth.Config{
			Users: map[string]string{
				config.Server.AdminDocsUsers[0].Username: config.Server.AdminDocsUsers[0].Password,
				config.Server.AdminDocsUsers[1].Username: config.Server.AdminDocsUsers[1].Password,
			},
			Realm: "Restricted",
		}),
	)

	app.Get("/docs/*", fiberSwagger.WrapHandler)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html")
	})
	return app
}

func New() *App {
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	return &App{
		Logger: logger.SetupLogger(),
		Cache:  cache.New(),
		DB:     database.NewDB(),
		Router: NewFiberApp(),
		Config: config,
	}
}

func initTracer() *sdktrace.TracerProvider {
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tracer")
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("my-service"),
			)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func (a *App) Run() {
	controllers := NewControllers(a.DB.Database(a.Config.DB.Database), a.Cache)
	SetupRoutes(a.Router, controllers)
	a.Router.Listen(a.Config.Server.Port)
}

func (app *App) MigrateDatabase() {

	// migrate db to json and save to s3

}
