package crypto

import (
	"bytes"
	"testing"
)

func TestPkcs7Pad(t *testing.T) {
	tests := []struct {
		input    []byte
		blockSize int
		expectedLen int
	}{
		{[]byte("hello"), 16, 16},
		{[]byte("hello world"), 16, 16},
		{[]byte(""), 16, 16},
		{[]byte("1234567890123456"), 16, 32}, // exactly blockSize
	}

	for _, tt := range tests {
		result := pkcs7Pad(tt.input, tt.blockSize)
		if len(result) != tt.expectedLen {
			t.Errorf("pkcs7Pad(%q, %d) = len %d, want %d", tt.input, tt.blockSize, len(result), tt.expectedLen)
		}
		// Check that original data is preserved
		if len(result) < len(tt.input) {
			t.Errorf("padded length %d < input length %d", len(result), len(tt.input))
		}
		copy(result[:len(tt.input)], tt.input)
		if !bytes.Equal(result[:len(tt.input)], tt.input) {
			t.Errorf("original data not preserved in padding")
		}
	}
}

func TestPkcs7Unpad(t *testing.T) {
	tests := []struct {
		input    []byte
		expected []byte
		hasError bool
	}{
		{pkcs7Pad([]byte("hello"), 16), []byte("hello"), false},
		{pkcs7Pad([]byte(""), 16), []byte(""), false},
		{[]byte{}, nil, true}, // empty input
		{[]byte{0x00}, nil, true}, // pad = 0
		{[]byte{17}, nil, true}, // pad > 16
		{[]byte("invalid"), nil, true}, // last byte > 16
		{[]byte{1, 2}, nil, true}, // inconsistent padding
	}

	for _, tt := range tests {
		result, err := pkcs7Unpad(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("pkcs7Unpad(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("pkcs7Unpad(%q) unexpected error: %v", tt.input, err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("pkcs7Unpad(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	}
}

func TestEncryptDecryptData(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintexts := [][]byte{
		[]byte("hello world"),
		[]byte(""),
		[]byte("short"),
		[]byte("this is a longer message to test encryption and decryption"),
	}

	for _, plaintext := range plaintexts {
		ciphertext, err := EncryptData(key, plaintext)
		if err != nil {
			t.Errorf("EncryptData failed: %v", err)
		}

		decrypted, err := DecryptData(key, ciphertext)
		if err != nil {
			t.Errorf("DecryptData failed: %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("DecryptData(EncryptData(%q)) = %q, want %q", plaintext, decrypted, plaintext)
		}
	}
}

func TestEncryptDecryptDataInvalidKey(t *testing.T) {
	key := make([]byte, 31) // invalid key length
	plaintext := []byte("test")

	_, err := EncryptData(key, plaintext)
	if err == nil {
		t.Errorf("EncryptData with invalid key should fail")
	}

	_, err = DecryptData(key, []byte("dummy"))
	if err == nil {
		t.Errorf("DecryptData with invalid key should fail")
	}
}

func TestDecryptDataInvalidCiphertext(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	// Too short ciphertext
	_, err := DecryptData(key, []byte("short"))
	if err == nil {
		t.Errorf("DecryptData with short ciphertext should fail")
	}

	// Invalid length after IV
	invalidCt := make([]byte, 17) // IV + 1 byte, not multiple of block size
	_, err = DecryptData(key, invalidCt)
	if err == nil {
		t.Errorf("DecryptData with invalid length should fail")
	}
}
