package crypto

import (
	"bytes"
	"testing"
)

func TestHMACStreebog256(t *testing.T) {
	key := []byte("test key")
	data := []byte("test data")

	hmac1 := HMACStreebog256(key, data)
	hmac2 := HMACStreebog256(key, data)

	if !bytes.Equal(hmac1, hmac2) {
		t.Errorf("HMACStreebog256 not deterministic")
	}

	if len(hmac1) != 32 {
		t.Errorf("HMACStreebog256 length = %d, want 32", len(hmac1))
	}

	// Different key should produce different hmac
	key2 := []byte("different key")
	hmac3 := HMACStreebog256(key2, data)
	if bytes.Equal(hmac1, hmac3) {
		t.Errorf("HMACStreebog256 with different keys produced same result")
	}
}

func TestKDF_GOSTR3411_2012_256(t *testing.T) {
	seed := []byte("seed")
	label := []byte("label")
	context := []byte("context")

	key, err := KDF_GOSTR3411_2012_256(seed, label, context, 32)
	if err != nil {
		t.Errorf("KDF_GOSTR3411_2012_256 failed: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("KDF_GOSTR3411_2012_256 length = %d, want 32", len(key))
	}

	// Same inputs should produce same key
	key2, err := KDF_GOSTR3411_2012_256(seed, label, context, 32)
	if err != nil {
		t.Errorf("KDF_GOSTR3411_2012_256 failed second time: %v", err)
	}
	if !bytes.Equal(key, key2) {
		t.Errorf("KDF_GOSTR3411_2012_256 not deterministic")
	}

	// Different keySize
	key16, err := KDF_GOSTR3411_2012_256(seed, label, context, 16)
	if err != nil {
		t.Errorf("KDF_GOSTR3411_2012_256 with keySize 16 failed: %v", err)
	}
	if len(key16) != 16 {
		t.Errorf("KDF_GOSTR3411_2012_256 length = %d, want 16", len(key16))
	}

	// Invalid keySize
	_, err = KDF_GOSTR3411_2012_256(seed, label, context, 0)
	if err == nil {
		t.Errorf("KDF_GOSTR3411_2012_256 with keySize 0 should fail")
	}

	_, err = KDF_GOSTR3411_2012_256(seed, label, context, 65)
	if err == nil {
		t.Errorf("KDF_GOSTR3411_2012_256 with keySize 65 should fail")
	}
}

func TestDeriveKeyFromPassword(t *testing.T) {
	password := []byte("password")
	salt := []byte("salt")
	iterations := 1000

	key1, err := DeriveKeyFromPassword(password, salt, iterations)
	if err != nil {
		t.Errorf("DeriveKeyFromPassword failed: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("DeriveKeyFromPassword length = %d, want 32", len(key1))
	}

	// Same inputs should produce same key
	key2, err := DeriveKeyFromPassword(password, salt, iterations)
	if err != nil {
		t.Errorf("DeriveKeyFromPassword failed second time: %v", err)
	}
	if !bytes.Equal(key1, key2) {
		t.Errorf("DeriveKeyFromPassword not deterministic")
	}

	// Different password should produce different key
	password2 := []byte("different")
	key3, err := DeriveKeyFromPassword(password2, salt, iterations)
	if err != nil {
		t.Errorf("DeriveKeyFromPassword with different password failed: %v", err)
	}
	if bytes.Equal(key1, key3) {
		t.Errorf("DeriveKeyFromPassword with different passwords produced same key")
	}

	// Different salt
	salt2 := []byte("different salt")
	key4, err := DeriveKeyFromPassword(password, salt2, iterations)
	if err != nil {
		t.Errorf("DeriveKeyFromPassword with different salt failed: %v", err)
	}
	if bytes.Equal(key1, key4) {
		t.Errorf("DeriveKeyFromPassword with different salts produced same key")
	}
}
