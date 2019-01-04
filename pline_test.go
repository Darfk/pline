package pline

import (
	"testing"
)

type TestPlineTask struct {
	n int
	t *testing.T
}

func (t *TestPlineTask) Kind() int { return 0 }
func (t *TestPlineTask) Do() []Task {
	t.t.Log(t.n)

	if t.n < 4 {
		return []Task{
			&TestPlineTask{t: t.t, n: t.n + 1},
			&TestPlineTask{t: t.t, n: t.n + 1},
		}
	} else {
		return nil
	}
}

func TestPline(t *testing.T) {
	line := NewLine()
	line.Start()
	line.Push(&TestPlineTask{n: 1, t: t})
	line.Hire(0, 5)
	line.Hire(0, -2)
	line.Wait()
}

type CancellableTask struct {
	cancel chan struct{}
	t      *testing.T
}

func (t *CancellableTask) Kind() int { return 0 }
func (t *CancellableTask) Do() []Task {
	t.t.Log("waiting for task cancel")
	<-t.cancel
	t.t.Log("task cancelled")
	return nil
}

func TestCancel(t *testing.T) {
	cancel := make(chan struct{})
	line := NewLine()
	line.Start()
	line.Push(&CancellableTask{cancel, t})
	line.Hire(0, 1)

	go func() {
		cancel <- struct{}{}
	}()

	line.Finish()
}
