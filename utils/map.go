package utils

import "errors"

func ToStringMap(m map[interface{}]interface{}) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	for k, v := range m {
		if kStr, ok := k.(string); ok {
			if vMap, ok := v.(map[interface{}]interface{}); ok {
				vCon, err := ToStringMap(vMap)
				if err != nil {
					return res, err
				} else {
					res[kStr] = vCon
				}
			} else {
				res[kStr] = v
			}
		} else {
			return res, errors.New("key in the map is not a string")
		}
	}

	return res, nil
}
