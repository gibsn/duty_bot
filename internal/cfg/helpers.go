package cfg

import "fmt"

func ParamWithPrefix(prefix string) func(string) string {
	return func(name string) string {
		return fmt.Sprintf("%s.%s", prefix, name)
	}
}
