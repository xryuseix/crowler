package main

import (
	"fmt"
	"sync"
)

var THREAD_MAX = 8

type Thread struct {
	mu   sync.Mutex
	left int
}

func (t *Thread) Inc() error {
	t.mu.Lock()
	if t.left == THREAD_MAX {
		t.mu.Unlock()
		return fmt.Errorf("no thread left")
	}
	t.left++
	t.mu.Unlock()
	return nil
}
func (t *Thread) Dec() error {
	t.mu.Lock()
	if t.left == 0 {
		t.mu.Unlock()
		return fmt.Errorf("no thread left")
	}
	t.left--
	t.mu.Unlock()
	return nil
}
