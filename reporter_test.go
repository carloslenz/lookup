package lookup_test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/carloslenz/lookup"
)

var update = flag.Bool("update", false, "update .golden files")

func TestReporter(t *testing.T) {
	buf := new(bytes.Buffer)
	mr := lookup.NewMapReporter()
	r := lookup.DupReporter{
		lookup.FilterSecretsReporter{
			Reporter: lookup.FmtReporter{
				Writer: buf,
				Prefix: "- ",
			},
			Regexp: regexp.MustCompile(`.*SECRET.*`),
		},
		mr,
	}
	type cfg struct {
		AnySecret    string `lookup:"ANY_SECRET"`
		SecretAsWell string `lookup:"SECRET_AS_WELL,optional"`
		Public       string `json:"PUBLIC,omitempty"`
		OnTheRecord  string `json:"ON_THE_RECORD,omitempty"`
	}
	var data cfg
	defaults := lookup.Map{
		"ANY_SECRET":    "007 identity",
		"PUBLIC":        "Old news",
		"ON_THE_RECORD": "Everybody knows",
	}
	err := lookup.Lookup(&data, r, defaults)
	if err != nil {
		t.Fatal(err)
	}
	expectedMap := lookup.Map{
		"ANY_SECRET":     "007 identity",
		"SECRET_AS_WELL": "",
		"PUBLIC":         "Old news",
		"ON_THE_RECORD":  "Everybody knows",
	}
	if !reflect.DeepEqual(mr.Map(), expectedMap) {
		t.Errorf("Unexpected Map:\n***got***\n%v\n***\n%v", mr.Map(), expectedMap)
	}
	expectedData := cfg{
		AnySecret:    "007 identity",
		SecretAsWell: "",
		Public:       "Old news",
		OnTheRecord:  "Everybody knows",
	}
	if data != expectedData {
		t.Errorf("Unexpected data:\n***got***\n%v\n***\n%v", data, expectedData)
	}
	s := buf.String()
	golden := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		os.Mkdir("testdata", os.ModePerm)
		ioutil.WriteFile(golden, []byte(s), os.ModePerm)
	}
	b, _ := ioutil.ReadFile(golden)
	expected := string(b)
	if s != expected {
		t.Errorf("Unexpected output:\n***got***\n%s\n***expecting***\n%s", s, expected)
	}
}
