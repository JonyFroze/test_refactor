package processors

import (
	"fmt"

	"sync"
	"task_generator/tasks"
	"time"
)

// TaskProcessor defines the processor for tasks
type TaskProcessor struct {
	taskChan     chan tasks.Task
	wg           *sync.WaitGroup
	mu           *sync.Mutex
	successCount *int
	failCount    *int
	doneTasks    chan tasks.Task
	undoneTasks  chan error
	result       map[int]tasks.Task
	errors       []error
}

type ProcessorParams struct {
	Chan         chan tasks.Task
	Wg           *sync.WaitGroup
	Mu           *sync.Mutex
	SuccessCount *int
	FailCount    *int
	DoneTasks    chan tasks.Task
	UndoneTasks  chan error
	Result       map[int]tasks.Task
	Errors       []error
}

// NewTaskProcessor creates a new TaskProcessor
func NewTaskProcessor(Params ProcessorParams) *TaskProcessor {
	return &TaskProcessor{
		taskChan:     Params.Chan,
		wg:           Params.Wg,
		mu:           Params.Mu,
		successCount: Params.SuccessCount,
		failCount:    Params.FailCount,
		doneTasks:    Params.DoneTasks,
		undoneTasks:  Params.UndoneTasks,
		result:       Params.Result,
		errors:       Params.Errors,
	}
}

// Process handles task processing
// Process handles task processing
func (p *TaskProcessor) Process() {
	for task := range p.taskChan {
		p.wg.Add(1)
		go p.processTask(task)
	}
}

func (p *TaskProcessor) Sort() {

	go func() {
		for task := range p.doneTasks {
			//fmt.Println("done task", task)
			p.mu.Lock()
			p.result[task.ID] = task
			p.mu.Unlock()
			//fmt.Println("Result", result)
		}
	}()
	go func() {
		for err := range p.undoneTasks {
			//fmt.Println("Error", err)
			p.mu.Lock()
			p.errors = append(p.errors, err)
			p.mu.Unlock()
			//fmt.Println("Errors", errors)
		}
	}()
}

func (p *TaskProcessor) processTask(task tasks.Task) {

	defer p.wg.Done()
	CreateTime, err := time.Parse(time.RFC3339Nano, task.CreatedAt)
	if err != nil {
		fmt.Println("Error parsing created at time", err)
		return
	}

	if CreateTime.After(time.Now().Add(-20 * time.Microsecond)) {

		time.Sleep(150 * time.Millisecond)
		task.Result = []byte("task has been successed")

		if CreateTime.Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
			task.Result = []byte("some error occured")

			p.mu.Lock()
			(*p.failCount)++
			p.mu.Unlock()

			task.ProcessedAt = time.Now().Format(time.RFC3339Nano)
			p.undoneTasks <- fmt.Errorf("Task ID: %d time %s, error %s", task.ID, task.CreatedAt, task.Result)
		} else {
			p.mu.Lock()
			(*p.successCount)++
			p.mu.Unlock()

			task.ProcessedAt = time.Now().Format(time.RFC3339Nano)
			p.doneTasks <- task
		}

	} else {
		task.Result = []byte("something went wrong")

		p.mu.Lock()
		(*p.failCount)++
		p.mu.Unlock()

		task.ProcessedAt = time.Now().Format(time.RFC3339Nano)
		p.undoneTasks <- fmt.Errorf("Task ID: %d time %s, error %s", task.ID, task.CreatedAt, task.Result)
	}

}

func (p *TaskProcessor) PrintFinalResult() {

	fmt.Println("Final counts:")
	p.mu.Lock()
	fmt.Printf("Successful tasks: %d, Failed tasks: %d\n", *p.successCount, *p.failCount)
	p.mu.Unlock()

	fmt.Println("Done tasks:")
	for key, task := range p.result {
		fmt.Printf("Task ID: %d, Created At: %s, Processed At: %s, Result: %s\n",
			key, task.CreatedAt, task.ProcessedAt, task.Result)
	}

	fmt.Println("Undone tasks:")
	for _, err := range p.errors {
		fmt.Printf("Errors: %s \n", err.Error())
	}
}

// PeriodicStatusOutput periodically outputs the number of successful and failed tasks

func PeriodicStatusOutput(mu *sync.Mutex, successCount, failCount *int) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 3; i++ {
		<-ticker.C
		mu.Lock()
		fmt.Printf("Successful tasks: %d, Failed tasks: %d\n", *successCount, *failCount)
		mu.Unlock()
	}
}

// FinalizeTasks finalizes remaining tasks as failed
