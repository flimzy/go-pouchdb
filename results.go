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
func (rw *resultWaiter) Read() (*js.Object, error) {
	rawResult := <-rw.resultChan
	if rawResult.err == nil {
		return rawResult.result, nil
	}
	return rawResult.result, &js.Error{rawResult.err}
}

// Error returns just the error of a PouchDB callback, for methods
// which don't need the result value
func (rw *resultWaiter) Error() error {
	_, err := rw.Read()
	return err
}

func (rw *resultWaiter) ReadRev() (string, error) {
	obj, err := rw.ReadResult()
	if err != nil {
		return "", err
	}
	return obj["rev"].(string), nil
}

func (rw *resultWaiter) ReadResult() (Result, error) {
	result, err := rw.Read()
	if err != nil {
		return Result{}, err
	}
	return result.Interface().(map[string]interface{}), err
}

func (rw *resultWaiter) ReadBulkResults() ([]Result, error) {
	result, err := rw.Read()
	results := make([]Result, result.Length())
	for i := 0; i < result.Length(); i++ {
		results[i] = result.Index(i).Interface().(map[string]interface{})
	}
	return results, err
}

func (rw *resultWaiter) Done(err *js.Object, result *js.Object) {
	rw.resultChan <- &resultWaiterTuple{result, err}
}
