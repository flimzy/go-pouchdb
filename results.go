package pouchdb

import (
	"github.com/gopherjs/gopherjs/js"
	// 	"honnef.co/go/js/console"
)

type pouchResultTuple struct {
	result *js.Object
	err    *js.Object
}

type pouchResult struct {
	resultChan chan *pouchResultTuple
}

func newResult() *pouchResult {
	return &pouchResult{
		make(chan *pouchResultTuple),
	}
}

func (pr *pouchResult) Read() (*js.Object, error) {
	rawResult := <-pr.resultChan
	if rawResult.err == nil {
		return rawResult.result, nil
	}
	return rawResult.result, &js.Error{rawResult.err}
}

func (pr *pouchResult) ReadResult() (Result, error) {
	result, err := pr.Read()
	if err != nil {
		return Result{}, err
	}
	return result.Interface().(map[string]interface{}), err
}

func (pr *pouchResult) ReadBulkResults() ([]Result, error) {
	result, err := pr.Read()
	results := make([]Result, result.Length())
	for i := 0; i < result.Length(); i++ {
		results[i] = result.Index(i).Interface().(map[string]interface{})
	}
	return results, err
}

func (pr *pouchResult) Done(err *js.Object, result *js.Object) {
	pr.resultChan <- &pouchResultTuple{result, err}
}
