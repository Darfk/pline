package main

import (
	pline ".."
	"time"
)

const (
	TaskGreet = iota
	TaskPrint
)

type GreetTask struct {
	name string
}

func (t *GreetTask) Kind() int { return TaskGreet }
func (t *GreetTask) Do() []pline.Task {
	if t.name == "Andy" || t.name == "Nathan" {
		// say hi to our friends
		return []pline.Task{&PrintTask{"Hey " + t.name + "!"}}
	} else if t.name == "Kirk" {
		// give Kirk the cold shoulder
		return nil
	}
	// we haven't met them before
	return []pline.Task{
		&PrintTask{"Hi " + t.name},
		&PrintTask{"Welcome to pline " + t.name + "."},
	}
}

type PrintTask struct {
	text string
}

func (t *PrintTask) Kind() int { return TaskPrint }
func (t *PrintTask) Do() []pline.Task {
	// printing takes a while, we will use a worker pool
	time.Sleep(time.Millisecond * 250)
	println(t.text)
	return nil
}

func main() {
	line := pline.NewLine()

	// start the line so that it can accept tasks
	line.Start()

	// push all the tasks into the production line
	line.Push([]pline.Task{
		&PrintTask{"This is pline"},
		&GreetTask{"Andy"},
		&GreetTask{"Nathan"},
		&GreetTask{"Kirk"},
		&GreetTask{"Rebecca"},
		&GreetTask{"Miles"},
		&GreetTask{"Emily"},
		&GreetTask{"Steven"},
	})

	// hire a worker to create greetings
	line.Hire(TaskGreet, 1)

	// hire some workers to print the greetings
	line.Hire(TaskPrint, 8)

	// stop and wait for the line to finish all of the tasks
	line.Wait()
}
