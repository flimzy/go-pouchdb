package pouchdb

import (
	"fmt"
	"testing"

	"github.com/gopherjs/gopherjs/js"
)

func BenchmarkConvertJSObject(b *testing.B) {
	jsObj := js.Global.Get("Object").New()
	jsObj.Set("foo", "bar")
	jsObj.Set("bar", 100)
	jsObj.Set("baz", true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x testObj
		if err := ConvertJSObject(jsObj, &x); err != nil {
			panic(err)
		}
	}
}

func BenchmarkConvertToJS(b *testing.B) {
	doc := map[string]interface{}{
		"foo": "bar",
		"bar": 100,
		"baz": "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x10\x11\x12\x13\x14\x15\x16\x17\x18",
	}
	for i := 0; i < b.N; i++ {
		if _, err := convertToJS(doc); err != nil {
			panic(err)
		}
	}
}

func BenchmarkCreateDoc(b *testing.B) {
	db := newPouch("BenchmarkCreateDoc")
	doc := map[string]interface{}{
		"foo": "bar",
		"bar": 100,
		"baz": "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x10\x11\x12\x13\x14\x15\x16\x17\x18",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc["_id"] = fmt.Sprintf("%d", i)
		if _, err := db.Put(doc); err != nil {
			panic(err)
		}
	}
}
