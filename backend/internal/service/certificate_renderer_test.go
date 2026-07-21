package service

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGotenbergRenderer_RenderHTML_PostsMultipartForm(t *testing.T) {
	var (
		gotMethod     string
		gotPath       string
		gotFilePart   []byte
		gotFileName   string
		gotFormFields = map[string]string{}
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "multipart/form-data" {
			t.Fatalf("expected multipart/form-data content-type, got %q (err=%v)", r.Header.Get("Content-Type"), err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("multipart read error: %v", err)
			}
			data, err := io.ReadAll(part)
			if err != nil {
				t.Fatalf("read part error: %v", err)
			}
			if part.FormName() == "files" || part.FileName() != "" {
				gotFilePart = data
				gotFileName = part.FileName()
			} else {
				gotFormFields[part.FormName()] = string(data)
			}
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("%PDF-fake-bytes"))
	}))
	defer srv.Close()

	r := newGotenbergRenderer(srv.URL, srv.Client())
	htmlInput := []byte("<html><body>hello</body></html>")

	pdfBytes, err := r.RenderHTML(context.Background(), htmlInput)
	if err != nil {
		t.Fatalf("RenderHTML returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/forms/chromium/convert/html" {
		t.Errorf("path = %q, want /forms/chromium/convert/html", gotPath)
	}
	if gotFileName != "index.html" {
		t.Errorf("file part name = %q, want index.html", gotFileName)
	}
	if string(gotFilePart) != string(htmlInput) {
		t.Errorf("file part content = %q, want %q", gotFilePart, htmlInput)
	}

	wantFields := map[string]string{
		"printBackground":   "true",
		"preferCssPageSize": "true",
		"marginTop":         "0",
		"marginBottom":      "0",
		"marginLeft":        "0",
		"marginRight":       "0",
	}
	for k, want := range wantFields {
		if got := gotFormFields[k]; got != want {
			t.Errorf("form field %q = %q, want %q", k, got, want)
		}
	}

	if string(pdfBytes) != "%PDF-fake-bytes" {
		t.Errorf("returned bytes = %q, want %q", pdfBytes, "%PDF-fake-bytes")
	}
}

func TestGotenbergRenderer_RenderHTML_NonOKStatusReturnsWrappedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()

	r := newGotenbergRenderer(srv.URL, srv.Client())

	_, err := r.RenderHTML(context.Background(), []byte("<html></html>"))
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestGotenbergRenderer_ImplementsCertificateRenderer(t *testing.T) {
	var _ certificateRenderer = (*gotenbergRenderer)(nil)
}
