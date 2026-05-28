package flags

import (
	"fmt"
	"strings"
)

type TagMap map[string]string

func (t *TagMap) Type() string   { return "tags" }
func (t *TagMap) String() string { return fmt.Sprintf("%v", *t) }
func (t *TagMap) Set(value string) error {
	pairs := strings.Split(value, ",")

	if *t == nil {
		*t = make(TagMap)
	}

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid tag format: %s (must be key=value)", pair)
		}

		(*t)[parts[0]] = parts[1]
	}

	return nil
}
