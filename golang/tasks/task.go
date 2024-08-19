package tasks

import (
	"sync"
	"time"
)

// Task represents an abstract task
type Task struct {
	ID          int
	CreatedAt   string
	ProcessedAt string
	Result      []byte
}

// CreateTasks generates tasks and sends them to the channel every 150ms
func CreateTasks(taskChan chan Task, stopChan chan struct{}, idCount *int, mu *sync.Mutex) {

	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			createdAt := time.Now().Format(time.RFC3339Nano)

			task := Task{
				ID:        *idCount,
				CreatedAt: createdAt,
			}
			taskChan <- task
			mu.Lock()
			(*idCount)++
			mu.Unlock()
		}
	}

}
