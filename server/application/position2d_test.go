package application

import (
	"testing"
)

func TestPosition2D_EncodeAndParse(t *testing.T) {
	original := &Position2D{X: 1.5, Y: -2.5}

	encoded := original.Encode()
	if len(encoded) != Position2DSize {
		t.Fatalf("encoded size = %d, want %d", len(encoded), Position2DSize)
	}

	decoded, err := ParsePosition2D(encoded)
	if err != nil {
		t.Fatalf("ParsePosition2D failed: %v", err)
	}

	if decoded.X != original.X || decoded.Y != original.Y {
		t.Errorf("decoded = %+v, want %+v", decoded, original)
	}
}

func TestPosition2D_ParseInvalidData(t *testing.T) {
	// 短すぎるデータ
	_, err := ParsePosition2D([]byte{0x01, 0x02, 0x03})
	if err != ErrInvalidPosition2DData {
		t.Errorf("expected ErrInvalidPosition2DData, got %v", err)
	}
}

func TestPosition2D_Zero(t *testing.T) {
	pos := &Position2D{X: 0, Y: 0}
	encoded := pos.Encode()
	decoded, err := ParsePosition2D(encoded)
	if err != nil {
		t.Fatalf("ParsePosition2D failed: %v", err)
	}

	if decoded.X != 0 || decoded.Y != 0 {
		t.Errorf("decoded = %+v, want {0, 0}", decoded)
	}
}
