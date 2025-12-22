package main

import (
	"sort"
	"sync"
)

type LockManager struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func (lm *LockManager) getLock(account string) *sync.Mutex {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, exists := lm.locks[account]; !exists {
		lm.locks[account] = &sync.Mutex{}
	}
	return lm.locks[account]
}

func (lm *LockManager) LockKeys(accounts ...string) {
	sort.Strings(accounts)

	for _, account := range accounts {
		l := lm.getLock(account)
		l.Lock()
	}
}

func (lm *LockManager) UnlockKeys(accounts ...string) {
	sort.Strings(accounts)
	for _, account := range accounts {
		l := lm.getLock(account)
		l.Unlock()
	}
}
