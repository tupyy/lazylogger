package reader

import (
	"errors"
	"testing"
)

func TestHandleStateChange(t *testing.T) {
	s := NewState(1)

	stderr := errors.New("stderr")
	err := errors.New("err")

	s.HandleStateChange(stderr, nil)
	if s.Health != DEGRADED {
		t.Errorf("Expected: DEGRADED. Actual: %s", s.String())
	}

	s.HandleStateChange(stderr, err)
	if s.Health != FAILED {
		t.Errorf("Expected: FAILED. Actual: %s", s.String())
	}

	s.HandleStateChange(nil, nil)
	if s.Health != HEALTHY {
		t.Errorf("Expected: HEALTHY. Actual: %s", s.String())
	}
}
