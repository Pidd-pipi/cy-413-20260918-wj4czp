package repository

import (
	"errors"
	"testing"
)

func TestErrNotFoundIsStableSentinel(t *testing.T) {
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Fatal("ErrNotFound must be usable with errors.Is")
	}
}
