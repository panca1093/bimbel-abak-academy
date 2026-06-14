package server

import (
	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/service"
	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo, h *handler.Handler, svc *service.Service, jwtSigner *platform.JWTSigner) {
	v1 := e.Group("/api/v1")
	v1.GET("/health", h.Health)

	auth := v1.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login, handler.LoginRateLimiter())
	auth.POST("/google", h.GoogleLogin)
	auth.POST("/otp/send", h.SendOTP)
	auth.POST("/otp/verify", h.VerifyOTP)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout, handler.JWTMiddleware(svc, jwtSigner))
	auth.POST("/password/forgot", h.ForgotPassword)
	auth.POST("/password/reset", h.ResetPassword)
	auth.PATCH("/password/change", h.ChangePassword, handler.JWTMiddleware(svc, jwtSigner))
	auth.GET("/me", h.Me, handler.JWTMiddleware(svc, jwtSigner))

	// Student product routes
	products := v1.Group("/products")
	products.GET("", h.ListProducts)
	products.GET("/:id", h.GetProduct)

	// Admin product routes
	admin := v1.Group("/admin")
	admin.Use(handler.JWTMiddleware(svc, jwtSigner))

	adminProducts := admin.Group("/products")
	adminProducts.GET("", h.AdminListProducts)
	adminProducts.POST("", h.AdminCreateProduct)
	adminProducts.GET("/:id", h.AdminGetProduct)
	adminProducts.PATCH("/:id", h.AdminUpdateProduct)
	adminProducts.POST("/:id/publish", h.AdminPublishProduct)
	adminProducts.DELETE("/:id", h.AdminDeleteProduct)

	// Admin course routes
	adminCourses := admin.Group("/products/:id/sections")
	adminCourses.GET("", h.AdminListSections)
	adminCourses.POST("", h.AdminCreateSection)
	adminCourses.PUT("/:sId", h.AdminUpdateSection)
	adminCourses.DELETE("/:sId", h.AdminDeleteSection)
	adminCourses.PATCH("/reorder", h.AdminReorderSections)

	// Admin lesson routes
	adminLessons := admin.Group("/products/:id/sections/:sId/lessons")
	adminLessons.POST("", h.AdminCreateLesson)
	adminLessons.PUT("/:lId", h.AdminUpdateLesson)
	adminLessons.DELETE("/:lId", h.AdminDeleteLesson)
	adminLessons.PATCH("/reorder", h.AdminReorderLessons)

	// Student order routes
	orders := v1.Group("/orders")
	orders.Use(handler.JWTMiddleware(svc, jwtSigner))
	orders.POST("", h.MintCart)
	orders.GET("", h.GetOrders)
	orders.GET("/:id", h.GetOrder)
	orders.POST("/:id/items", h.AddItem)
	orders.DELETE("/:id/items/:itemId", h.RemoveItem)
	orders.PATCH("/:id", h.PatchCart)
	orders.POST("/:id/checkout", h.Checkout)
	orders.POST("/:id/retry", h.RetryPayment)

	// Shipping and promo routes
	shipping := v1.Group("/orders/shipping")
	shipping.POST("", h.GetShipping)

	promo := v1.Group("/promo-codes")
	promo.POST("/validate", h.ValidatePromo)

	// Payment webhook route (no auth, uses HMAC signature)
	webhooks := v1.Group("/webhooks")
	webhooks.POST("/payment", h.HandlePaymentWebhook)

	// Admin order routes
	adminOrders := admin.Group("/orders")
	adminOrders.Use(handler.RBACMiddleware("orders:write"))
	adminOrders.GET("", h.AdminListOrders)
	adminOrders.GET("/:id", h.AdminGetOrder)
	adminOrders.POST("/:id/confirm", h.AdminConfirmOrder)
	adminOrders.POST("/:id/ship", h.AdminShipOrder)
	adminOrders.POST("/:id/refund", h.AdminRefundOrder)
	adminOrders.POST("/:id/reconcile", h.AdminReconcileOrder)

	// Admin promo code routes
	adminPromos := admin.Group("/promo-codes")
	adminPromos.Use(handler.RBACMiddleware("promos:write"))
	adminPromos.GET("", h.AdminListPromoCodes)
	adminPromos.POST("", h.AdminCreatePromoCode)
	adminPromos.PUT("/:id", h.AdminUpdatePromoCode)
	adminPromos.DELETE("/:id", h.AdminDeletePromoCode)

	// Admin revenue and notification routes
	adminRevenue := admin.Group("/revenue")
	adminRevenue.Use(handler.RBACMiddleware("revenue:read"))
	adminRevenue.GET("", h.AdminGetRevenue)

	adminNotifs := admin.Group("/notifications")
	adminNotifs.Use(handler.RBACMiddleware("notifications:read"))
	adminNotifs.GET("", h.AdminListNotifications)
	adminNotifs.PATCH("/:id/read", h.AdminMarkNotificationRead)
}

// RegisterRoutesForTest is the same as registerRoutes but exported for handler tests.
func RegisterRoutesForTest(e *echo.Echo, h *handler.Handler, svc *service.Service, jwtSigner *platform.JWTSigner) {
	registerRoutes(e, h, svc, jwtSigner)
}
