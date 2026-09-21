package numbers

import (
	"math"
	"sync"
)

const Epsilon float64 = 2.220446049250313e-16

// var Epsilon = math.Nextafter(1.0, 2.0) - 1.0 // same above

type DoublePrecisionManager struct {
	numberMap sync.Map
}

func NewDoublePrecisionNumberManager() *DoublePrecisionManager {
	return &DoublePrecisionManager{}
}

func (h *DoublePrecisionManager) GetDoublePrecisionNumber(scale int) DoublePrecisionNumber {
	if val, ok := h.numberMap.Load(scale); ok {
		return val.(DoublePrecisionNumber)
	}

	newDoublePrecisionNumber := MustGetDoublePrecisionNumber(scale)

	actual, _ := h.numberMap.LoadOrStore(scale, newDoublePrecisionNumber)

	return actual.(DoublePrecisionNumber)
}

type DoublePrecisionNumber interface {
	Round(number float64) float64

	Ceil(number float64) float64

	Floor(number float64) float64
}

func MustGetDoublePrecisionNumber(scale int) DoublePrecisionNumber {
	return &doublePrecisionNumber{
		scale: scale,
	}
}

type (
	doublePrecisionNumber struct {
		scale int
	}
)

func (h *doublePrecisionNumber) Round(number float64) float64 {
	return math.Round((number+Epsilon)*math.Pow10(h.scale)) / math.Pow10(h.scale)
}

func (h *doublePrecisionNumber) Ceil(number float64) float64 {
	return math.Ceil(number*math.Pow10(h.scale)) / math.Pow10(h.scale)
}

func (h *doublePrecisionNumber) Floor(number float64) float64 {
	return math.Floor(number*math.Pow10(h.scale)) / math.Pow10(h.scale)
}

func NearlyEqual(a, b, epsilon float64) bool {
	if a == b {
		return true
	}
	diff := math.Abs(a - b)
	if a == 0 || b == 0 || diff < math.SmallestNonzeroFloat64 {
		return diff < epsilon
	}
	return diff/(math.Abs(a)+math.Abs(b)) < epsilon
}
