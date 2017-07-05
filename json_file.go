package lookup

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type jsonLooker struct {
	filename string

	mutex sync.Mutex
	data  map[string]interface{}
}

// NewJSONFile returns a Looker that can extracts data from JSON file. File is loaded only once.
func NewJSONFile(filename string) Looker {
	return &jsonLooker{
		filename: filename,
	}
}

func (l *jsonLooker) LookupKey(k string) (string, bool, error) {
	l.mutex.Lock()
	if l.data == nil {
		// If file fails to load, don't try again for the same instance:
		l.data = make(map[string]interface{})

		f, err := os.Open(l.filename)
		if err != nil {
			l.mutex.Unlock()
			return "", false, err
		}

		err = json.NewDecoder(f).Decode(&l.data)
		f.Close()
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
