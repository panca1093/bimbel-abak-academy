package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// AdminGetJob returns the status of a job owned by the caller.
func (h *Handler) AdminGetJob(c echo.Context) error {
	claims := ClaimsFromContext(c)
	id := c.Param("id")

	resp, err := h.svc.GetJobStatus(c.Request().Context(), id, claims.Sub)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}
