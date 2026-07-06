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

	// Student exam routes (registrations + card download)
	exam := v1.Group("/exam")
	exam.Use(handler.JWTMiddleware(svc, jwtSigner))
	exam.GET("/registrations", h.StudentListRegistrations)
	exam.GET("/registrations/:id", h.StudentGetRegistration)
	exam.GET("/registrations/:id/card", h.StudentGetExamCard)
	exam.POST("/checkin", h.StudentCheckIn)
	exam.POST("/sessions", h.StudentStartSession)
	exam.GET("/sessions/:id", h.StudentReconnectSession)
	exam.PATCH("/sessions/:id/answers", h.StudentSaveAnswers)
	exam.POST("/sessions/:id/submit", h.StudentSubmitSession)
	exam.POST("/sessions/:id/violations", h.StudentLogViolation)
	exam.GET("/sessions/:id/result", h.StudentGetSessionResult)
	exam.GET("/sessions/:id/leaderboard", h.StudentGetSessionLeaderboard)

	// Admin session routes (exam proctoring)
	adminSessions := admin.Group("/sessions")
	adminSessions.Use(handler.RBACMiddleware("sessions:write"))
	adminSessions.POST("/:id/reopen", h.AdminReopenSession)
	adminSessions.POST("/:id/force-submit", h.AdminForceSubmitSession)
	adminSessions.GET("/:id/essays", h.AdminGetSessionEssays)
	adminSessions.POST("/:id/grade", h.AdminGradeEssay)

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
	// Announcement CRUD + send
	adminNotifs.POST("/announcements", h.AdminCreateAnnouncement, handler.RBACMiddleware("notifications:write"))
	adminNotifs.GET("/announcements", h.AdminListAnnouncements, handler.RBACMiddleware("notifications:read"))
	adminNotifs.PATCH("/announcements/:id", h.AdminUpdateAnnouncement, handler.RBACMiddleware("notifications:write"))
	adminNotifs.DELETE("/announcements/:id", h.AdminDeleteAnnouncement, handler.RBACMiddleware("notifications:write"))
	adminNotifs.POST("/announcements/:id/send", h.AdminSendAnnouncement, handler.RBACMiddleware("notifications:write"))

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
	adminStudents.POST("/bulk/presign", h.AdminPresignStudentBulkUpload)
	adminStudents.POST("/bulk", h.AdminBulkImportStudents)
	adminStudents.POST("/bulk/credentials", h.AdminBulkReissueCredentials)

	// Admin results routes (school-scoped exam results)
	adminResults := admin.Group("/results")
	adminResults.Use(handler.RBACMiddleware("results:read"))
	adminResults.GET("", h.AdminListResults)
	adminResults.GET("/export", h.AdminExportResults)
	adminResults.GET("/:session_id", h.AdminGetResultDetail)

	// Admin job routes (JWT-only; any authenticated user may poll their own job)
	adminJobs := admin.Group("/jobs")
	adminJobs.GET("/:id", h.AdminGetJob)

	// Admin exam routes (Tests + Questions)
	adminTests := admin.Group("/tests")
	adminTests.Use(handler.RBACMiddleware("tests:*"))
	adminTests.GET("", h.AdminListTests)
	adminTests.POST("", h.AdminCreateTest)
	adminTests.GET("/:id", h.AdminGetTest)
	adminTests.PATCH("/:id", h.AdminUpdateTest)
	adminTests.DELETE("/:id", h.AdminDeleteTest)
	adminTests.GET("/:id/questions", h.AdminListQuestions)
	adminTests.POST("/:id/questions", h.AdminCreateQuestion)

	adminQuestions := admin.Group("/questions")
	adminQuestions.Use(handler.RBACMiddleware("questions:*"))
	adminQuestions.PATCH("/:id", h.AdminUpdateQuestion)
	adminQuestions.DELETE("/:id", h.AdminDeleteQuestion)

	// Admin exam package routes (Exam + linked Product)
	adminExams := admin.Group("/exams")
	adminExams.Use(handler.RBACMiddleware("products(exam):write"))
	adminExams.GET("", h.AdminListExams)
	adminExams.POST("", h.AdminCreateExam)
	adminExams.GET("/:id", h.AdminGetExam)
	adminExams.PATCH("/:id", h.AdminUpdateExam)
	adminExams.PUT("/:id/tests", h.AdminReplaceExamTests)
	adminExams.PATCH("/:id/price", h.AdminUpdateExamPrice)
	adminExams.POST("/:id/publish", h.AdminPublishExam)
	adminExams.GET("/:id/grading", h.AdminListGradingSessions)
	adminExams.GET("/:id/leaderboard", h.AdminGetExamLeaderboard)
	adminExams.GET("/:id/analytics", h.AdminGetExamAnalytics)
	adminExams.GET("/:id/certificate-preview", h.AdminGetExamCertificatePreview)

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
