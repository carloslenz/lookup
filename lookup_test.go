package lookup_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

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

	const filename = "testdata/lookup.json"
	os.Mkdir("testdata", 0777)
	if err := ioutil.WriteFile(filename, []byte(`{"E1":"lorem ipsum", "B": 2}`), 0666); err != nil {
		t.Fatalf("Cannot write testdata file: %s", err)
	}

	defaults := lookup.Map{
		"C": "-4",
	}

	var c conf
	var e entries

	e = nil
	json := lookup.NewJSON(filename)
	if err := lookup.Lookup(&c, &e, lookup.Env, json, defaults); err == nil {
		t.Fatalf("There are missing fields, why no error?! conf = %#v, entries = %#v", c, e)
	}

	mustLookup := func() {
		if err := lookup.Lookup(&c, nil, lookup.Env, json, defaults); err != nil {
			t.Fatalf("There shouldn't be missing fields, yet error = %s", err)
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

	if err := os.Unsetenv("C"); err != nil {
		t.Fatalf("Could set unset env C: %s", err)
	}

	e = nil
	if err := lookup.Lookup(&c, &e, lookup.Env, json, defaults); err == nil {
		t.Fatalf("C value in defaults has invalid type, why no error?! conf = %#v, entries = %#v", c, e)
	}

}

func mustSetenv(t *testing.T, k, v string) {
	if err := os.Setenv(k, v); err != nil {
		t.Fatalf("Could set env %s: %s", k, err)
	}
}

type entries []string

func (e *entries) Report(key string, v interface{}) {
	*e = append(*e, key, fmt.Sprint(v))
}
