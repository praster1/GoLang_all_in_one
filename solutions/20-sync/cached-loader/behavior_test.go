package main

import (
	"errors"
	"sync"
	"testing"
)

func TestLoaderRunsOnceAndCachesFailure(t *testing.T) {
	calls := 0
	want := errors.New("source failed")
	loader, err := newLoader(func() (Config, error) { calls++; return Config{Workers: 4}, want })
	if err != nil {
		t.Fatal(err)
	}
	values := make([]Config, 16)
	failures := make([]error, 16)
	var group sync.WaitGroup
	for index := range values {
		group.Go(func() { values[index], failures[index] = loader.Get() })
	}
	group.Wait()
	if calls != 1 {
		t.Fatal(calls)
	}
	for i := range values {
		if values[i] != (Config{}) || !errors.Is(failures[i], want) {
			t.Fatal(values[i], failures[i])
		}
	}
}
func TestLoaderValidation(t *testing.T) {
	if value, err := newLoader(nil); value != nil || err == nil {
		t.Fatal(value, err)
	}
	for _, workers := range []int{0, 33} {
		loader, err := newLoader(func() (Config, error) { return Config{Workers: workers}, nil })
		if err != nil {
			t.Fatal(err)
		}
		if value, err := loader.Get(); value != (Config{}) || err == nil {
			t.Fatal(value, err)
		}
	}
	loader, err := newLoader(func() (Config, error) { return Config{Workers: 4}, nil })
	if err != nil {
		t.Fatal(err)
	}
	if value, err := loader.Get(); err != nil || value.Workers != 4 {
		t.Fatal(value, err)
	}
}
