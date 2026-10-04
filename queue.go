package main

import "sync"

type serialQueue struct {
	mutex sync.Mutex
}

func (queue *serialQueue) run(job func() error) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	return job()
}

var printerQueue serialQueue
