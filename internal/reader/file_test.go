package reader

import "testing"

func TestComputeNextChunk(t *testing.T) {
	res := computeNextChunkSize(100, 20, 20)
	if res != 20 {
		t.Errorf("Expected: %d. Actual: %d", 40, res)
	}

	res = computeNextChunkSize(100, 100, 20)
	if res != 0 {
		t.Errorf("Expected: %d. Actual: %d", 0, res)
	}

	res = computeNextChunkSize(100, 90, 20)
	if res != 10 {
		t.Errorf("Expected: %d. Actual: %d", 10, res)
	}
}
