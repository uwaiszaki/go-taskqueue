package main

import (
	"fmt"
	"time"

	"github.com/uwaiszaki/go-taskqueue/taskqueue"
)

type TaskA struct {
	message string
}

func (task TaskA) Execute() {
	time.Sleep(time.Second)
	fmt.Printf("\nExecuting TaskA Message = %v\n", task.message)
}

type TaskB struct {
	message string
}

func (task TaskB) Execute() {
	time.Sleep(time.Second)
	fmt.Printf("\nExecuting TaskB Message = %v\n", task.message)
}

func main() {
	taskQueue := taskqueue.NewTaskQueue()
	taskQueue.Start()
	defer taskQueue.Close()
	workerGroup := taskqueue.NewWorkerGroup(taskQueue)
	workerGroup.Start()
	taskA := TaskA{
		message: "[Task A Message]",
	}
	taskB := TaskB{
		message: "[Task B Message]",
	}
	taskQueue.AddTask(taskA)
	taskQueue.AddTask(taskB)
	timer := time.NewTicker(2 * time.Second)
	<-timer.C
	fmt.Println(taskQueue.GetQueue(), taskQueue.RunnableWorkers, taskQueue.TotalWorkers)
}
