package main

import (
	"fmt"
	"sort"
	"sync"
)

type Store struct {
	mu     sync.RWMutex
	values map[string]string
}

func (s *Store) Put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.values == nil {
		s.values = make(map[string]string)
	}
	s.values[key] = value
}
func (s *Store) Snapshot() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]string, len(s.values))
	for key, value := range s.values {
		result[key] = value
	}
	return result
}
func main() {
	var store Store
	var group sync.WaitGroup
	group.Go(func() { store.Put("language", "Go") })
	group.Go(func() { store.Put("topic", "동기화") })
	group.Wait()
	snapshot := store.Snapshot()
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s=%s\n", key, snapshot[key])
	}
	snapshot["language"] = "changed"
	fmt.Println("원본 유지:", store.Snapshot()["language"])
}
