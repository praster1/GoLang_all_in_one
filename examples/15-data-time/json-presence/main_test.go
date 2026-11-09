package main

import (
	"encoding/json"
	"testing"
)

func TestPresence(t *testing.T) {
	cases := []struct{ input, description, output string }{
		{`{}`, "미지정", `{}`},
		{`{"enabled":null}`, "미지정", `{}`},
		{`{"enabled":false}`, "꺼짐", `{"enabled":false}`},
		{`{"enabled":true}`, "켜짐", `{"enabled":true}`},
	}
	for _, test := range cases {
		var settings Settings
		if err := json.Unmarshal([]byte(test.input), &settings); err != nil {
			t.Fatal(err)
		}
		output, err := json.Marshal(settings)
		if err != nil || string(output) != test.output || describe(settings) != test.description {
			t.Errorf("%s => %s, %s, %v", test.input, describe(settings), output, err)
		}
	}
	var settings Settings
	if err := json.Unmarshal([]byte(`{"enabled":"false"}`), &settings); err == nil {
		t.Fatal("string must not decode as bool")
	}
}

func Example_main() {
	main()
	// Output:
	// {} => 미지정 => {}
	// {"enabled":false} => 꺼짐 => {"enabled":false}
	// {"enabled":true} => 켜짐 => {"enabled":true}
	// {"enabled":null} => 미지정 => {}
}
