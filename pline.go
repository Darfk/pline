package pline

import (
	"context"
)

type Task interface {
	Run(context.Context)
}

type hire struct {
	group *Group
	count int
}

type completion struct {
	group *Group
}

type Result struct {
	load    int
	workers int
}

type lock struct {
	group *Group
}

type input struct {
	group *Group
	task  Task
}

// A Line describes a production line full of workers
type Line struct {
	input    chan input
	result   chan Result
	complete chan completion
	hire     chan hire
	wait     chan bool
	lock     chan lock
	ctx      context.Context
	cancel   context.CancelFunc
	load     int
	waiting  bool
	ignorant bool
}

// Creates a new production line, the returned line needs to be started before
// it can be used.
func NewLine() (line *Line) {
	line = &Line{
		input:    make(chan input),
		result:   make(chan Result),
		complete: make(chan completion),
		hire:     make(chan hire),
		wait:     make(chan bool),
		lock:     make(chan lock),
		waiting:  false,
		load:     0,
	}

	return
}

type Group struct {
	line    *Line
	tasks   []Task
	idle    int
	workers int
}

func (line *Line) NewGroup(workers int) (group *Group) {
	group = &Group{
		workers: workers,
		idle:    workers,
		line:    line,
	}

	return
}

func (line *Line) main() {
	line.ctx, line.cancel = context.WithCancel(line.ctx)

	for {
		var group *Group = nil

		select {
		case <-line.ctx.Done():
			return
		case ignorant := <-line.wait:
			line.waiting = true
			line.ignorant = ignorant
		case hire := <-line.hire:
			group = hire.group
			group.idle += hire.count
			line.result <- Result{
				workers: group.workers,
				load:    len(group.tasks),
			}
		case lock := <-line.lock:
			group = lock.group
			line.result <- Result{
				workers: lock.group.workers,
				load:    len(group.tasks),
			}
		case input := <-line.input:
			group = input.group
			if !line.ignorant {
				line.load++
				group.tasks = append(group.tasks, input.task)
			}
			line.result <- Result{
				workers: group.workers,
				load:    len(group.tasks),
			}
		case completion := <-line.complete:
			group = completion.group
			line.load--
			group.idle++
		}

		if group != nil {
			for group.idle > 0 && len(group.tasks) > 0 {
				group.idle--
				go func(task Task, group *Group) {
					task.Run(line.ctx)
					line.complete <- completion{group: group}
				}(group.tasks[0], group)
				group.tasks = group.tasks[1:]
			}
		}

		if line.load == 0 && line.waiting {
			line.cancel()
		}
	}
}

// Push a task into this group.
// This function returns when the new task has been added and accounted for.
func (group *Group) Push(task Task) Result {
	group.line.input <- input{group: group, task: task}
	return <-group.line.result
}

// Walk through the group's tasks calling fn for each task
func (group *Group) WalkTasks(fn func(int, Task)) Result {
	group.line.lock <- lock{group: group}
	for i, task := range group.tasks {
		fn(i, task)
	}
	return <-group.line.result
}

// Starts the production line, the production line is now able
// to process requests to hire workers, push tasks, etc.
func (line *Line) Start(ctx context.Context) {
	line.ctx = ctx
	go line.main()
}

// Hire increases the number of workers by count
// The kind of tasks the workers work on is specified by kind
// If count < 0, the number of workers are decremented
func (group *Group) Hire(count int) Result {
	group.line.hire <- hire{group: group, count: count}
	return <-group.line.result
}

// Wait blocks until the production line has no more tasks to complete
// and all workers are idle.
// The production line is still able to hire new workers.
func (line *Line) Wait() {
	line.wait <- false
	<-line.ctx.Done()
}

// Finish is the same as Wait except that the production line is
// set to ignore all tasks placed into it via Push.
// The production line is still able to hire new workers.
func (line *Line) Finish() {
	line.wait <- true
	<-line.ctx.Done()
}

// Cancel immediately ends the production line and returns,
// further calls to the production line will block forever.
func (line *Line) Cancel() {
	line.cancel()
}

// Report captures information about a production line and returns a reference to a Report.
func (line *Line) Report() (r *Report) {
	r = &Report{
		Workers:  make(map[int]int),
		Idle:     make(map[int]int),
		Tasks:    make(map[int]int),
		Index:    make(map[int]int),
		Waiting:  false,
		Ignorant: false,
	}

	line.report <- r
	return <-line.report
}
