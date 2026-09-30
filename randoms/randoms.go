package randoms

import (
	"math/rand/v2"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// from Go 1.22+, faster and automatically handles global seeding

func RandIntN(max int) int { // [0,max)
	return rand.IntN(max)
}

func RandIntNInRange(min, max int) int { // [min,max)
	return rand.IntN((max - min)) + min
}

func RandIntNInRangeIncludeMax(min, max int) int { // [min,max]
	return rand.IntN((max - min + 1)) + min
}

func RandFloat64() float64 { // [0.0, 1.0)
	return rand.Float64()
}

func RandFloat64InRange(min, max float64) float64 {
	return rand.Float64() * (max - min)
}

func RandBytes(length int) []byte {
	buf := make([]byte, 16)
	rng := rand.NewChaCha8([32]byte{1, 2, 3, 4})
	rng.Read(buf)
	return buf
}

func RandStringFromSource(length int) string {
	result := make([]byte, length)
	alphabetLength := len(alphabet)

	for i := 0; i < length; i++ {
		num := rand.IntN(alphabetLength)
		result[i] = alphabet[num]
	}

	return string(result)
}
