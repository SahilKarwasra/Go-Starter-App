package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateNumericOTP generates a cryptographically secure random numeric OTP with the specified length.
func GenerateNumericOTP(length int) (string, error) {
	const digits = "0123456789"
	otp := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}
	return string(otp), nil
}
