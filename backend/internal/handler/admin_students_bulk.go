package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// AdminPresignStudentBulkUpload issues a presigned PUT URL to the private
// bucket for a student-bulk CSV upload.
func (h *Handler) AdminPresignStudentBulkUpload(c echo.Context) error {
	claims := ClaimsFromContext(c)
	schoolID, err := h.resolveSchoolScope(c, claims)
	if scopeHandled(err) {
		return nil
	}
	if err != nil {
		return err
	}

	filename := c.QueryParam("filename")
	contentType := c.QueryParam("content_type")
	if filename == "" {
		return badRequest(c, "filename is required")
	}

	resp, err := h.svc.GeneratePresignedPrivateUploadURL(c.Request().Context(), schoolID, filename, contentType)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// AdminBulkImportStudents enqueues an async student_bulk job from an
// already-uploaded CSV.
func (h *Handler) AdminBulkImportStudents(c echo.Context) error {
	claims := ClaimsFromContext(c)
	schoolID, err := h.resolveSchoolScope(c, claims)
	if scopeHandled(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var req struct {
		FileKey string `json:"file_key"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.FileKey == "" {
		return badRequest(c, "file_key is required")
	}

	jobID, err := h.svc.EnqueueStudentBulkJob(c.Request().Context(), schoolID, claims.Sub, req.FileKey)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusAccepted, map[string]string{"job_id": jobID})
}

// AdminBulkReissueCredentials reissues credentials for a batch of students
// (or the whole school) and returns the per-row report as a CSV attachment.
func (h *Handler) AdminBulkReissueCredentials(c echo.Context) error {
	claims := ClaimsFromContext(c)
	schoolID, err := h.resolveSchoolScope(c, claims)
	if scopeHandled(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var req struct {
		StudentIDs []string `json:"student_ids"`
		All        bool     `json:"all"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.All == (len(req.StudentIDs) > 0) {
		return badRequest(c, "specify either student_ids or all, not both/neither")
	}

	csvBytes, err := h.svc.ReissueStudentCredentialsBulk(c.Request().Context(), schoolID, req.StudentIDs, req.All)
	if err != nil {
		return mapServiceError(c, err)
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="credentials.csv"`)
	return c.Blob(http.StatusOK, "text/csv", csvBytes)
}
