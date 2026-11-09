package main

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestMergeValuesAndClosure(t *testing.T) {
	a := make(chan int, 2)
	a <- 2
	a <- 2
	close(a)
	b := make(chan int, 1)
	b <- 1
	close(b)
	out, done := merge(context.Background(), a, nil, b)
	values := []int{}
	for value := range out {
		values = append(values, value)
	}
	<-done
	sort.Ints(values)
	if !reflect.DeepEqual(values, []int{1, 2, 2}) {
		t.Fatal(values)
	}
}
func TestMergeCancellationWithIdleInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out, done := merge(ctx, make(chan int))
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("workers did not join")
	}
	if _, ok := <-out; ok {
		t.Fatal("unexpected value")
	}
}
func TestMergeNoInputs(t *testing.T) {
	out, done := merge(context.Background())
	<-done
	if _, ok := <-out; ok {
		t.Fatal("not closed")
	}
}
