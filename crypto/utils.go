package crypto

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"

	"github.com/reinbowARA/PassLedger/models"
)

// GenerateSalt генерирует случайную соль заданной длины n байтов.
// Алгоритм: использует криптографически безопасный генератор случайных чисел (rand.Reader) для заполнения байтового массива.
func GenerateSalt(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

// HmacEqual сравнивает два HMAC значения в постоянное время, чтобы избежать атак по времени.
// Алгоритм: проверяет длины, затем побитово XOR каждого байта и возвращает true, если результат ноль.
func HmacEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var res byte = 0
	for i := range a {
		res |= a[i] ^ b[i]
	}
	return res == 0
}

// GeneratePassword генерирует случайный пароль на основе заданных опций (длина, наборы символов).
// Алгоритм: собирает charset из выбранных наборов, затем для каждой позиции пароля выбирает случайный символ из charset.
func GeneratePassword(options models.PasswordGeneratorOptions) (string, error) {
	var charset string
	if options.UseLowercase {
		charset += "abcdefghijklmnopqrstuvwxyz"
	}
	if options.UseUppercase {
		charset += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	if options.UseDigits {
		charset += "0123456789"
	}
	if options.UseSpecial {
		charset += "!@#$%^&*-_=+;:,.?/~`"
	}
	if options.UseSpace {
		charset += " "
	}
	if options.UseBrackets {
		charset += "[]{}()<>"
	}

	if len(charset) == 0 {
		return "", fmt.Errorf("no character sets selected")
	}

	charsetLen := big.NewInt(int64(len(charset)))
	password := make([]byte, options.Length)
	for i := 0; i < options.Length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		password[i] = charset[randomIndex.Int64()]
	}
	return string(password), nil
}
