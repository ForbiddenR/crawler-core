package utils

import "encoding/json"

func JsonToBytes(d any) (bytes []byte, err error) {
	switch t := d.(type) {
	case []byte:
		return t, nil
	default:
		return json.Marshal(d)
	}
}
