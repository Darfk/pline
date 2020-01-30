package main

import (
	pline ".."
	"log"
	"time"
	"context"
	"fmt"
)

var (
	line *pline.Line
	printGroup *pline.Group
	greetGroup *pline.Group
)

type GreetTask struct {
	name string
}

func (t *GreetTask) String() string {
	return fmt.Sprintf("greet task for: %s", t.name)
}

func (t *GreetTask) Run(ctx context.Context) {
	log.Println("greet task:", t.name)

	time.Sleep(time.Millisecond * 250)

	if t.name == "Andy" || t.name == "Nathan" {
		// say hi to our friends
		printGroup.Push(&PrintTask{"Hey " + t.name + "!"})
		return
	} else if t.name == "Kirk" {
		// give Kirk the cold shoulder
		return
	} else if t.name == "Tim" {
		// Tim ruins it for everyone
		return
	}
	// we haven't met them before

	printGroup.Push(&PrintTask{"Hi " + t.name})
	printGroup.Push(&PrintTask{"Welcome to pline " + t.name + "."})
}

type PrintTask struct {
	text string
}

func (t *PrintTask) String() string {
	return fmt.Sprintf("print task for: %s", t.text)
}

func (t *PrintTask) Run(ctx context.Context) {
	// printing takes a while, we will use a worker pool
	time.Sleep(time.Millisecond * 250)
	log.Println("print task:", t.text)
	return
}

func main() {
	// create the production lines
	line = pline.NewLine()

	greetGroup = line.NewGroup(1)
	printGroup = line.NewGroup(8)

	ctx := context.Background()

	// start the line so that it can accept tasks
	line.Start(ctx)

	// push all the print tasks into the production line
	printGroup.Push(&PrintTask{"This is pline"})

	// push all the greet tasks into the production line
	greetGroup.Push(&GreetTask{"Andy"})
	greetGroup.Push(&GreetTask{"Nathan"})
	greetGroup.Push(&GreetTask{"Kirk"})
	greetGroup.Push(&GreetTask{"Rebecca"})
	greetGroup.Push(&GreetTask{"Miles"})
	greetGroup.Push(&GreetTask{"Tim"})
	greetGroup.Push(&GreetTask{"Emily"})
	greetGroup.Push(&GreetTask{"Steven"})
	greetGroup.Push(&GreetTask{"George"})

	// stop and wait for the line to finish all of the tasks
	line.Wait()
}
