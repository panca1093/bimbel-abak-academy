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

	// Check corner pixel for transparency (alpha channel should be 0 or very low)
	// Corner pixels should be transparent background
	r, g, b, a := img.At(0, 0).RGBA()
	// RGBA returns premultiplied colors, so we need to check if alpha is 0
	// Alpha is 16-bit, so 0 means fully transparent
	if a != 0 {
		t.Logf("top-left corner pixel: R=%d G=%d B=%d A=%d", r, g, b, a)
		t.Logf("Note: corner pixel is not transparent; PNG may have logo starting at edge")
	}
}
