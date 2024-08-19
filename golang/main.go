package main

import (
	"sync"
	"task_generator/processors"
	"task_generator/tasks"
	"time"
)

func main() {
	superChan := make(chan tasks.Task, 10)
	doneTasks := make(chan tasks.Task)
	undoneTasks := make(chan error)
	stopChan := make(chan struct{})

	var mu sync.Mutex
	var wg sync.WaitGroup

	successCount := 0
	failCount := 0
	idTasks := 0

	result := map[int]tasks.Task{}
	errors := []error{}

	countGenerators := 3

	processorParams := processors.ProcessorParams{
		Chan:         superChan,
		Mu:           &mu,
		SuccessCount: &successCount,
		FailCount:    &failCount,
		DoneTasks:    doneTasks,
		UndoneTasks:  undoneTasks,
		Wg:           &wg,
		Result:       result,
		Errors:       errors,
	}

	processor := processors.NewTaskProcessor(processorParams)

	for i := 0; i < countGenerators; i++ {
		go tasks.CreateTasks(superChan, stopChan, &idTasks, &mu)
	}

	go processor.Process()
	go processors.PeriodicStatusOutput(&mu, &successCount, &failCount)

	go processor.Sort()

	// Работа приложения 10 секунд
	time.Sleep(10 * time.Second)

	close(stopChan)
	close(superChan)
	wg.Wait()
	close(doneTasks)
	close(undoneTasks)

	processor.PrintFinalResult()
}
