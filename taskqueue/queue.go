package taskqueue

import (
	"sync"
	"time"
)

type Task interface {
	Execute()
}

type TaskQueue struct {
	enque chan Task
	deque chan Task
	done  chan struct{}
	Cond  *sync.Cond

	mu              *sync.RWMutex
	queue           []Task
	TotalWorkers    int
	RunnableWorkers int
	MaxWorkers      int
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		queue: make([]Task, 0),
		enque: make(chan Task),
		deque: make(chan Task),
		done:  make(chan struct{}),
		Cond:  sync.NewCond(&sync.Mutex{}),
		mu:    &sync.RWMutex{},
	}
}

func (q *TaskQueue) IsQueueEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.queue) == 0
}

func (q *TaskQueue) GetQueue() []Task {
	return q.queue
}

func (q *TaskQueue) AddTask(tasks ...Task) {
	for _, task := range tasks {
		if task != nil {
			q.enque <- task
		}
	}
}

func (q *TaskQueue) Close() {
	close(q.done)
}

func (q *TaskQueue) ShouldRunWorker() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.RunnableWorkers == 0 && len(q.queue) > 0
}

func (q *TaskQueue) Start() {
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		for {
			if len(q.queue) > 0 {
				select {
				case <-q.done:
					return
				case q.deque <- q.queue[0]:
					q.mu.Lock()
					if len(q.queue)%10 == 0 {
						newSlice := make([]Task, len(q.queue))
						copy(newSlice, q.queue[1:])
						q.queue = newSlice
					} else {
						q.queue = q.queue[1:]
					}
					q.mu.Unlock()
				case task := <-q.enque:
					q.mu.Lock()
					q.queue = append(q.queue, task)
					q.mu.Unlock()
				case <-ticker.C:
					shouldRunWorker := q.ShouldRunWorker()
					// fmt.Println("\nChecking if should wake up the manager", wakeUpWorkerManager)
					if shouldRunWorker {
						// This will inform the WorkerGroup to get new workers
						// fmt.Println("Waking up the Worker Manager")
						q.Cond.L.Lock()
						q.Cond.Signal()
						q.Cond.L.Unlock()
					}
				}
			} else {
				select {
				case <-q.done:
					return
				case task := <-q.enque:
					q.mu.Lock()
					q.queue = append(q.queue, task)
					q.mu.Unlock()
				}
			}
		}
	}()
}
