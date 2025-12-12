package main

import (
	"sync"
)

type Bank struct {
	mu       sync.Mutex
	accounts map[string]int64
}
