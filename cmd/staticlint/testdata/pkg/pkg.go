package main

import (
	"sync"
)

func main() {
	// Для copylock: копирование мьютекса
	var mu sync.Mutex
	muCopy := mu // want "assignment copies lock value to muCopy: sync.Mutex"
	muCopy.Lock()
}
