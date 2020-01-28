package pline

import (
	"sync"
	"context"
)

var defaultWaitGroup *sync.WaitGroup

type Task interface {
	Run(context.Context)
}

type hire struct {
	count int
}

type completion struct{}

// A Line describes a production line
type Line struct {
	tasks    []Task
	idle     int
	index    int
	workers  int
	input    chan Task
	complete chan completion
	hire     chan hire
	waiting  chan bool
	cancel   chan bool
	done     chan struct{}
	wg       *sync.WaitGroup
}

// Creates a new production line, the returned line needs to be started before
// it can be used.
func NewLine() (line *Line) {
	line = &Line{
		input:    make(chan Task),
		complete: make(chan completion),
		hire:     make(chan hire),
		waiting:  make(chan bool),
		cancel:   make(chan bool),
		done:     make(chan struct{}),
	}

	if defaultWaitGroup == nil {
		defaultWaitGroup = &sync.WaitGroup{}
	}

	line.wg = defaultWaitGroup

	return
}

func NewLineWithWaitGroup(wg *sync.WaitGroup) (line *Line) {
	line = NewLine()
	line.wg = wg
	return
}

// TODO: use package context
func (line *Line) main(ctx context.Context) {
	var (
		cancelled bool = false
		ignorant  bool = false
		idle      int
	)

	for {
		select {
		case graceful := <-line.cancel:
			if graceful {
				ignorant = true
			} else {
				cancelled = true
			}
		// case waiting = <-line.waiting:
		case hire := <-line.hire:
			idle += hire.count
		case task := <-line.input:
			if ignorant {
				break
			}
			line.tasks = append(line.tasks, task)
			line.wg.Add(1)
		case _ = <-line.complete:
			idle++
			line.wg.Done()
		}

		if cancelled {
			break
		}

		for idle > 0 && len(line.tasks) > 0 {
			idle--
			go func(task Task) {
				task.Run(ctx)
				line.complete <- completion{}
			}(line.tasks[0])
			line.tasks = line.tasks[1:]
		}
	}
	close(line.done)
}

// Starts the production line, the production line is now able
// to process requests to hire workers, push tasks, etc.
func (line *Line) Start(ctx context.Context) {
	go line.main(ctx)
}

// Hire increases the number of workers by count
// The kind of tasks the workers work on is specified by kind
// If count < 0, the number of workers are decremented
func (line *Line) Hire(count int) {
	line.hire <- hire{count}
}

// Push tasks into this production line.
func (line *Line) Push(tasks Task) {
	line.input <- tasks
}

// Wait blocks until the production line has no more tasks to complete
// and all workers are idle.
// The production line is still able to hire new workers.
func (line *Line) Wait() {
	// line.waiting <- true
	line.wg.Wait()
}

// Finish is the same as Wait except that the production line is
// set to ignore all tasks placed into it either from the result of
// complete tasks or via use of Push.
// The production line is still able to hire new workers.
func (line *Line) Finish() {
	line.cancel <- true
	<-line.done
}

// Cancel immediately ends the production line and returns,
// further calls to the production line will block forever.
func (line *Line) Cancel() {
	line.cancel <- false
	<-line.done
}
