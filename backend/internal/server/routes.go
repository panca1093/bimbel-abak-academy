package server

import (
	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/infra"
	"akademi-bimbel/internal/service"
	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo, h *handler.Handler, svc *service.Service, jwtSigner *infra.JWTSigner) {
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

	// Admin course CRUD
	adminCourses := admin.Group("/courses")
	adminCourses.GET("", h.AdminListCourses)
	adminCourses.POST("", h.AdminCreateCourse)
	adminCourses.GET("/:id", h.AdminGetCourse)
	adminCourses.PATCH("/:id", h.AdminUpdateCourse)
	adminCourses.DELETE("/:id", h.AdminDeleteCourse)

	// Admin section routes (re-keyed from product to course)
	adminSections := admin.Group("/courses/:id/sections")
	adminSections.GET("", h.AdminListSections)
	adminSections.POST("", h.AdminCreateSection)
	adminSections.PUT("/:sId", h.AdminUpdateSection)
	adminSections.DELETE("/:sId", h.AdminDeleteSection)
	adminSections.PATCH("/reorder", h.AdminReorderSections)

	// Admin lesson routes (re-keyed from product to course)
	adminLessons := admin.Group("/courses/:id/sections/:sId/lessons")
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
	orders.PATCH("/:id/items/:itemId", h.UpdateItemQty)
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

	// Public config (client key is safe to expose)
	v1.GET("/config/payment-client-key", h.GetPaymentClientKey)

	// Public school list
	v1.GET("/schools", h.ListSchools)

	// Upload presign (authenticated)
	uploads := v1.Group("/uploads")
	uploads.Use(handler.JWTMiddleware(svc, jwtSigner))
	uploads.POST("/presign", h.GeneratePresignUploadURL)

	// Student profile routes
	students := v1.Group("/students")
	students.Use(handler.JWTMiddleware(svc, jwtSigner))
	students.GET("/dashboard", h.StudentDashboard)
	students.GET("/profile", h.StudentProfile)
	students.PATCH("/profile", h.StudentUpdateProfile)
	students.PATCH("/photo", h.UpdatePhoto)

	// Student course routes
	studentCourses := v1.Group("/courses")
	studentCourses.Use(handler.JWTMiddleware(svc, jwtSigner))
	studentCourses.GET("", h.StudentListCourses)
	studentCourses.GET("/:id", h.StudentGetCourse)
	studentCourses.POST("/:id/lessons/:lId/complete", h.StudentMarkLessonComplete)
	studentCourses.GET("/:id/progress", h.StudentCourseProgress)

	// Admin order routes
	adminOrders := admin.Group("/orders")
	adminOrders.Use(handler.RBACMiddleware("orders:write"))
	adminOrders.GET("", h.AdminListOrders)
	adminOrders.GET("/:id", h.AdminGetOrder)
	adminOrders.POST("/:id/confirm", h.AdminConfirmOrder)
	adminOrders.POST("/:id/ship", h.AdminShipOrder)
	adminOrders.POST("/:id/complete", h.AdminCompleteOrder)
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

	// Admin school routes
	adminSchools := admin.Group("/schools")
	adminSchools.Use(handler.RBACMiddleware("schools:write"))
	adminSchools.GET("", h.AdminListSchools)
	adminSchools.POST("", h.AdminCreateSchool)
	adminSchools.PUT("/:id", h.AdminUpdateSchool)
	adminSchools.PATCH("/:id", h.AdminChangeSchoolStatus)

	// Admin student routes (row-scoped via JWT schoolID)
	adminStudents := admin.Group("/students")
	adminStudents.Use(handler.RBACMiddleware("students:*"))
	adminStudents.GET("", h.AdminListStudents)
	adminStudents.POST("", h.AdminRegisterStudent)
	adminStudents.PATCH("/:id", h.AdminChangeStudentStatus)
	adminStudents.GET("/:id/credentials", h.AdminGetStudentCredentials)

	// Admin system routes (super_admin only)
	adminSystem := admin.Group("/system")
	adminSystem.Use(handler.RBACMiddleware("system:admin"))
	adminSystem.GET("/accounts", h.AdminListAccounts)
	adminSystem.POST("/accounts", h.AdminCreateAccount)
	adminSystem.PATCH("/accounts/:id/role", h.AdminChangeAccountRole)
	adminSystem.PATCH("/accounts/:id/status", h.AdminChangeAccountStatus)
	adminSystem.POST("/accounts/:id/reset-password", h.AdminResetAccountPassword)
	adminSystem.GET("/audit", h.AdminListAuditLog)
	adminSystem.GET("/config", h.AdminGetSystemConfig)
	adminSystem.PUT("/config", h.AdminUpdateSystemConfig)
}

// RegisterRoutesForTest is the same as registerRoutes but exported for handler tests.
func RegisterRoutesForTest(e *echo.Echo, h *handler.Handler, svc *service.Service, jwtSigner *infra.JWTSigner) {
	registerRoutes(e, h, svc, jwtSigner)
}
