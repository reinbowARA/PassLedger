package crypto

import (
	"crypto/hmac"
	"encoding/binary"
	"fmt"
	"hash"

	gost_streebog "github.com/pedroalbanese/gogost/gost34112012256"
	"golang.org/x/crypto/pbkdf2"
)

// HMACStreebog256 вычисляет HMAC (Hash-based Message Authentication Code) используя хеш-функцию Streebog-256.
// Алгоритм: создаёт HMAC объект с ключом, записывает данные и возвращает итоговый хеш.
func HMACStreebog256(key, data []byte) []byte {
	mac := hmac.New(func() hash.Hash { return gost_streebog.New() }, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// KDF_GOSTR3411_2012_256 реализует ключевой выводной функцию (KDF) на основе HMAC с Streebog, как определено в ГОСТ.
// Алгоритм: использует счётчик, для каждого шага вычисляет HMAC от seed с label, 0x00, context и счётчиком,
// собирает выходные байты до достижения требуемого размера ключа.
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

// DeriveKeyFromPassword деривирует криптографический ключ из пароля, соли и количества итераций.
// Алгоритм: сначала вычисляет HMAC от пароля для начального ключа, затем применяет PBKDF2 с Streebog,
// и наконец использует дополнительную KDF для получения финального 32-байтового ключа.
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
