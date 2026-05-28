package flags

import "testing"

func Test_TagMap_Set(t *testing.T) {
	value := "tag1=value1,tag2=value2"
	flag := TagMap(map[string]string{})

	if err := flag.Set(value); err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
}

func Test_TagMap_Error_Cases(t *testing.T) {

	cases := []struct {
		value string
	}{
		{value: "tag1="},
		{value: "tag1=value1,"},
		{value: ","},
		{value: "tag"},
	}

	for _, tt := range cases {
		t.Run(tt.value, func(t *testing.T) {
			flag := TagMap(map[string]string{})

			if err := flag.Set(tt.value); err == nil {
				t.Errorf("unexpected error for %q, but got none", tt.value)
			}

		})
	}
}
