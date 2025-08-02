package parse

import "encoding/json"

// JSONParseFunc returns a function that parses a JSON-encoded byte slice into a value of type T.
func JSONParseFunc[T any](data []byte) (T, error) {
	var t T
	if len(data) == 0 {
		return t, nil
	}
	err := json.Unmarshal(data, &t)
	return t, err
}
