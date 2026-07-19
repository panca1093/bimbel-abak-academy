package assets

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"
)

func TestLogoMarkPNG_Decodable(t *testing.T) {
	if len(LogoMarkPNG) == 0 {
		t.Fatal("LogoMarkPNG is empty")
	}

	img, format, err := image.Decode(bytes.NewReader(LogoMarkPNG))
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected format png, got %s", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 512 || bounds.Dy() != 512 {
		t.Fatalf("expected 512x512, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestLogoMarkPNG_TransparentBackground(t *testing.T) {
	if len(LogoMarkPNG) == 0 {
		t.Fatal("LogoMarkPNG is empty")
	}

	img, _, err := image.Decode(bytes.NewReader(LogoMarkPNG))
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}

	// RGBA returns premultiplied colors; alpha is 16-bit, 0 means fully transparent.
	_, _, _, a := img.At(0, 0).RGBA()
	if a != 0 {
		t.Fatalf("expected transparent top-left corner pixel, got alpha=%d", a)
	}
}
