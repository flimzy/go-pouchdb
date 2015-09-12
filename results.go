package pouchdb

import (
	"github.com/gopherjs/gopherjs/js"
)

type pouchResultTuple struct {
	result	*js.Object
	err		*js.Object
}

type pouchResult struct {
	resultChan		chan *pouchResultTuple
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

func (pr *pouchResult) ReadResult() (Result,error) {
	result, err := pr.Read()
	return result.Interface().(map[string]interface{}), err
}

func (pr *pouchResult) Done(err *js.Object, result *js.Object) {
	pr.resultChan <- &pouchResultTuple{result, err}
}
