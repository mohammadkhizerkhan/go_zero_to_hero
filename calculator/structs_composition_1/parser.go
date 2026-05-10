package structscomposition1

import "encoding/json"

func ParseStruct(data []byte, v interface{}) interface{} {
	err := json.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
	return v
}
