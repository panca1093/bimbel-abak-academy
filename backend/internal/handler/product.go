package handler

import (
	"net/http"

	"akademi-bimbel/internal/infra"
	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
	"github.com/labstack/echo/v4"
)

func (h *Handler) ListProducts(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	filter := repository.ProductFilter{
		Type:   c.QueryParam("type"),
		Status: c.QueryParam("status"),
	}

	products, nextCursor, err := h.svc.ListProducts(c.Request().Context(), filter, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        products,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) GetProduct(c echo.Context) error {
	id := c.Param("id")
	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	product, err := h.svc.GetProduct(c.Request().Context(), id, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, product)
}

func (h *Handler) AdminListProducts(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	filter := repository.ProductFilter{
		Type:   c.QueryParam("type"),
		Status: c.QueryParam("status"),
	}

	products, nextCursor, err := h.svc.ListProducts(c.Request().Context(), filter, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        products,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) AdminCreateProduct(c echo.Context) error {
	var req struct {
		Type        string   `json:"type"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       int64    `json:"price"`
		Stock       int      `json:"stock"`
		CourseIDs   []string `json:"course_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.Type == "" || req.Name == "" {
		return badRequest(c, "type and name are required")
	}

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	p := model.Product{
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Status:      "draft",
	}

	var product model.Product
	var err error
	if req.Type == "course" || len(req.CourseIDs) > 0 {
		product, err = h.svc.CreateProductWithCourses(c.Request().Context(), p, req.CourseIDs, role)
	} else {
		product, err = h.svc.CreateProduct(c.Request().Context(), p, role)
	}
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, product)
}

func (h *Handler) AdminGetProduct(c echo.Context) error {
	id := c.Param("id")
	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	product, err := h.svc.GetProduct(c.Request().Context(), id, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, product)
}

func (h *Handler) AdminUpdateProduct(c echo.Context) error {
	id := c.Param("id")
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       int64    `json:"price"`
		Stock       int      `json:"stock"`
		Status      string   `json:"status"` // published ↔ hidden visibility flip only
		CourseIDs   []string `json:"course_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	p := model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Status:      req.Status,
	}

	var product model.Product
	var err error
	if req.CourseIDs != nil {
		product, err = h.svc.UpdateProductWithCourses(c.Request().Context(), id, p, req.CourseIDs, role)
	} else {
		product, err = h.svc.UpdateProduct(c.Request().Context(), id, p, role)
	}
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, product)
}

func (h *Handler) AdminPublishProduct(c echo.Context) error {
	id := c.Param("id")
	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.PublishProduct(c.Request().Context(), id, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "product published",
	})
}

func (h *Handler) AdminDeleteProduct(c echo.Context) error {
	id := c.Param("id")
	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.DeleteProduct(c.Request().Context(), id, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
