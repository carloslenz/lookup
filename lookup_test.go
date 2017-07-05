package lookup_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"net/http"

	"github.com/carloslenz/lookup"
)

func TestLookup(t *testing.T) {
	type conf struct {
		A bool   `lookup:"A,optional"`
		B int    `lookup:"B"`
		C int64  `json:"C"`
		D string `lookup:"D"`
		E string `json:"E1"`
	}

	const (
		filename     = "testdata/lookup.json"
		jsonContents = `{"E1":"lorem ipsum", "B": 2}`
	)
	os.Mkdir("testdata", 0777)
	if err := ioutil.WriteFile(filename, []byte(jsonContents), 0666); err != nil {
		t.Fatalf("Cannot write testdata file: %s", err)
	}

	var c conf
	var e entries

	e = nil
	json1 := lookup.NewJSONFile(filename)

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(jsonContents))
	if err != nil {
		t.Fatalf("Cannot create request: %s", err)
	}
	json2 := lookup.NewJSONRequest(req)

	for i, json := range []lookup.Looker{json1, json2} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			defaults := lookup.Map{
				"C": "-4",
			}

			if err := lookup.Lookup(&c, &e, lookup.Env, json, defaults); err == nil {
				t.Fatalf("There are missing fields, why no error?! conf = %#v, entries = %#v", c, e)
			}

			mustLookup := func() {
				if err := lookup.Lookup(&c, nil, lookup.Env, json, defaults); err != nil {
					t.Errorf("There shouldn't be missing fields, yet error = %s", err)
				}
			}
			mustSetenv(t, "D", "something")
			mustLookup()

			expected := conf{
				A: false,
				B: 2,
				C: -4,
				D: "something",
				E: "lorem ipsum",
			}

			if c != expected {
				t.Errorf("Unexpected result: %#v, expecting %#v", c, expected)
			}

			mustSetenv(t, "C", "8")
			mustLookup()

			expected.C = 8
			if c != expected {
				t.Errorf("Unexpected result: %#v, expecting %#v", c, expected)
			}

			mustSetenv(t, "A", "1")
			mustLookup()

			expected.A = true
			if c != expected {
				t.Errorf("Unexpected result: %#v, expecting %#v", c, expected)
			}

			mustSetenv(t, "E1", "unimportant")
			mustLookup()

			expected.E = "unimportant"
			if c != expected {
				t.Errorf("Unexpected result: %#v, expecting %#v", c, expected)
			}

			e = nil
			mustSetenv(t, "A", "invalid bool")
			if err := lookup.Lookup(&c, &e, lookup.Env, json, defaults); err == nil {
				t.Fatalf("A value has invalid type, why no error?! conf = %#v, entries = %#v", c, e)
			}

			e = nil
			mustSetenv(t, "A", "f")
			mustLookup()
			if err := lookup.Lookup(&c, &e, lookup.Env, json, defaults); err != nil {
				t.Fatalf("There shouldn't be missing fields, yet error = %s", err)
			}

			expected.A = false
			if c != expected {
				t.Errorf("Unexpected result: %#v, expecting %#v", c, expected)
			}

			expectedReports := entries{"A", "false", "B", "2", "C", "8", "D", "something", "E1", "unimportant"}
			if !reflect.DeepEqual(e, expectedReports) {
				t.Errorf("Unexpected reports: %#v, expecting %#v", e, expectedReports)
			}

			e = nil
			defaults = lookup.Map{
				"C": "4.9",
			}

			mustUnsetenv(t, "C")

			e = nil
			if err := lookup.Lookup(&c, &e, lookup.Env, json, defaults); err == nil {
				t.Fatalf("C value in defaults has invalid type, why no error?! conf = %#v, entries = %#v", c, e)
			}

			mustUnsetenv(t, "A")
			mustUnsetenv(t, "D")
			mustUnsetenv(t, "E1")
		})
	}
}

func mustSetenv(t *testing.T, k, v string) {
	if err := os.Setenv(k, v); err != nil {
		t.Fatalf("Could set env %s: %s", k, err)
	}
}

func mustUnsetenv(t *testing.T, k string) {
	if err := os.Unsetenv(k); err != nil {
		t.Fatalf("Could unset env %s: %s", k, err)
	}
}

type entries []string

func (e *entries) Report(key string, v interface{}) {
	*e = append(*e, key, fmt.Sprint(v))
}
