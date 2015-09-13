// +build js

package pouchdb

import (
	"github.com/gopherjs/gopherjs/js"
)

type resultWaiterTuple struct {
	result *js.Object
	err    *js.Object
}

type resultWaiter struct {
	resultChan chan *resultWaiterTuple
}

func newResultWaiter() *resultWaiter {
	return &resultWaiter{
		make(chan *resultWaiterTuple),
	}
}

// Read returns the raw results of a PouchDB callback
func (pr *resultWaiter) Read() (*js.Object, error) {
	rawResult := <-pr.resultChan
	if rawResult.err == nil {
		return rawResult.result, nil
	}
	return rawResult.result, &js.Error{rawResult.err}
}

// Error returns just the error of a PouchDB callback, for methods
// which don't need the result value
func (pr *resultWaiter) Error() error {
	_, err := pr.Read()
	return err
}

func (pr *resultWaiter) ReadResult() (Result, error) {
	result, err := pr.Read()
	if err != nil {
		return Result{}, err
	}
	return result.Interface().(map[string]interface{}), err
}

func (pr *resultWaiter) ReadBulkResults() ([]Result, error) {
	result, err := pr.Read()
	results := make([]Result, result.Length())
	for i := 0; i < result.Length(); i++ {
		results[i] = result.Index(i).Interface().(map[string]interface{})
	}
	return results, err
}

func (pr *resultWaiter) Done(err *js.Object, result *js.Object) {
	pr.resultChan <- &resultWaiterTuple{result, err}
}
