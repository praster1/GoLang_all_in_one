package main

import (
	"errors"
	"fmt"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func retryInputs(results []batch.Result) []string {
	inputs := []string{}
	for _, result := range results {
		if result.State == batch.Failed && errors.Is(result.Err, batch.ErrTemporary) {
			inputs = append(inputs, result.Job.Input)
		}
	}
	return inputs
}
func main() {
	results := []batch.Result{
		{Job: batch.Job{Index: 0, Input: "ok"}, State: batch.Succeeded},
		{Job: batch.Job{Index: 1, Input: "busy"}, State: batch.Failed, Err: fmt.Errorf("server busy: %w", batch.ErrTemporary)},
		{Job: batch.Job{Index: 2, Input: "invalid"}, State: batch.Failed, Err: errors.New("invalid input")},
		{Job: batch.Job{Index: 3, Input: "canceled"}, State: batch.Canceled},
	}
	fmt.Println("검토할 재시도 입력:", retryInputs(results))
}
