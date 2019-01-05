package pline

import ()

type Task interface {
	Kind() int
	Do() []Task
}

type hire struct {
	kind  int
	count int
}

type completion struct {
	tasks []Task
	kind  int
}

// A Line describes a production line
type Line struct {
	tasks    map[int][]Task
	idle     map[int]int
	index    map[int]int
	workers  map[int]int
	input    chan []Task
	complete chan completion
	hire     chan hire
	waiting  chan bool
	cancel   chan bool
	done     chan struct{}
	shrink   int
}

// Creates a new production line, the returned line needs to be started before
// it can be used.
func NewLine() (line *Line) {
	line = &Line{
		tasks:    make(map[int][]Task),
		idle:     make(map[int]int),
		index:    make(map[int]int),
		workers:  make(map[int]int),
		input:    make(chan []Task),
		complete: make(chan completion),
		hire:     make(chan hire),
		waiting:  make(chan bool),
		cancel:   make(chan bool),
		done:     make(chan struct{}),
		shrink:   128,
	}

	return
}

func (line *Line) main() {
	var (
		cancelled bool = false
		waiting   bool = false
		ignorant  bool = false
		empty     bool
		idle      bool
	)

	for {
		select {
		case graceful := <-line.cancel:
			if graceful {
				waiting = true
				ignorant = true
			} else {
				cancelled = true
			}
		case waiting = <-line.waiting:
		case hire := <-line.hire:
			if _, exists := line.workers[hire.kind]; exists {
				line.workers[hire.kind] += hire.count
				line.idle[hire.kind] += hire.count
			} else {
				line.workers[hire.kind] = hire.count
				line.idle[hire.kind] = hire.count
				line.index[hire.kind] = 0
			}
		case tasks := <-line.input:
			if ignorant {
				break
			}
			for _, task := range tasks {
				kind := task.Kind()
				line.tasks[kind] = append(line.tasks[kind], task)
			}
		case completion := <-line.complete:
			line.idle[completion.kind]++
			if ignorant {
				break
			}
			for _, task := range completion.tasks {
				kind := task.Kind()
				line.tasks[kind] = append(line.tasks[kind], task)
			}
		}

		if cancelled {
			break
		}

		idle = true
		for kind, _ := range line.workers {
			if line.workers[kind] != line.idle[kind] {
				idle = false
				break
			}
		}

		empty = true
		for kind, list := range line.tasks {
			if len(list)-line.index[kind] > 0 {
				empty = false
				break
			}
		}

		if waiting && idle && empty {
			break
		}

		for kind, list := range line.tasks {
			for ; line.idle[kind] > 0 && line.index[kind] < len(list); line.index[kind]++ {
				line.idle[kind]--
				go func(kind int, task Task) {
					line.complete <- completion{
						kind:  kind,
						tasks: task.Do(),
					}
				}(kind, list[line.index[kind]])
			}
		}

		for kind, _ := range line.tasks {
			if line.index[kind] >= line.shrink {
				line.tasks[kind] = line.tasks[kind][line.shrink:]
				line.index[kind] -= line.shrink
			}
		}

	}
	close(line.done)
}

// Starts the production line, the production line is now able
// to process requests to hire workers, push tasks, etc.
func (line *Line) Start() {
	go line.main()
}

// Hire increases the number of workers by count
// The kind of tasks the workers work on is specified by kind
// If count < 0, the number of workers are decremented
func (line *Line) Hire(kind int, count int) {
	line.hire <- hire{kind, count}
}

// Push tasks into the production line.
// Tasks are placed into the list respective of their kind.
func (line *Line) Push(tasks ...Task) {
	line.input <- tasks
}

// Wait blocks until the production line has no more tasks to complete
// and all workers are idle.
// The production line is still able to hire new workers.
func (line *Line) Wait() {
	line.waiting <- true
	<-line.done
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
