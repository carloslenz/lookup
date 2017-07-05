package lookup_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/carloslenz/lookup"
)

const (
	q    = "A=1&B=&C=2&D&-A=really&-other&blah"
	root = "/"
)

func TestFormLooker(t *testing.T) {
	tests := []struct {
		method, contentType, q, body string
	}{
		{http.MethodGet, "application/octet-stream", "?" + q, ""},
		{http.MethodPost, "application/x-www-form-urlencoded", "", q},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			b := bytes.NewBufferString(test.body)
			req, err := http.NewRequest(test.method, root+test.q, b)
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Content-Type", test.contentType)

			l := lookup.NewForm(req)
			tests := []struct {
				key, val string
				found    bool
			}{
				{"A", "1", true},
				{"-A", "really", true},
				{"B", "1", true},
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
		})
	}
}
