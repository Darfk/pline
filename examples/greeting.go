package main

import (
	pline ".."
	"log"
	"time"
	"context"
)

var (
	printLine *pline.Line
	greetLine *pline.Line
)

type GreetTask struct {
	name string
}

func (t *GreetTask) Run(ctx context.Context) {
	log.Println("greet task:", t.name)

	time.Sleep(time.Millisecond * 250)

	if t.name == "Andy" || t.name == "Nathan" {
		// say hi to our friends
		printLine.Push(&PrintTask{"Hey " + t.name + "!"})
		return
	} else if t.name == "Kirk" {
		// give Kirk the cold shoulder
		return
	}
	// we haven't met them before

	printLine.Push(&PrintTask{"Hi " + t.name})
	printLine.Push(&PrintTask{"Welcome to pline " + t.name + "."})
}

type PrintTask struct {
	text string
}

func (t *PrintTask) Run(ctx context.Context) {
	// printing takes a while, we will use a worker pool
	time.Sleep(time.Millisecond * 250)
	log.Println("print task:", t.text)
	return
}

func main() {
	// create the production lines
	greetLine = pline.NewLine()
	printLine = pline.NewLine()

	ctx := context.Background()

	// start the lines so that they can accept tasks
	greetLine.Start(ctx)
	printLine.Start(ctx)

	// push all the print tasks into the production line
	printLine.Push(&PrintTask{"This is pline"})

	// push all the greet tasks into the production line
	greetLine.Push(&GreetTask{"Andy"})
	// greetLine.Push(&GreetTask{"Nathan"})
	// greetLine.Push(&GreetTask{"Kirk"})
	// greetLine.Push(&GreetTask{"Rebecca"})
	// greetLine.Push(&GreetTask{"Miles"})
	// greetLine.Push(&GreetTask{"Emily"})
	// greetLine.Push(&GreetTask{"Steven"})
	// greetLine.Push(&GreetTask{"George"})

	// hire a worker to create greetings
	greetLine.Hire(1)

	// hire some workers to print the greetings
	printLine.Hire(8)

	// stop and wait for the lines to finish all of the tasks
	greetLine.Wait()
	printLine.Wait()
}
