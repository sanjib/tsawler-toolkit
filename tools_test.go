package toolkit_test

import (
	"github.com/sanjib/tsawler-toolkit"
	"testing"
)

func TestRandomString(t *testing.T) {
	randStr := toolkit.RandomString(8)
	t.Run("MatchLen", func(t *testing.T) {
		want := 8
		got := len(randStr)
		if want != got {
			t.Errorf("want %d, got %d", want, got)
		}
	})
}

func TestRandomStringMethod(t *testing.T) {
	tools := toolkit.Tools{}
	randStr := tools.RandomString(8)
	t.Run("MatchLen", func(t *testing.T) {
		want := 8
		got := len(randStr)
		if want != got {
			t.Errorf("want %d, got %d", want, got)
		}
	})
}
