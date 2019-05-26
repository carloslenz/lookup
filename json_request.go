package lookup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type jsonRequestLooker struct {
	*http.Request

	mutex sync.Mutex
	data  map[string]interface{}
}

// NewJSONRequest returns a Looker to access r.Body.
func NewJSONRequest(req *http.Request) Looker {
	return &jsonRequestLooker{
		Request: req,
	}
}

func (l *jsonRequestLooker) LookupKey(k string) (string, bool, error) {
	l.mutex.Lock()
	if l.data == nil {
		// If body fails to load, don't try again for the same instance:
		l.data = make(map[string]interface{})

		err := json.NewDecoder(l.Body).Decode(&l.data)
		l.Body.Close()
		if err != nil {
			l.mutex.Unlock()
			return "", false, err
		}
	}
	l.mutex.Unlock()

	v, ok := l.data[k]
	if !ok {
		return "", false, nil
	}
	return fmt.Sprint(v), true, nil
}
