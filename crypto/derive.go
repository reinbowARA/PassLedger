package crypto

import (
	"crypto/hmac"
	"encoding/binary"
	"fmt"
	"hash"

	gost_streebog "github.com/pedroalbanese/gogost/gost34112012256"
	"golang.org/x/crypto/pbkdf2"
)

// HMACStreebog256 вычисляет HMAC с использованием Streebog-256
func HMACStreebog256(key, data []byte) []byte {
	mac := hmac.New(func() hash.Hash { return gost_streebog.New() }, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// KDF_GOSTR3411_2012_256 -- простой контр-ориентированный KDF на HMAC(Streebog)
// seed — ключ (или псевдослучай), label/context — дополнительные поля
func KDF_GOSTR3411_2012_256(seed, label, context []byte, keySize int) ([]byte, error) {
	if keySize <= 0 || keySize > 64 {
		return nil, fmt.Errorf("неверный размер ключа: %d", keySize)
	}
	var out []byte
	counter := uint32(1)
	for len(out) < keySize {
		h := hmac.New(func() hash.Hash { return gost_streebog.New() }, seed)
		// data = label || 0x00 || context || counter_be
		h.Write(label)
		h.Write([]byte{0})
		h.Write(context)
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], counter)
		h.Write(b[:])
		out = append(out, h.Sum(nil)...)
		counter++
	}
	return out[:keySize], nil
}

// DeriveKeyFromPassword: комбинируем HMAC(Streebog) + PBKDF2(Streebog) + KDF
// Возвращает 32-байтовый ключ (для Кузнечика используем 32 байта)
func DeriveKeyFromPassword(password []byte, salt []byte, iterations int) ([]byte, error) {
	// 1) первичный HMAC от пароля
	hmacKey := HMACStreebog256(password, password)

	// 2) PBKDF2 с функцией Streebog
	pbkdf2Key := pbkdf2.Key(hmacKey, salt, iterations, 32, func() hash.Hash {
		return gost_streebog.New()
	})

	// 3) Дополнительный KDF
	label := []byte("шифр")
	context := []byte("")
	finalKey, err := KDF_GOSTR3411_2012_256(pbkdf2Key, label, context, 32)
	if err != nil {
		return nil, err
	}
	return finalKey, nil
}
