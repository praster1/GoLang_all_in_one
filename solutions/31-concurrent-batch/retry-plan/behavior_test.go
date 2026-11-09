package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func TestRetryInputsSelectOnlyRetryableFailures(t *testing.T) {
	results := []batch.Result{{Job: batch.Job{Input: "b"}, State: batch.Failed, Err: batch.ErrTemporary}, {Job: batch.Job{Input: "a"}, State: batch.Failed, Err: fmt.Errorf("wrapped: %w", batch.ErrTemporary)}, {Job: batch.Job{Input: "bad"}, State: batch.Failed, Err: errors.New("bad")}, {Job: batch.Job{Input: "stopped"}, State: batch.Canceled, Err: context.Canceled}, {Job: batch.Job{Input: "ok"}, State: batch.Succeeded}}
	if got := retryInputs(results); !reflect.DeepEqual(got, []string{"b", "a"}) {
		t.Fatal(got)
	}
	if got := retryInputs(nil); got == nil || len(got) != 0 {
		t.Fatal(got)
	}
}
