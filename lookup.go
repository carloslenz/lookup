/*
Package lookup loads configuration (e.g, from environment variables) into struct.

Instructions

Define "lookup" tags for struct fields. The value should consist of the key to lookup followed by
",optional" when the field is not required. There is also compatibility with "encoding/json" tags,
so you don't need to define both if the keys match.

lookup.Lookup() accepts multiple Looker functions like lookup.Env. To adapt existing functions use
lookup.NoError and lookup.NoBool. To load system configuration files use lookup.NewJSONFile. Typically
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
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
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
		return errors.New("Lookup needs a pointer argument")
	}

	if r == nil {
		r = discard
	}

	value = value.Elem()
	t := value.Type()

	for i := 0; i < t.NumField(); i++ {
		field := value.Field(i)
		fieldType := t.Field(i)

		fieldKey, optional := findTag(fieldType.Tag)
		if fieldKey == notFound {
			continue
		}

		v, ok, err := lookupKey(fieldKey, seq)
		switch {
		case err != nil:
			return fmt.Errorf("lookup for for field %q failed: %s", fieldType.Name, err)
		case ok:
			if err = setField(field, v, fieldKey, fieldType.Name, r); err != nil {
				return fmt.Errorf(
					"value %q for field %q is not %T: %s", v, fieldType.Name, field.Interface(), err)
			}

		case !optional:
			return fmt.Errorf("missing value for required field %q", fieldType.Name)
		default:
			r.Report(fieldKey, v)
		}
	}
	return nil
}

const notFound = ""

func findTag(tag reflect.StructTag) (key string, optional bool) {
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
			return key, optional
		}
	}
	return notFound, false
}

func setField(field reflect.Value, v, fieldKey, fieldName string, r Reporter) error {
	val := field.Interface()
	switch val.(type) {
	case string:
		field.SetString(v)

	case []byte:
		b, err := base64.RawStdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
		field.SetBytes(b)
		r.Report(fieldKey, b)
		return nil

	default:
		if !field.CanAddr() {
			return fmt.Errorf("field %q of type %T is not addressable", v, fieldName)
		}
		n, err := fmt.Sscanln(v+"\n", field.Addr().Interface())
		if err != nil {
			return err
		}
		if n != 1 {
			return errors.New("nothing to read")
		}
	}
	r.Report(fieldKey, field.Interface())
	return nil
}
