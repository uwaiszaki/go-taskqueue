package taskqueue

import (
	"testing"
	"time"
)

type TestTask struct {
	id int
}

func (t TestTask) Execute() {
	time.Sleep(100 * time.Millisecond)
	// fmt.Printf("Executing Task [%v] Message [%v]", t.id, fmt.Sprintf("Task%v Message Executed", t.id))
}

func getTask(id int) TestTask {
	return TestTask{
		id: id,
	}
}

func TestQueue(t *testing.T) {
	// Start the Queue, Init the Worker Group, push 2 tasks, number of total workers should be 2
	taskQueue := NewTaskQueue()
	taskQueue.Start()
	defer taskQueue.Close()

	task1, task2, task3 := getTask(1), getTask(2), getTask(3)
	taskQueue.AddTask(task1, task2, task3)
	wg := NewWorkerGroup(taskQueue)
	wg.Start()

	timer := time.NewTimer(5 * time.Second)
	<-timer.C

	taskQueue.mu.RLock()
	if taskQueue.TotalWorkers != 3 {
		t.Error("Total Workers Should be = 3, but it's = ", taskQueue.TotalWorkers)
	}
	taskQueue.mu.RUnlock()
}

func BenchmarkQueue(b *testing.B) {
	tq := NewTaskQueue()
	tq.Start()
	defer tq.Close()
	wg := NewWorkerGroup(tq)
	wg.Start()
	for i := 0; i < b.N; i++ {
		tasks := make([]Task, i+1)
		for j := 0; j <= i; j++ {
			tasks = append(tasks, getTask(j))
		}
		tq.AddTask(tasks...)
	}
}
