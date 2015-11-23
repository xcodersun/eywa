package utils

import "encoding/json"

type MarshallableErrors map[string]error

func (m MarshallableErrors) MarshalJSON() ([]byte, error) {
	es := make(map[string]string)
	for key, e := range map[string]error(m) {
		es[key] = e.Error()
	}

	return json.Marshal(es)
}
