/*
Package lookup loads configuration (e.g, from environment variables) into struct.

Instructions

Define "lookup" tags for struct fields. The value should consist of the key to lookup followed by
",optional" when the field is not required. There is also compatibility with "encoding/json" tags,
so you don't need to define both if the keys match.

lookup.Lookup() accepts multiple Looker functions like lookup.Env. To adapt existing functions use
lookup.NoError and lookup.NoBool. To load system configuration files use lookup.NewJSON. Typically
the last step has the defaults in a lookup.Map.

Supported types

Everything fmt.Sscanln supports (because fmt.Sscan does not report an error when bools or floats
don't consume the string entirely) but newline is inserted internally. This means custom types can
implement fmt.Scanner. Exceptions:

	- string: used directly.
	- []byte: decoded as base64.
*/
package lookup

import (
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/carloslenz/epp"
)

type (
	// Looker is used by Lookup to lookup each field.
	Looker interface {
		LookupKey(string) (string, bool, error)
	}
	// NoError adapts functions like os.LookupEnv to match Looker signature.
	NoError struct {
		F func(string) (string, bool)
	}
	// NoBool adapts functions that return only value and error to match Looker signature.
	NoBool struct {
		F func(string) (string, error)
	}
	// Map implements Looker. Use it to store defaults.
	Map map[string]string
	// Reporter is used by Lookup to report each successfully loaded entry. It can be used for logs, etc.
	Reporter interface {
		Report(key string, e interface{})
	}
)

// Env wraps os.LookupEnv.
var Env = NoError{F: os.LookupEnv}

// LookupKey always returns err == nil.
func (l NoError) LookupKey(s string) (v string, b bool, err error) {
	v, b = l.F(s)
	return v, b, nil
}

// LookupKey returns b == True when err == nil.
func (l NoBool) LookupKey(s string) (v string, b bool, err error) {
	v, err = l.F(s)
	return v, err != nil, err
}

func lookupKey(s string, l []Looker) (v string, b bool, err error) {
	for _, e := range l {
		v, b, err = e.LookupKey(s)
		if err == nil && b {
			break
		}
	}
	return v, b, err
}

// LookupKey searches s in map.
func (l Map) LookupKey(s string) (v string, b bool, err error) {
	v, b = l[s]
	return v, b, nil
}

type discardReporter struct{}

var discard discardReporter

func (r discardReporter) Report(key string, e interface{}) {}

var lookupTags = []struct {
	tag, optional string
}{
	{"json", "omitempty"},
	{"lookup", "optional"},
}

// Lookup uses seq to fill in struct fields according to their tags.
// e should be a pointer to struct with "lookup" tags defined on its fields.
// For each field, items in seq are tried in sequence and lookup fails only if all of them fail.
// Can be nil.
func Lookup(e interface{}, r Reporter, seq ...Looker) error {
	value := reflect.ValueOf(e)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return epp.New("Lookup needs a pointer argument")
	}

	if r == nil {
		r = discard
	}

	value = value.Elem()
	t := value.Type()

	for i := 0; i < t.NumField(); i++ {
		field := value.Field(i)
		fieldType := t.Field(i)
		tag := fieldType.Tag
		fieldKey := fieldType.Name
		optional := false
		found := false

		for _, def := range lookupTags {
			if s, ok := tag.Lookup(def.tag); ok && s != "" {
				parts := strings.Split(s, ",")
				var key string
				switch len(parts) {
				case 0:
					// Default: use field name, not optional.
				case 1:
					key = s
				default:
					key = parts[0]
					optional = parts[1] == def.optional
				}
				fieldKey = key
				found = true
			}
		}
		if !found {
			continue
		}

		v, ok, err := lookupKey(fieldKey, seq)
		if err != nil {
			return epp.New("lookup for for field %q failed: %s", fieldType.Name, err)
		}
		if ok {
			var err error
			val := field.Interface()
			switch val.(type) {
			case string:
				field.SetString(v)
				r.Report(fieldKey, v)

			case []byte:
				b, err := base64.RawStdEncoding.DecodeString(v)
				if err != nil {
					field.SetBytes(b)
				}
				r.Report(fieldKey, b)

			default:
				if field.CanAddr() {
					var n int
					n, err = fmt.Sscanln(v+"\n", field.Addr().Interface())
					if err == nil && n != 1 {
						err = epp.New("")
					}
					if err == nil {
						r.Report(fieldKey, v)
					}
				} else {
					return epp.New("field %q of type %T is not addressable", v, fieldType.Name)
				}

			}
			if err != nil {
				return epp.New(
					"value %q for field %q is not %T: %s", v, fieldType.Name, val, err)
			}

		} else if !optional {
			return epp.New("missing value for required field %q", fieldType.Name)
		} else {
			r.Report(fieldKey, v)
		}
	}
	return nil
}
