package domain

import (
	"errors"
	"testing"
)

func TestNormalizeAppError_WrapsUnknownToInternal(t *testing.T) {
	orig := errors.New("boom")
	err := NormalizeAppError(orig)
	if err == nil || !IsErrorCode(err, ErrorCodeInternal) {
		t.Fatalf("expected internal, got: %v", err)
	}
}

func TestNormalizeAppError_Passthrough(t *testing.T) {
	orig := NotFoundError("x", nil)
	err := NormalizeAppError(orig)
	if err != orig {
		t.Fatalf("expected same error")
	}
}

func TestIsErrorCode(t *testing.T) {
	err := ConflictError("c", nil)
	if !IsErrorCode(err, ErrorCodeConflict) {
		t.Fatalf("expected conflict")
	}
}
