package cfgmngr

import (
	"reflect"
	"testing"
)

func TestGetAction(t *testing.T) {
	var tests = []struct {
		args    []string
		action  string
		newArgs []string
	}{
		{[]string{"test/test.bin"}, "", nil},
		{[]string{"test/test.bin", "-v"}, "", nil},
		{[]string{"test/test.bin", "--help"}, "", nil},
		{[]string{"test/test.bin", "/version"}, "", nil},
		{[]string{"test/test.bin", "/h"}, "", nil},
		{[]string{"test/test.bin", "open"}, "open", nil},
		{[]string{"test/test.bin", "open", "-f"}, "open", []string{"test/test.bin", "-f"}},
		{[]string{"test/test.bin", "put", "--file=test.txt", "--filter pat:123", "-d"}, "put", []string{"test/test.bin", "--file=test.txt", "--filter pat:123", "-d"}},
	}

	for i, tt := range tests {
		got := getAction(&tt.args)
		if got != tt.action {
			t.Errorf("\n%d. feed: %s -> expected action: %s - got: %s", i, tt.args, tt.action, got)
		}
		if len(tt.args) > 2 {
			if reflect.DeepEqual(tt.args, tt.newArgs) {
				t.Errorf("\n%d. expected args: %s - got: %s", i, tt.newArgs, tt.args)
			}
		}
	}

}
