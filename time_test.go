package deterministic_test

import (
	"testing"
	"time"

	"github.com/selesy/deterministic"
)

func TestNowFunc_ReturnsFunction(t *testing.T) {
	f := deterministic.NowFunc()
	if f == nil {
		t.Fatal("NowFunc returned nil")
	}
}

func TestNowFunc_InitialTime(t *testing.T) {
	f := deterministic.NowFunc()
	result := f()
	expected, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	if !result.Equal(expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestNowFunc_IncrementsOneSecond(t *testing.T) {
	f := deterministic.NowFunc()
	first := f()
	second := f()
	expected := first.Add(time.Second)
	if !second.Equal(expected) {
		t.Errorf("got %v, want %v", second, expected)
	}
}

func TestNowFunc_ConsecutiveCalls(t *testing.T) {
	f := deterministic.NowFunc()
	baseTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	for i := 0; i < 10; i++ {
		result := f()
		expected := baseTime.Add(time.Duration(i) * time.Second)
		if !result.Equal(expected) {
			t.Errorf("call %d: got %v, want %v", i, result, expected)
		}
	}
}

func TestNowFunc_MultipleInstances(t *testing.T) {
	f1 := deterministic.NowFunc()
	f2 := deterministic.NowFunc()

	t1 := f1()
	t2 := f2()

	if !t1.Equal(t2) {
		t.Errorf("different instances should start at same time: got %v, want %v", t1, t2)
	}

	// Each instance maintains its own state independently
	t1_second := f1()
	t2_second := f2()

	expected := t1.Add(time.Second)
	if !t1_second.Equal(expected) {
		t.Errorf("f1 second: got %v, want %v", t1_second, expected)
	}
	if !t2_second.Equal(expected) {
		t.Errorf("f2 second: got %v, want %v", t2_second, expected)
	}
}
