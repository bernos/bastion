package flags

import (
	"fmt"
	"sort"
	"strings"
)

type TagMap map[string]string

func (t *TagMap) Type() string { return "tags" }

// String returns tags as "key=value,key2=value2" — the same format accepted by Set.
// Viper reads this via BindPFlags when the flag is changed, so the format must be
// parseable by the config decode hook.
func (t *TagMap) String() string {
	if len(*t) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(*t))
	for k, v := range *t {
		pairs = append(pairs, k+"="+v)
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}
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
