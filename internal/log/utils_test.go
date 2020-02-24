package log

import "testing"

import (
	"github.com/tupyy/lazylogger/internal/conf"
)

func TestMapToArray(t *testing.T) {
	a := conf.LoggerConfiguration{}
	arr := []conf.LoggerConfiguration{a, a}

	m := mapFromArray(arr)

	if _, ok := m[0]; !ok {
		t.Error("Expected one element. Actual none")
	}

	if len(m) != 2 {
		t.Errorf("Expected len(m)=2. Actual: %d", len(m))
	}
}
