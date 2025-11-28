package crypto

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"

	"github.com/reinbowARA/PassLedger/models"
)

func GenerateSalt(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

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
