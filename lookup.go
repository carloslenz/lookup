/*
Package lookup loads data (e.g, from environment variables) into struct.

Instructions

Define "lookup" tags for struct fields. The value should consist of the key to lookup followed by ",optional" when the field is not required.

Provide an extraction function (e.g, os.LookupEnv), using NoError or NoBool to adapt functions with different signatures.

Lookup sequences may be defined. Typically the last step has the defaults in a Map.

Encoding

complex64 and complex128: r,i separated by comma.

[]byte: base64.

*/
package lookup

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type (
	// Looker is used by Lookup to lookup each field.
	Looker interface {
		LookupKey(string) (string, bool, error)
	}
	// Seq tries a series of Looker instances in sequence.
	Seq []Looker
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

// LookupKey returns the first response among its items where: err == nil && b == True. Otherwise the last response is returned.
func (l Seq) LookupKey(s string) (v string, b bool, err error) {
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

// Lookup uses l to fill in fields according to their struct tags.
// e should be a pointer to struct with "lookup" tags defined on its fields.
// Only r can be nil.
func Lookup(e interface{}, l Looker, r Reporter) error {
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
		tag := fieldType.Tag
		fieldKey := fieldType.Name
		optional := false

		if s, ok := tag.Lookup("lookup"); ok && s != "" {
			parts := strings.Split(s, ",")
			var key string
			switch len(parts) {
			case 1:
				key = s
			default:
				key = parts[0]
				optional = parts[1] == "optional"
			case 0:
			}
			fieldKey = key
		} else {
			continue
		}

		v, ok, err := l.LookupKey(fieldKey)
		if err != nil {
			return fmt.Errorf("lookup for for field %q failed: %s", fieldType.Name, err)
		}
		if ok {
			var err error
			val := field.Interface()
			switch val.(type) {
			case string:
				field.SetString(v)
				r.Report(fieldKey, v)

			case int:
				err = setInt(field, v, 64, r, fieldKey)
			case int8:
				err = setInt(field, v, 8, r, fieldKey)
			case int16:
				err = setInt(field, v, 16, r, fieldKey)
			case int32:
				err = setInt(field, v, 32, r, fieldKey)
			case int64:
				err = setInt(field, v, 64, r, fieldKey)

			case uint:
				err = setUint(field, v, 64, r, fieldKey)
			case uint8:
				err = setUint(field, v, 8, r, fieldKey)
			case uint16:
				err = setUint(field, v, 16, r, fieldKey)
			case uint32:
				err = setUint(field, v, 32, r, fieldKey)
			case uint64:
				err = setUint(field, v, 64, r, fieldKey)

			case bool:
				b, err := strconv.ParseBool(v)
				if err != nil {
					return fmt.Errorf(
						"value %q for field %q is not bool: %s", v, fieldType.Name, err)
				}
				field.SetBool(b)
				r.Report(fieldKey, b)

			case float32:
				err = setFloat(field, v, 32, r, fieldKey)
			case float64:
				err = setFloat(field, v, 64, r, fieldKey)

			case complex64:
				err = setComplex(field, v, 32, r, fieldKey)
			case complex128:
				err = setComplex(field, v, 64, r, fieldKey)

			case []byte:
				b, err := base64.RawStdEncoding.DecodeString(v)
				if err != nil {
					field.SetBytes(b)
				}
				r.Report(fieldKey, b)
			}
			if err != nil {
				return fmt.Errorf(
					"value %q for field %q is not %T: %s", v, fieldType.Name, val, err)
			}

		} else if !optional {
			return fmt.Errorf("missing value for required field %q", fieldType.Name)
		} else {
			r.Report(fieldKey, v)
		}
	}
	return nil
}

func setUint(field reflect.Value, s string, bits int, r Reporter, fieldKey string) error {
	u, err := strconv.ParseUint(s, 0, bits)
	if err != nil {
		return err
	}
	field.SetUint(u)
	r.Report(fieldKey, u)
	return nil
}

func setInt(field reflect.Value, s string, bits int, r Reporter, fieldKey string) error {
	i, err := strconv.ParseInt(s, 0, bits)
	if err != nil {
		return err
	}
	field.SetInt(i)
	r.Report(fieldKey, i)
	return nil
}

func setFloat(field reflect.Value, s string, bits int, r Reporter, fieldKey string) error {
	f, err := strconv.ParseFloat(s, bits)
	if err != nil {
		return err
	}
	field.SetFloat(f)
	r.Report(fieldKey, f)
	return nil
}

func setComplex(field reflect.Value, s string, bits int, r Reporter, fieldKey string) error {
	var c [2]float64
	for i, q := range strings.Split(s, ",")[:2] {
		f, err := strconv.ParseFloat(q, bits)
		if err != nil {
			return err
		}
		c[i] = f
	}
	v := complex(c[0], c[1])
	field.SetComplex(v)
	r.Report(fieldKey, v)
	return nil
}
