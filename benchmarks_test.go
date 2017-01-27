package pouchdb

import (
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
