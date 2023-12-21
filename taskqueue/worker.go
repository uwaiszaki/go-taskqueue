package taskqueue

import (
	"sync"
)

type WorkerGroup struct {
	WorkerPool sync.Pool
	taskQueue  *TaskQueue
}

func NewWorkerGroup(taskQueue *TaskQueue) *WorkerGroup {
	wg := &WorkerGroup{
		taskQueue: taskQueue,
	}
	wg.WorkerPool = sync.Pool{
		New: func() any {
			taskQueue.mu.Lock()
			taskQueue.TotalWorkers += 1
			worker := NewWorker(taskQueue.TotalWorkers, wg)
			// fmt.Printf("\nNew Worker [%v] Created\n", taskQueue.TotalWorkers)
			taskQueue.mu.Unlock()
			return worker
		},
	}
	return wg
}

func (wg *WorkerGroup) Start() {
	// This goroutine starts a new worker if no worker is running a task is submitted
	go func() {
		for {
			select {
			case <-wg.taskQueue.done:
				return
			default:
				wg.taskQueue.Cond.L.Lock()
				for !wg.taskQueue.ShouldRunWorker() {
					wg.taskQueue.Cond.Wait()
				}
				worker := wg.WorkerPool.Get()
				// fmt.Printf("\nStarting the worker[%v] after fetching from the pool\n", worker.(*Worker).id)
				wg.taskQueue.mu.Lock()
				wg.taskQueue.RunnableWorkers += 1
				wg.taskQueue.mu.Unlock()
				worker.(*Worker).Start()
				wg.taskQueue.Cond.L.Unlock()
			}
		}
	}()
}

type Worker struct {
	wg   *WorkerGroup
	id   int
	done chan struct{}
}

func NewWorker(id int, wg *WorkerGroup) *Worker {
	return &Worker{
		id:   id,
		wg:   wg,
		done: make(chan struct{}),
	}
}

func (w *Worker) Close() {
	close(w.done)
}

func (w *Worker) Start() {
	go func() {
		for {
			select {
			case task := <-w.wg.taskQueue.deque:
				if task == nil {
					return
				}
				w.wg.taskQueue.mu.Lock()
				w.wg.taskQueue.RunnableWorkers -= 1
				w.wg.taskQueue.mu.Unlock()
				task.Execute()
				// fmt.Printf("\nWorker[%v] Task Executed\n", w.id)
				if w.wg.taskQueue.IsQueueEmpty() {
					// When there is nothing left in the Queue
					w.wg.WorkerPool.Put(w)
					// fmt.Printf("\nReturning Worker [%v] to the pool\n, Running Workers = %v", w.id, w.wg.taskQueue.RunnableWorkers)
				}
				w.wg.taskQueue.mu.Lock()
				w.wg.taskQueue.RunnableWorkers += 1
				w.wg.taskQueue.mu.Unlock()
			case <-w.done:
				return
			case <-w.wg.taskQueue.done:
				return
			}
		}
	}()
}
