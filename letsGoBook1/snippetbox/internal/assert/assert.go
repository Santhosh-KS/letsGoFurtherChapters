package assert

import (
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, actual, exptected T) {
	t.Helper()

	if actual != exptected {
		t.Errorf("got: %v; want: %v", actual, exptected)
	}
}

func StringContains(t *testing.T, actual, expectedSubstring string) {
	t.Helper()
	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("got: %q; epected to contain: %q", actual, expectedSubstring)
	}
}

func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got: %v; expected: nil", actual)
	}
}
