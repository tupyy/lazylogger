package log

import "testing"

func TestRead(t *testing.T) {
	c := newCache()

	data := []byte{'l', 'o', 'g', 'g', 'e', 'r'}
	n, err := c.Write(data)
	if err != nil {
		t.Error(err)
	}

	if n != len(data) {
		t.Error("error writing in cache")
	}

	dataRead := make([]byte, len(data))
	c.ReadAt(dataRead, 0)
	if string(dataRead) != string(data) {
		t.Errorf("Expected: %s. Actual: %s", string(dataRead), string(data))
	}

}

func TestReadOffset(t *testing.T) {
	c := newCache()

	data := []byte{'l', 'o', 'g', 'g', 'e', 'r'}
	n, err := c.Write(data)
	if err != nil {
		t.Error(err)
	}

	if n != len(data) {
		t.Error("error writing in cache")
	}

	dataRead := make([]byte, 1)
	c.ReadAt(dataRead, 2)
	if string(dataRead) != "g" {
		t.Errorf("Expected: \"g\". Actual: %s", string(dataRead))
	}
}

func TestWrite(t *testing.T) {
	c := newCache()

	data := make([]byte, MaxCacheSize)
	for i, _ := range data {
		data[i] = 0
	}
	data[0] = 1
	data[1] = 1

	n, err := c.Write(data)
	if err != nil {
		t.Error(err)
	}

	if n != len(data) {
		t.Error("error writing in cache")
	}

	// write other 2 bytes at the end. The first 2 bytes should be 0 and 0.
	d := []byte{2, 2}
	c.Write(d)

	dataRead := make([]byte, 2)
	c.ReadAt(dataRead, 0)
	if dataRead[0] != 0 || dataRead[1] != 0 {
		t.Errorf("Expected: 0 0. Actual: %d %d", dataRead[0], dataRead[1])
	}

	dataEndRead := make([]byte, 2)
	c.ReadAt(dataEndRead, MaxCacheSize-2)
	if dataEndRead[0] != 2 || dataEndRead[1] != 2 {
		t.Errorf("Expected: 2 2. Actual: %d %d", dataEndRead[0], dataEndRead[1])
	}

}
