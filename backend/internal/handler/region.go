package handler

import (
	"net/http"

	"akademi-bimbel/internal/model"

	"github.com/labstack/echo/v4"
)

// ListProvinces returns all provinces. Public endpoint, no JWT required.
func (h *Handler) ListProvinces(c echo.Context) error {
	provinces, err := h.svc.ListProvinces(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, provinces)
}

// ListCitiesByProvince returns cities for the given province id.
// Returns an empty JSON array (not a 404) when the province id is unknown.
// Public endpoint, no JWT required.
func (h *Handler) ListCitiesByProvince(c echo.Context) error {
	provinceID := c.Param("id")

	cities, err := h.svc.ListCitiesByProvince(c.Request().Context(), provinceID)
	if err != nil {
		return mapServiceError(c, err)
	}
	if cities == nil {
		cities = []model.City{}
	}
	return c.JSON(http.StatusOK, cities)
}

// ListDistrictsByCity returns districts for the given city id.
// Returns an empty JSON array (not a 404) when the city id is unknown.
// Public endpoint, no JWT required.
func (h *Handler) ListDistrictsByCity(c echo.Context) error {
	cityID := c.Param("id")

	districts, err := h.svc.ListDistrictsByCity(c.Request().Context(), cityID)
	if err != nil {
		return mapServiceError(c, err)
	}
	if districts == nil {
		districts = []model.District{}
	}
	return c.JSON(http.StatusOK, districts)
}
