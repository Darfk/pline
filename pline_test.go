package pline

import (
	"context"
	"fmt"
	"testing"
)

var (
	group1 *Group
	group2 *Group
)

type TestPlineTask struct {
	n int
	t *testing.T
}

func (t *TestPlineTask) Run(ctx context.Context) {
	t.t.Log(t.n)
	if t.n < 4 {
		group1.Push(&TestPlineTask{t: t.t, n: t.n + 1})
		group1.Push(&TestPlineTask{t: t.t, n: t.n + 1})
	}
}

func (t *TestPlineTask) String() string {
	return fmt.Sprintf("test task n:%d", t.n)
}

func TestPline(t *testing.T) {
	ctx := context.Background()
	line := NewLine()
	group1 = line.NewGroup(3)
	line.Start(ctx)
	group1.Push(&TestPlineTask{n: 1, t: t})
	line.Wait()
}

func TestWalk(t *testing.T) {
	ctx := context.Background()

	line := NewLine()

	group1 = line.NewGroup(0)

	line.Start(ctx)

	group1.Push(&TestPlineTask{n: 1, t: t})
	group1.Push(&TestPlineTask{n: 2, t: t})
	group1.Push(&TestPlineTask{n: 3, t: t})
	group1.Push(&TestPlineTask{n: 4, t: t})

	t.Log(group1.WalkTasks(func(i int, task Task) {
		t := task.(*TestPlineTask)
		t.t.Log(t.n)
	}))

	t.Log(group1.Hire(1))

	line.Finish()
}

type CancellableTask struct {
	cancel chan struct{}
	t      *testing.T
}

func (t *CancellableTask) Run(ctx context.Context) {
	t.t.Log("waiting for task cancel")
	<-t.cancel
	t.t.Log("task cancelled")
}

func TestCancel(t *testing.T) {
	ctx := context.Background()

	cancel := make(chan struct{})
	line := NewLine()
	group1 = line.NewGroup(1)
	line.Start(ctx)
	group1.Push(&CancellableTask{cancel, t})

	go func() {
		cancel <- struct{}{}
	}()

	line.Finish()
}
