package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

type Config struct{ Workers int }
type Loader struct {
	once  sync.Once
	load  func() (Config, error)
	value Config
	err   error
}

func newLoader(load func() (Config, error)) (*Loader, error) {
	if load == nil {
		return nil, errors.New("읽기 함수가 필요합니다")
	}
	return &Loader{load: load}, nil
}
func (l *Loader) Get() (Config, error) {
	l.once.Do(func() {
		l.value, l.err = l.load()
		if l.err == nil && (l.value.Workers < 1 || l.value.Workers > 32) {
			l.err = errors.New("작업자 수는 1~32여야 합니다")
		}
		if l.err != nil {
			l.value = Config{}
		}
	})
	return l.value, l.err
}
func main() {
	calls := 0
	sourceError := errors.New("설정을 읽을 수 없습니다")
	loader, err := newLoader(func() (Config, error) { calls++; return Config{}, sourceError })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	failures := make([]error, 3)
	var group sync.WaitGroup
	for index := range failures {
		group.Go(func() { _, failures[index] = loader.Get() })
	}
	group.Wait()
	fmt.Println("읽기 시도:", calls)
	fmt.Println("같은 오류 보존:", errors.Is(failures[0], sourceError) && errors.Is(failures[1], sourceError) && errors.Is(failures[2], sourceError))
}
