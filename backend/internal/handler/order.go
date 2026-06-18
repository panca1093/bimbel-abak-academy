package handler

import (
	"net/http"

	"akademi-bimbel/internal/infra"
	"akademi-bimbel/internal/service"
	"github.com/labstack/echo/v4"
)

func (h *Handler) MintCart(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}

	order, isNew, err := h.svc.MintCart(c.Request().Context(), claims.Sub)
	if err != nil {
		return mapServiceError(c, err)
	}

	status := http.StatusOK
	if isNew {
		status = http.StatusCreated
	}
	return c.JSON(status, order)
}

func (h *Handler) GetOrders(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}

	cursor := c.QueryParam("cursor")
	limit := 20

	orders, nextCursor, err := h.svc.ListStudentOrders(c.Request().Context(), claims.Sub, cursor, limit)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        orders,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) GetOrder(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}

	orderID := c.Param("id")
	order, err := h.svc.GetStudentOrder(c.Request().Context(), claims.Sub, orderID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, order)
}

func (h *Handler) AddItem(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}

	orderID := c.Param("id")

	var req struct {
		ProductID string `json:"product_id"`
		Qty       int    `json:"qty"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.ProductID == "" || req.Qty <= 0 {
		return badRequest(c, "product_id and qty are required")
	}

	err := h.svc.AddItem(c.Request().Context(), claims.Sub, orderID, req.ProductID, req.Qty)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "item added",
	})
}

func (h *Handler) RemoveItem(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}

	orderID := c.Param("id")
	itemID := c.Param("itemId")

	err := h.svc.RemoveItem(c.Request().Context(), claims.Sub, orderID, itemID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchCart(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}

	orderID := c.Param("id")

	var req struct {
		ShippingAddress []byte `json:"shipping_address"`
		Courier         string `json:"courier"`
		PromoCode       *string `json:"promo_code"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	patch := service.CartPatch{
		ShippingAddress: req.ShippingAddress,
		Courier:         req.Courier,
		PromoCode:       req.PromoCode,
	}

	err := h.svc.PatchCart(c.Request().Context(), claims.Sub, orderID, patch)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "order updated",
	})
}

func (h *Handler) Checkout(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}

	orderID := c.Param("id")
	key := c.Request().Header.Get("Idempotency-Key")
	if key == "" {
		return badRequest(c, "Idempotency-Key header is required")
	}

	result, err := h.svc.Checkout(c.Request().Context(), claims.Sub, orderID, key)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"gateway_ref":        result.GatewayRef,
		"snap_token":         result.SnapToken,
		"payment_expires_at": result.PaymentExpiresAt,
	})
}

func (h *Handler) RetryPayment(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}

	orderID := c.Param("id")
	key := c.Request().Header.Get("Idempotency-Key")
	if key == "" {
		return badRequest(c, "Idempotency-Key header is required")
	}

	result, err := h.svc.RetryPayment(c.Request().Context(), claims.Sub, orderID, key)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"gateway_ref":        result.GatewayRef,
		"snap_token":         result.SnapToken,
		"payment_expires_at": result.PaymentExpiresAt,
	})
}

func (h *Handler) GetShipping(c echo.Context) error {
	var req struct {
		DestinationZip string `json:"destination_zip"`
		WeightGrams    int    `json:"weight_grams"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.DestinationZip == "" || req.WeightGrams <= 0 {
		return badRequest(c, "destination_zip and weight_grams are required")
	}

	rates, err := h.svc.GetShippingRates(c.Request().Context(), service.ShippingQuoteRequest{
		DestinationZip: req.DestinationZip,
		WeightGrams:    req.WeightGrams,
	})
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"rates": rates,
	})
}

func (h *Handler) ValidatePromo(c echo.Context) error {
	var req struct {
		Code     string  `json:"code"`
		Subtotal float64 `json:"subtotal"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.Code == "" || req.Subtotal <= 0 {
		return badRequest(c, "code and subtotal are required")
	}

	validation, err := h.svc.ValidatePromo(c.Request().Context(), req.Code, req.Subtotal)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":           validation.Code,
		"discount":       validation.Discount,
		"final_total":    validation.Total,
	})
}
