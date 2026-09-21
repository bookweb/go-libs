package numbers

import "testing"

func TestAdd01And02(t *testing.T) {
	doublePrecisionNumberMgmt := NewDoublePrecisionNumberManager()
	doublePrecisionNumber := doublePrecisionNumberMgmt.GetDoublePrecisionNumber(2)
	var a1 float64 = 0.1
	var a2 float64 = 0.2
	a3 := a1 + a2
	want := 0.3
	got := doublePrecisionNumber.Round(a3)
	if got != want {
		t.Errorf("round decimal failed, want: %v, got: %v", want, got)
	}
}

func TestRound1005(t *testing.T) {
	doublePrecisionNumberMgmt := NewDoublePrecisionNumberManager()
	doublePrecisionNumber := doublePrecisionNumberMgmt.GetDoublePrecisionNumber(2)
	var a1 float64 = 1.005
	want := 1.01
	got := doublePrecisionNumber.Round(a1)
	if got != want {
		t.Errorf("round decimal failed, want: %v, got: %v", want, got)
	}
}

func TestCeil1005(t *testing.T) {
	doublePrecisionNumberMgmt := NewDoublePrecisionNumberManager()
	doublePrecisionNumber := doublePrecisionNumberMgmt.GetDoublePrecisionNumber(2)
	var a1 float64 = 1.005
	want := 1.01
	got := doublePrecisionNumber.Ceil(a1)
	if got != want {
		t.Errorf("round decimal failed, want: %v, got: %v", want, got)
	}
}

func TestFloor1005(t *testing.T) {
	doublePrecisionNumberMgmt := NewDoublePrecisionNumberManager()
	doublePrecisionNumber := doublePrecisionNumberMgmt.GetDoublePrecisionNumber(2)
	var a1 float64 = 1.005
	want := 1.0
	got := doublePrecisionNumber.Floor(a1)
	if got != want {
		t.Errorf("round decimal failed, want: %v, got: %v", want, got)
	}
}
