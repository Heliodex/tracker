package song

import (
	"reflect"
	"testing"
)

func TestAtCountsRowsFromTheStartOfThePattern(t *testing.T) {
	if got, want := At(2, 5), 2*RowsPerBar+5; got != want {
		t.Errorf("At(2, 5) = %d, want %d", got, want)
	}
}

func TestPoints(t *testing.T) {
	want := []Point{
		{
			X: 0,
			Y: 0,
		},
		{
			X: 24,
			Y: 64,
		},
	}
	if got := Points(0, 0, 24, 64); !reflect.DeepEqual(got, want) {
		t.Errorf("Points(0, 0, 24, 64) = %v, want %v", got, want)
	}
}

func TestPointsOddPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("expected a panic")
		}
	}()

	Points(0, 0, 24)
}

func TestSteps(t *testing.T) {
	if got, want := Steps(0, 16, 4), []int{0, 4, 8, 12}; !reflect.DeepEqual(got, want) {
		t.Errorf("Steps(0, 16, 4) = %v, want %v", got, want)
	}

	if got := Steps(4, 4, 1); len(got) != 0 {
		t.Errorf("Steps(4, 4, 1) = %v, want none", got)
	}
}

func TestVolOpts(t *testing.T) {
	if got := VolOpts(0); len(got) != 0 {
		t.Errorf("VolOpts(0) = %d options, want none", len(got))
	}

	if got := VolOpts(32, Pan(128)); len(got) != 2 {
		t.Errorf("VolOpts(32, Pan) = %d options, want 2", len(got))
	}
}
