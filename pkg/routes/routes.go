package routes

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

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/rs/zerolog/log"
)

func DashboardRoutes(router *fiber.App, dashboardController *analytics.DashboardHandler, middleware *middleware.Middlewares) {
	dashboard := router.Group("/dashboard")
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	auth := basicauth.New(basicauth.Config{
		Users: map[string]string{
			config.Server.AdminDocsUsers[0].Username: config.Server.AdminDocsUsers[0].Password,
			config.Server.AdminDocsUsers[1].Username: config.Server.AdminDocsUsers[1].Password,
		},
		Realm: "Restricted",
	})

	// log.Debug().Interface("auth", auth).Msg("Auth middleware initialized")
	dashboard.Get("/journals", auth, dashboardController.ServeDashBoardDays)
	dashboard.Get("/general", auth, dashboardController.ServeDashBoardGeneral)
	dashboard.Get("/comparison", auth, dashboardController.ServeDashBoardComparison)
	dashboard.Get("/", auth, dashboardController.MainPage)
	// dashboard.Get("/branches")
}

func AuthRoutes(router *fiber.App, authController *auth.AuthControllers, middleware *middleware.Middlewares) {
	auth := router.Group("/auth")
	auth.Post("/login", authController.Login)                                // login -- activity logged here if succesfull
	auth.Post("/register", authController.Register)                          // register -- activity logged here if succesfull
	router.Get("/api/auth/me", authController.InfoMe)                        // get user info
	router.Get("/api/activities/recent", authController.GetRecentActivities) // get recent activities
	router.Get("/api/activities/me", authController.GetActivitesOfUser)      // get activities of user
}

func SuppliersRoutes(router *fiber.App, suppliersController *suppliers.SuppliersController, middleware *middleware.Middlewares) {
	api := router.Group("/api")
	api.Get("/suppliers", suppliersController.GetSuppliers)                                         // get all suppliers
	api.Get("/suppliers/:id", suppliersController.GetSupplierByID)                                  // get supplier by id
	api.Post("/suppliers", suppliersController.CreateSupplier)                                      // create supplier -- activity logged here if succesfull                                   // create supplier
	api.Put("/suppliers/:id", suppliersController.UpdateSupplier)                                   // update supplier
	api.Delete("/suppliers/:id", suppliersController.DeleteSupplier)                                // delete supplier
	api.Post("/suppliers/:branch_id/:supplier_id/transactions", suppliersController.NewTransaction) // create transaction
}

func SalesRoutes(router *fiber.App, salesController *sales.SalesTransactionsController, middleware *middleware.Middlewares) {
	api := router.Group("/api")
	api.Post("/sales/transactions/:branch_id", salesController.CreateSalesTransaction)        // create sales transaction -- activity logged here if succesfull
	api.Delete("/sales/transactions/:transaction_id", salesController.DeleteSalesTransaction) // delete sales transaction -- activity logged here if succesfull
	// sales session routes
	api.Post("/sales/session/branch/:branch_id", salesController.OpenSalesSession)          // open sales session
	api.Post("/sales/session/:session_id/product", salesController.AddProductItemToSession) // add product to session
	api.Post("/sales/session/:session_id/close", salesController.CloseSalesSession)         // close sales session
	api.Get("/sales/session/:session_id", salesController.GetSalesSession)                  // get sales session
	api.Delete("/sales/session/:session_id", salesController.DeleteSalesSession)            // delete sales session
	api.Get("/sales/session/branch/:branch_id", salesController.GetSalesSessionsOfBranch)   // get sales of session
	// router.Get("/sales/branch/:branch_id/sessions", salesController.GetSalesOfSession)

}

func ProductsRoutes(router *fiber.App, productsController *products.ProductsController, middleware *middleware.Middlewares) {
	api := router.Group("/api")

	api.Post("/products", productsController.CreateProduct)                        // create product -- activity logged here if succesfull
	api.Put("/products/:id", productsController.EditProduct)                       // edit product -- activity logged here if succesfull
	api.Delete("/products/:id", productsController.DeleteProduct)                  // delete product -- activity logged here if succesfull
	api.Get("/products/:id", productsController.GetProductByID)                    // get product by id
	api.Get("/products", productsController.QueryProducts)                         // query products
	api.Post("/products/:id/income", productsController.NewIncome)                 // create income -- activity logged here if succesfull
	api.Post("/products/transfer", productsController.NewTransfer)                 // create transfer -- activity logged here if succesfull
	api.Post("/products/:id/images", productsController.UploadProductImage)        // upload product image
	api.Delete("/products/:id/images/:key", productsController.DeleteProductImage) // delete product image
	api.Get("/products/:id/images", productsController.GetImagesOfProduct)         // get images of product
	api.Get("/products/images/:key", productsController.GetImage)                  // get image
}

func JournalsRoutes(router *fiber.App, journalsController *journal_handlers.JournalHandlers, operationsController *journal_handlers.OperationHandlers, middleware *middleware.Middlewares) {
	api := router.Group("/api")
	api.Get("/journals/:id", journalsController.GetJournalEntryByID)                                                  // get journal entry by id
	api.Get("/journals/branch/:branch_id", journalsController.QueryJournalEntries)                                    // query journal entries
	api.Post("/journals", journalsController.NewJournalEntry)                                                         // create journal entry -- activity logged here if succesfull
	api.Post("/journals/:id/close", operationsController.ShiftIsOpenMiddleware, journalsController.CloseJournalEntry) // close journal entry -- activity logged here if succesfull
	api.Post("/journals/:id/reopen", journalsController.ReOpenJournalEntry)                                           // reopen journal entry -- activity logged here if succesfull

	// operations
	api.Post("/journals/:id/operations", operationsController.ShiftIsOpenMiddleware, operationsController.NewOperationTransaction)                        // create operation transaction -- activity logged here if succesfull
	api.Put("/journals/:id/operations/:operation_id", operationsController.ShiftIsOpenMiddleware, operationsController.UpdateOperationTransactionByID)    // update operation transaction by id -- activity logged here if succesfull
	api.Delete("/journals/:id/operations/:operation_id", operationsController.ShiftIsOpenMiddleware, operationsController.DeleteOperationTransactionByID) // delete operation transaction by id -- activity logged here if succesfull
	api.Get("/journals/:id/operations/:operation_id", operationsController.GetOperationTransactionByID)                                                   // get operation transaction by id

}

func InternalExpensesRoutes(router *fiber.App, middleware *middleware.Middlewares) {

}

func FinanceRoutes(router *fiber.App, financeController *finance.FinanceController, middleware *middleware.Middlewares) {
	api := router.Group("/api")
	api.Get("/finance/branches", financeController.GetBranches)                            // get all branches
	api.Get("/finance/branch/id/:id", financeController.GetFinanceBranchByBranchID)        // get branch by id
	api.Get("/finance/branch/name/:branch_name", financeController.GetFinanceByBranchName) // get branch by name
	api.Get("/finance/id/:id", financeController.GetFinanceByID)                           // get finance by id
	api.Post("/finance", financeController.NewFinanceOfBranch)                             // create new finance of branch -- activity logged here if succesfull

}

func TransactionsRoutes(router *fiber.App, transactionsController *transactions.TransactionsController, middleware *middleware.Middlewares) {
	api := router.Group("/api")
	api.Get("/transactions/branch/:branch_id", transactionsController.GetTransactionsByQueryParams) // get transactions by query params
	api.Get("/transactions/:id", transactionsController.GetTransactionByID)                         // get transaction by id
	// router.Post("/transactions/:branch_id", transactionsController.Tra)
	api.Put("/transactions/:id", transactionsController.UpdateTransactionByID)            // update transaction by id
	api.Delete("/transactions/:id", transactionsController.DeleteTransactionByID)         // delete transaction by id
	api.Get("/transactions/docs/initiator_type", transactionsController.GetInitiatorType) // get initiator type
	api.Get("/transactions/docs/type", transactionsController.GetTransactionType)         // get transaction type
	api.Get("/transactions/docs/payment_method", transactionsController.GetPaymentMethod) // get payment method
}

func CustomerRoutes(router *fiber.App, customerController *customers.CustomersController, middleware *middleware.Middlewares) {
	api := router.Group("/api")
	api.Get("/customers", customerController.GetCustomers)          // get all customers
	api.Get("/customers/:id", customerController.GetCustomerByID)   // get customer by id
	api.Post("/customers", customerController.CreateCustomer)       // create customer -- activity logged here if succesfull
	api.Put("/customers/:id", customerController.UpdateCustomer)    // update customer
	api.Delete("/customers/:id", customerController.DeleteCustomer) // delete customer
}

func BNPLRoutes(router *fiber.App, bnplController *bnpl.BNPLController, middleware *middleware.Middlewares) {
	api := router.Group("/api")
	api.Post("/bnpl", bnplController.NewBNPL)                                   // create bnpl -- activity logged here if succesfull
	api.Post("/bnpl/:id/credit", bnplController.CreditBNPL)                     // credit bnpl
	api.Delete("/bnpl/:id", bnplController.DeleteBNPL)                          // delete bnpl
	api.Get("/bnpl/:id", bnplController.GetBNPLByID)                            // get bnpl by id
	api.Get("/customers/:customer_id/bnpls", bnplController.GetBNPLSofCustomer) // get bnpls of customer
	api.Get("/branches/:branch_id/bnpls", bnplController.GetBNPLsOfBranch)      // get bnpls of branch
}
