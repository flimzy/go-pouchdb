package pouchdb

import (
	"reflect"
	"testing"
)

func TestOptions(t *testing.T) {
	opts := Options{}
	expected := make(map[string]interface{})
	compiled := opts.compile()
	if !reflect.DeepEqual(compiled, expected) {
		t.Fatalf("Got: %v, Expected: %v", compiled, expected)
	}
	opts.Limit = 10
	expected["limit"] = 10
	compiled = opts.compile()
	if !reflect.DeepEqual(compiled, expected) {
		t.Fatalf("Got: %v, Expected: %v", compiled, expected)
	}
}
