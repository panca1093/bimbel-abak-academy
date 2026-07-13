package handler_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/infra"
	"akademi-bimbel/internal/server"
	"akademi-bimbel/internal/service"
)

func TestExamSessionRoutes_Registered(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := service.NewForTest(rdb)
	h := handler.New(svc)
	signer := infra.NewJWTSigner("test-secret", time.Minute)
	e := echo.New()
	e.HideBanner = true
	server.RegisterRoutesForTest(e, h, svc, signer)

	routes := e.Routes()
	pathMap := make(map[string]string)
	for _, r := range routes {
		pathMap[r.Method+":"+r.Path] = r.Name
	}

	expected := []struct{ method, path string }{
		{"POST", "/api/v1/exam/checkin"},
		{"POST", "/api/v1/exam/sessions"},
		{"GET", "/api/v1/exam/sessions/:id"},
		{"PATCH", "/api/v1/exam/sessions/:id/answers"},
		{"POST", "/api/v1/exam/sessions/:id/submit"},
		{"POST", "/api/v1/exam/sessions/:id/violations"},
		{"POST", "/api/v1/exam/sessions/:id/sections/:testId/advance"},
		{"POST", "/api/v1/admin/sessions/:id/reopen"},
		{"POST", "/api/v1/admin/sessions/:id/force-submit"},
		{"GET", "/api/v1/admin/topics"},
		{"POST", "/api/v1/admin/topics"},
		{"PATCH", "/api/v1/admin/topics/:id"},
		{"DELETE", "/api/v1/admin/topics/:id"},
		{"POST", "/api/v1/admin/tests/:id/questions/attach"},
		{"DELETE", "/api/v1/admin/tests/:id/questions/:questionId"},
		{"PUT", "/api/v1/admin/tests/:id/questions/order"},
	}

	for _, exp := range expected {
		key := exp.method + ":" + exp.path
		if _, ok := pathMap[key]; !ok {
			t.Errorf("missing route: %s %s", exp.method, exp.path)
		}
	}
}
