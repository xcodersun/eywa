package utils

import "net/url"

func QueryToMap(values url.Values) map[string]string {
	q := map[string][]string(values)
	r := make(map[string]string)
	for k, v := range q {
		if len(v) > 0 {
			r[k] = v[0]
		}
	}
	return r
}
