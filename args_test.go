package lookup_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/carloslenz/lookup"
)

func TestArgsLooker(t *testing.T) {
	l := lookup.NewArgs("-", []string{"-A=1", "-B=", "-C=2", "-D", "--A=really", "--other", "blah"})
	tests := []struct {
		key, val string
		found    bool
	}{
		{"A", "1", true},
		{"-A", "really", true},
		{"B", "", true},
		{"C", "2", true},
		{"D", "1", true},
		{"E", "", false},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			v, got, err := l.LookupKey(test.key)
			if v != test.val {
				t.Errorf("Unexpected value: got %q, expecting %q", v, test.val)
			}

			if got != test.found {
				t.Errorf("Unexpected bool result: got %t, expecting %t", got, test.found)
			}

			if err != nil {
				t.Errorf("Unexpected error: got %q instead of nil", err)
			}
		})
	}

	extra := fmt.Sprint(l.ExtraArgs())
	expectedExtra := fmt.Sprint([]string{"blah"})
	if extra != expectedExtra {
		t.Errorf("Unexpected extra args: got %q, expecting %q", extra, expectedExtra)
	}
}

func TestArgsLookerEmptyPrefix(t *testing.T) {
	l := lookup.NewArgs("", []string{"A=0", "B", "-C-=2", "blah"})
	var b bytes.Buffer
	for _, k := range []string{"A", "B", "C", "-C-", "blah", "other"} {
		v, got, err := l.LookupKey(k)
		fmt.Fprintf(&b, "%s/%s/%t/%v\n", k, v, got, err)
	}
	expected := `A/0/true/<nil>
B/1/true/<nil>
C//false/<nil>
-C-/2/true/<nil>
blah/1/true/<nil>
other//false/<nil>
`
	obtained := b.String()
	if obtained != expected {
		t.Errorf("Unexpected results: got %q instead of %q", obtained, expected)
	}

	extra := fmt.Sprint(l.ExtraArgs())
	if extra != "[]" {
		t.Errorf("Unexpected extra args: got %q instead of []", extra)
	}
}
