package lookup

import (
	"net/http"

	"golang.org/x/xerrors"
)

type formLooker struct {
	*http.Request
}

// NewForm returns a Looker to access r.Form. Any key present in req but empty is read as "1".
func NewForm(req *http.Request) Looker {
	return &formLooker{
		Request: req,
	}
}

func (l *formLooker) LookupKey(k string) (string, bool, error) {
	if err := l.ParseForm(); err != nil {
		return "", false, xerrors.Errorf("ParseForm failed", err)
	}

	v, ok := l.Form[k]
	if !ok {
		return "", false, nil
	}
	s := "1"
	for _, val := range v {
		if val != "" {
			s = val
			break
		}
	}
	return s, true, nil
}
