package app

import (
	"github.com/aslon1213/go-pos-erp/pkg/configs"
	"github.com/aslon1213/go-pos-erp/pkg/controllers/analytics"
	"github.com/aslon1213/go-pos-erp/pkg/controllers/auth"
	"github.com/aslon1213/go-pos-erp/pkg/controllers/customers"
	"github.com/aslon1213/go-pos-erp/pkg/controllers/customers/bnpl"
	"github.com/aslon1213/go-pos-erp/pkg/controllers/finance"
	journal_handlers "github.com/aslon1213/go-pos-erp/pkg/controllers/journals"
	"github.com/aslon1213/go-pos-erp/pkg/controllers/products"
	"github.com/aslon1213/go-pos-erp/pkg/controllers/sales"
	"github.com/aslon1213/go-pos-erp/pkg/controllers/suppliers"
	"github.com/aslon1213/go-pos-erp/pkg/controllers/transactions"
	"github.com/aslon1213/go-pos-erp/pkg/middleware"
	"github.com/aslon1213/go-pos-erp/pkg/routes"
	"github.com/aslon1213/go-pos-erp/platform/cache"
	pasetoware "github.com/gofiber/contrib/paseto"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Controllers struct {
	Finance      *finance.FinanceController
	Suppliers    *suppliers.SuppliersController
	Transactions *transactions.TransactionsController
	Sales        *sales.SalesTransactionsController
	Journals     *journal_handlers.JournalHandlers
	Operations   *journal_handlers.OperationHandlers
	Products     *products.ProductsController
	Auth         *auth.AuthControllers
	BNPL         *bnpl.BNPLController
	Customers    *customers.CustomersController
	Middlewares  *middleware.Middlewares
	Dashboard    *analytics.DashboardHandler
}

func NewControllers(db *mongo.Database, cache *cache.Cache) *Controllers {
	log.Debug().Msg("Initializing new controllers")
	middleware := middleware.New(db)
	controllers := &Controllers{
		Finance:      finance.New(db),
		Suppliers:    suppliers.New(db),
		Transactions: transactions.New(db),
		Sales:        sales.New(db, cache),
		Journals:     journal_handlers.New(db, cache),
		Operations:   journal_handlers.NewOperationsHandler(db, cache),
		Products:     products.New(db),
		Auth:         auth.New(db),
		Customers:    customers.New(db, cache),
		BNPL:         bnpl.New(db, cache),
		Middlewares:  middleware,
		Dashboard:    analytics.New(db),
	}
	log.Debug().Msg("Controllers initialized successfully")
	return controllers
}

func SetupRoutes(app *fiber.App, controllers *Controllers) {
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	app.Group("/api", pasetoware.New(
		pasetoware.Config{
			SymmetricKey: []byte(config.Server.SecretSymmetricKey),
			// TokenPrefix:    "Bearer",
			SuccessHandler: controllers.Middlewares.AuthMiddleware,
		},
	))

	log.Debug().Msg("Setting up routes")
	routes.AuthRoutes(app, controllers.Auth, controllers.Middlewares)
	log.Debug().Msg("Auth routes set up successfully")
	routes.SuppliersRoutes(app, controllers.Suppliers, controllers.Middlewares)
	log.Debug().Msg("Suppliers routes set up successfully")
	routes.TransactionsRoutes(app, controllers.Transactions, controllers.Middlewares)
	log.Debug().Msg("Transactions routes set up successfully")
	routes.FinanceRoutes(app, controllers.Finance, controllers.Middlewares)
	log.Debug().Msg("Finance routes set up successfully")
	routes.SalesRoutes(app, controllers.Sales, controllers.Middlewares)
	log.Debug().Msg("Sales routes set up successfully")
	routes.JournalsRoutes(app, controllers.Journals, controllers.Operations, controllers.Middlewares)
	log.Debug().Msg("Journals routes set up successfully")
	routes.ProductsRoutes(app, controllers.Products, controllers.Middlewares)
	log.Debug().Msg("Products routes set up successfully")
	routes.CustomerRoutes(app, controllers.Customers, controllers.Middlewares)
	log.Debug().Msg("Customer routes set up successfully")
	routes.BNPLRoutes(app, controllers.BNPL, controllers.Middlewares)
	log.Debug().Msg("BNPL routes set up successfully")
	routes.DashboardRoutes(app, controllers.Dashboard, controllers.Middlewares)
	log.Debug().Msg("Dashboard routes set up successfully")
	log.Debug().Msg("All routes set up successfully")
}
