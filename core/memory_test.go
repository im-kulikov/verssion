package core

import (
	"testing"
)

func TestMemoryDB(t *testing.T) {
	m := NewMemory()
	InterfaceTestDB(t, m)
}

func TestMemoryCurated(t *testing.T) {
	m := NewMemory()
	InterfaceTestCurated(t, m)
}
