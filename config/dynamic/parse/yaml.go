package parse

import "github.com/goccy/go-yaml"

// YAMLParseFunc returns a YAML parsing function
func YAMLParseFunc[T any](data []byte) (T, error) {
	var t T
	if len(data) == 0 {
		return t, nil
	}
	err := yaml.Unmarshal(data, &t)
	return t, err
}
