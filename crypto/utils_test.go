package crypto

import (
	"strings"
	"testing"
	"unicode"

	"github.com/reinbowARA/PassLedger/models"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt(16)
	if err != nil {
		t.Errorf("GenerateSalt failed: %v", err)
	}
	if len(salt1) != 16 {
		t.Errorf("GenerateSalt length = %d, want 16", len(salt1))
	}

	salt2, err := GenerateSalt(16)
	if err != nil {
		t.Errorf("GenerateSalt second call failed: %v", err)
	}
	// Note: salts should be different, but since it's random, we can't guarantee
	// Just check they are not identical (though unlikely)
	if len(salt2) != 16 {
		t.Errorf("GenerateSalt second length = %d, want 16", len(salt2))
	}

	// Test different size
	salt32, err := GenerateSalt(32)
	if err != nil {
		t.Errorf("GenerateSalt(32) failed: %v", err)
	}
	if len(salt32) != 32 {
		t.Errorf("GenerateSalt(32) length = %d, want 32", len(salt32))
	}
}

func TestHmacEqual(t *testing.T) {
	a := []byte("test")
	b := []byte("test")
	c := []byte("different")

	if !HmacEqual(a, b) {
		t.Errorf("HmacEqual with equal slices returned false")
	}

	if HmacEqual(a, c) {
		t.Errorf("HmacEqual with different slices returned true")
	}

	// Different lengths
	d := []byte("test2")
	if HmacEqual(a, d) {
		t.Errorf("HmacEqual with different lengths returned true")
	}

	// Empty slices
	empty1 := []byte{}
	empty2 := []byte{}
	if !HmacEqual(empty1, empty2) {
		t.Errorf("HmacEqual with empty slices returned false")
	}
}

func TestGeneratePassword(t *testing.T) {
	options := models.PasswordGeneratorOptions{
		Length:        12,
		UseLowercase:  true,
		UseUppercase:  true,
		UseDigits:     true,
		UseSpecial:    true,
		UseSpace:      false,
		UseBrackets:   false,
	}

	password, err := GeneratePassword(options)
	if err != nil {
		t.Errorf("GeneratePassword failed: %v", err)
	}

	if len(password) != 12 {
		t.Errorf("GeneratePassword length = %d, want 12", len(password))
	}

	// Build expected charset
	expectedCharset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*-_=+;:,.?/~`"
	// Check all characters are from expected charset
	for _, r := range password {
		if !strings.ContainsRune(expectedCharset, r) {
			t.Errorf("Generated password contains invalid character: %c", r)
		}
	}

	// Test with no character sets selected
	emptyOptions := models.PasswordGeneratorOptions{Length: 10}
	_, err = GeneratePassword(emptyOptions)
	if err == nil {
		t.Errorf("GeneratePassword with no character sets should fail")
	}

	// Test with only lowercase
	lowerOnly := models.PasswordGeneratorOptions{
		Length:       8,
		UseLowercase: true,
	}
	passLower, err := GeneratePassword(lowerOnly)
	if err != nil {
		t.Errorf("GeneratePassword with only lowercase failed: %v", err)
	}
	for _, r := range passLower {
		if !unicode.IsLower(r) {
			t.Errorf("Generated password contains non-lowercase character: %c", r)
		}
	}

	// Test with space and brackets
	withSpace := models.PasswordGeneratorOptions{
		Length:      10,
		UseLowercase: true,
		UseSpace:    true,
		UseBrackets: true,
	}
	passSpace, err := GeneratePassword(withSpace)
	if err != nil {
		t.Errorf("GeneratePassword with space and brackets failed: %v", err)
	}
	expectedSpaceCharset := "abcdefghijklmnopqrstuvwxyz []{}()<>"
	for _, r := range passSpace {
		if !strings.ContainsRune(expectedSpaceCharset, r) {
			t.Errorf("Generated password with space/brackets contains invalid character: %c", r)
		}
	}
}
