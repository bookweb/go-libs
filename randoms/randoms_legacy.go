package randoms

import "math/rand"

// Go 1.21 and older

// Seed with the current time once at the start of your program
// rand.Seed(time.Now().UnixNano())

func RandIntnLegay(max int) int {
	return rand.Intn(max)
}

func RandIntnInRangeLegacy(min, max int) int {
	return rand.Intn((max - min)) + min
}

func RandFloat64Legacy() float64 { // 0.0 <= f < 1.0
	return rand.Float64()
}
