package utils

import "sync"

var cleanupsLock sync.Mutex //for possible future multi-threaded use
var scheduledCleanups []func() = []func(){}

func DeferGlobalCleanup(cleanup func()) {
	cleanupsLock.Lock()
	defer cleanupsLock.Unlock()
	scheduledCleanups = append(scheduledCleanups, cleanup)
}

func RunGlobalCleanup() {
	cleanupsLock.Lock()
	defer cleanupsLock.Unlock()
	currentCleanups := scheduledCleanups
	scheduledCleanups = []func(){}
	//run cleanup func
	for _, cleanup := range currentCleanups {
		cleanup()
	}
}
