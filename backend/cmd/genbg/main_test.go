package main

import (
	"bytes"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func renderTemplate(t *testing.T, name string) []byte {
	t.Helper()
	draw, ok := templateDrawers[name]
	if !ok {
		t.Fatalf("no drawer registered for template %q", name)
	}
	pdf := newPage()
	draw(pdf)
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		t.Fatalf("render %s: %v", name, err)
	}
	return buf.Bytes()
}

// TestGenerateDeterministic guards `make gen-cert-backgrounds` byte-comparability:
// gofpdf embeds wall-clock CreationDate/ModDate unless the generator pins them.
func TestGenerateDeterministic(t *testing.T) {
	for _, name := range templateNames {
		name := name
		t.Run(name, func(t *testing.T) {
			first := renderTemplate(t, name)
			second := renderTemplate(t, name)
			if !bytes.Equal(first, second) {
				t.Fatalf("cert_bg_%s.pdf is not byte-identical across two renders", name)
			}
		})
	}
}

// TestRasterDimensions rasterizes each generated PDF exactly as the Makefile target
// does and asserts the A4-landscape-at-150DPI pixel size from spec OQ3.
func TestRasterDimensions(t *testing.T) {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skip("pdftoppm not installed")
	}
	dir := t.TempDir()
	for _, name := range templateNames {
		pdfBytes := renderTemplate(t, name)
		pdfPath := filepath.Join(dir, name+".pdf")
		if err := os.WriteFile(pdfPath, pdfBytes, 0o644); err != nil {
			t.Fatal(err)
		}
		prefix := filepath.Join(dir, name)
		cmd := exec.Command("pdftoppm", "-r", "150", "-png", "-singlefile", pdfPath, prefix)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("pdftoppm %s: %v\n%s", name, err, out)
		}
		pngBytes, err := os.ReadFile(prefix + ".png")
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(pngBytes))
		if err != nil {
			t.Fatalf("decode %s.png: %v", name, err)
		}
		if cfg.Width != 1754 || cfg.Height != 1240 {
			t.Errorf("%s: got %dx%d px, want 1754x1240", name, cfg.Width, cfg.Height)
		}
	}
}

// TestCommittedBackgroundPNGs checks the actual assets committed by
// `make gen-cert-backgrounds` satisfy NFR-2's size budget and OQ3's dimensions.
func TestCommittedBackgroundPNGs(t *testing.T) {
	for _, name := range templateNames {
		path := filepath.Join("..", "..", "internal", "service", "assets", "cert_bg_"+name+".png")
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("missing committed PNG %s (run `make gen-cert-backgrounds`): %v", path, err)
		}
		const budget = 800 * 1024
		if info.Size() > budget {
			t.Errorf("%s: %d bytes exceeds %d byte budget", path, info.Size(), budget)
		}

		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		cfg, err := png.DecodeConfig(f)
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		if cfg.Width != 1754 || cfg.Height != 1240 {
			t.Errorf("%s: got %dx%d px, want 1754x1240", path, cfg.Width, cfg.Height)
		}
	}
}
