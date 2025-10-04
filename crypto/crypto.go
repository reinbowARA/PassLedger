package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	gost_kuznechik "github.com/pedroalbanese/gogost/gost3412128"
)

// PKCS7 padding
func pkcs7Pad(b []byte, blockSize int) []byte {
	pad := blockSize - (len(b) % blockSize)
	out := make([]byte, len(b)+pad)
	copy(out, b)
	for i := len(b); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

func pkcs7Unpad(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("пустой буфер при unpad")
	}
	pad := int(b[len(b)-1])
	if pad <= 0 || pad > 16 || pad > len(b) {
		return nil, fmt.Errorf("неверный PKCS7 padding")
	}
	// basic verification
	for i := len(b) - pad; i < len(b); i++ {
		if int(b[i]) != pad {
			return nil, fmt.Errorf("неверный PKCS7 padding (интегритет)")
		}
	}
	return b[:len(b)-pad], nil
}

// EncryptData шифрует данные Кузнечиком (CBC + PKCS7). Возвращает IV||CT.
func EncryptData(key, plaintext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("ключ должен быть 32 байта")
	}

	// создаём блок Кузнечик
	block := gost_kuznechik.NewCipher(key)


	blockSize := block.BlockSize() // должен быть 16
	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	padded := pkcs7Pad(plaintext, blockSize)
	ct := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ct, padded)

	return append(iv, ct...), nil
}

// DecryptData расшифровывает данные, ожидает IV||CT
func DecryptData(key, ciphertext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("ключ должен быть 32 байта")
	}

	block := gost_kuznechik.NewCipher(key)

	blockSize := block.BlockSize()
	if len(ciphertext) < blockSize {
		return nil, fmt.Errorf("короткий шифротекст")
	}
	iv := ciphertext[:blockSize]
	ct := ciphertext[blockSize:]
	if len(ct)%blockSize != 0 {
		return nil, fmt.Errorf("неправильная длина шифротекста")
	}

	pt := make([]byte, len(ct))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(pt, ct)

	return pkcs7Unpad(pt)
}
