package lookup_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/carloslenz/lookup"
)

func TestLookup(t *testing.T) {
	type conf struct {
		A bool   `lookup:"A,optional"`
		B int    `lookup:"B"`
		C int64  `lookup:"C"`
		D string `lookup:"D"`
		E string `lookup:"E"`
	}
	if err := os.Setenv("B", "2"); err != nil {
		t.Fatalf("Could set env B: %s", err)
	}

	defaults := lookup.Map{
		"C": "-4",
		"E": "lorem ipsum",
	}

	var c conf
	var e entries

	e = nil
	if err := lookup.Lookup(&c, &e, lookup.NoError{F: os.LookupEnv}, defaults); err == nil {
		t.Fatalf("There are missing fields, why no error?! conf = %#v, entries = %#v", c, e)
	}

	if err := os.Setenv("D", "something"); err != nil {
		t.Fatalf("Could set env D: %s", err)
	}

	if err := lookup.Lookup(&c, nil, lookup.NoError{F: os.LookupEnv}, defaults); err != nil {
		t.Fatalf("There shouldn't be missing fields, yet error = %#v", err)
	}

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

	if err := os.Setenv("C", "8"); err != nil {
		t.Fatalf("Could set env C: %s", err)
	}

	if err := lookup.Lookup(&c, nil, lookup.NoError{F: os.LookupEnv}, defaults); err != nil {
		t.Fatalf("There shouldn't be missing fields, yet error = %#v", err)
	}

	expected.C = 8
	if c != expected {
		t.Errorf("Unexpected result: %#v, expecting %#v", c, expected)
	}

	if err := os.Setenv("A", "1"); err != nil {
		t.Fatalf("Could set env A: %s", err)
	}

	if err := lookup.Lookup(&c, nil, lookup.NoError{F: os.LookupEnv}, defaults); err != nil {
		t.Fatalf("There shouldn't be missing fields, yet error = %#v", err)
	}

	expected.A = true
	if c != expected {
		t.Errorf("Unexpected result: %#v, expecting %#v", c, expected)
	}

	if err := os.Setenv("E", "unimportant"); err != nil {
		t.Fatalf("Could set env A: %s", err)
	}

	if err := lookup.Lookup(&c, nil, lookup.NoError{F: os.LookupEnv}, defaults); err != nil {
		t.Fatalf("There shouldn't be missing fields, yet error = %#v", err)
	}

	expected.E = "unimportant"
	if c != expected {
		t.Errorf("Unexpected result: %#v, expecting %#v", c, expected)
	}

	if err := os.Setenv("A", "invalid bool"); err != nil {
		t.Fatalf("Could set env A: %s", err)
	}

	e = nil
	if err := lookup.Lookup(&c, &e, lookup.NoError{F: os.LookupEnv}, defaults); err == nil {
		t.Fatalf("A value has invalid type, why no error?! conf = %#v, entries = %#v", c, e)
	}

	if err := os.Setenv("A", "f"); err != nil {
		t.Fatalf("Could set env A: %s", err)
	}

	e = nil
	if err := lookup.Lookup(&c, &e, lookup.NoError{F: os.LookupEnv}, defaults); err != nil {
		t.Fatalf("There shouldn't be missing fields, yet error = %#v", err)
	}

	expected.A = false
	if c != expected {
		t.Errorf("Unexpected result: %#v, expecting %#v", c, expected)
	}

	expectedReports := entries{"A", "false", "B", "2", "C", "8", "D", "something", "E", "unimportant"}
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
	if err := lookup.Lookup(&c, &e, lookup.NoError{F: os.LookupEnv}, defaults); err == nil {
		t.Fatalf("C value in defaults has invalid type, why no error?! conf = %#v, entries = %#v", c, e)
	}

}

type entries []string

func (e *entries) Report(key string, v interface{}) {
	*e = append(*e, key, fmt.Sprint(v))
}
