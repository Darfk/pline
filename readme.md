Production Line
===

pline is a small parallel task library fit for IO bound tasks

synopsis
---

pline was created because I built a web scaper and wanted to generalise the parallel recursive nature of fetching an internet resource, parsing it and fetching more resources based off the previous. It is very similar to [conduit](https://github.com/darfk/conduit) but much more simple.

usage
---

To use pline you must define how your tasks are to be 'done' and what 'kind' they are. You must also specify the tasks generated as a result of the task being completed.

    type Task interface {
      Kind() int
      Do() []Task
    }
    
    type SimpleTask struct {...}
    func (t *SimpleTask) Kind() int { return 1 }
    func (t *SimpleTask) Do() []pline.Task {
        println(t)
        return []pline.Task{
            &SimpleTask{...},
        }
    }

Create the production line, start it, hire workers and push tasks

    line := pline.NewLine()
    line.Start()
    // hire 8 workers to work on tasks with kind 1
    line.Hire(1, 8)
    line.Push([]pline.Task{&SimpleTask{...}})
    
Wait for all tasks to be completed and for all workers to become idle

    line.Wait()

For a slightly more comprehensive example; see [greeting example](https://github.com/Darfk/pline/blob/master/examples/greeting.go)
