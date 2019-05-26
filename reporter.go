package lookup

import (
	"fmt"
	"io"
	"regexp"
)

type (
	// Reporter is used by Lookup to report each successfully loaded entry. It can be used for logs, etc.
	Reporter interface {
		Report(key string, e interface{})
	}

	// FilterSecretsReporter forwards calls to Reporter replacing hidding the values of protected keys.
	FilterSecretsReporter struct {
		Reporter
		*regexp.Regexp
	}

	// FmtReporter outputs key-value pairs using fmt.Fprintf.
	FmtReporter struct {
		io.Writer
		Prefix string
	}

	// MapReporter stores key-value pairs a Map.
	MapReporter struct {
		dest Map
	}

	discardReporter struct{}
)

// Report forwards calls to embedded Reporter replacing protected entries with "(empty)" or
// "(not empty)".
func (r FilterSecretsReporter) Report(key string, e interface{}) {
	var v string
	if e != nil {
		v = fmt.Sprint(e)
	}
	if r.Regexp.MatchString(key) {
		if v == "" {
			v = "(empty)"
		} else {
			v = "(not empty)"
		}
	}
	r.Reporter.Report(key, v)
}

// Report outputs to embedded Writer.
func (r FmtReporter) Report(key string, e interface{}) {
	fmt.Fprintf(r.Writer, "%s%s=%v\n", r.Prefix, key, e)
}

// NewMapReporter creates a new MapReporter.
func NewMapReporter() MapReporter {
	return MapReporter{
		dest: make(Map),
	}
}

// Report stores key and e into an internal Map.
func (r MapReporter) Report(key string, e interface{}) {
	r.dest[key] = fmt.Sprint(e)
}

// Map returns the Map with stored key-value pairs.
func (r MapReporter) Map() Map {
	return r.dest
}

// DupReporter forwards Report calls to all items.
type DupReporter []Reporter

// Report is forwarded to all items.
func (r DupReporter) Report(key string, e interface{}) {
	for _, v := range r {
		v.Report(key, e)
	}
}

var discard discardReporter

func (r discardReporter) Report(key string, e interface{}) {}
