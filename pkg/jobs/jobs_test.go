package jobs_test

import (
	"sync"
	"testing"

	"github.com/joincivil/civil-api-server/pkg/jobs"
)

type Spy struct {
	RunCount int
}

func NewSpy() *Spy {
	return &Spy{
		RunCount: 0,
	}
}

func (s *Spy) Run() {
	s.RunCount = s.RunCount + 1
}

func buildJob(spy *Spy) *jobs.Job {
	jobID := "test"
	work := func(updates chan<- string) {
		spy.Run()
		updates <- "foo"
		updates <- "bar"
	}
	return jobs.NewJob(jobID, work)
}

func TestJob(t *testing.T) {

	t.Run("run work", func(t *testing.T) {
		spy := NewSpy()
		job := buildJob(spy)
		if job.GetStatus() != "initialized" {
			t.Fatalf("job status should be `initialized`")
		}
		if spy.RunCount > 0 {
			t.Fatalf("work function should not have run")
		}

		job.Start()
		if job.GetStatus() != "initialized" {
			t.Fatalf("job status should be `initialized`")
		}
		job.WaitForFinish()
		if spy.RunCount != 1 {
			t.Fatalf("work function should be 1")
		}
	})

	t.Run("subscriptions", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(2)
		spy := NewSpy()
		spySub1 := NewSpy()
		spySub2 := NewSpy()
		job := buildJob(spy)
		sub1 := job.Subscribe()
		sub2 := job.Subscribe()
		job.Start()

		go func() {
			for range sub1.Updates {
				spySub1.Run()
			}
			wg.Done()
		}()

		go func() {
			for range sub2.Updates {
				spySub2.Run()
			}
			wg.Done()
		}()

		wg.Wait()

		if job.GetStatus() != "complete" {
			t.Fatalf("job status should be `complete`")
		}
		if spy.RunCount != 1 {
			t.Fatalf("work RunCount should be 1 but is %v", spy.RunCount)
		}
		if spySub1.RunCount != 2 {
			t.Fatalf("work RunCount should be 2 but is %v", spySub1.RunCount)
		}
		if spySub2.RunCount != 2 {
			t.Fatalf("work RunCount should be 2 but is %v", spySub2.RunCount)
		}
	})

	t.Run("unsubscribe", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(2)
		spy := NewSpy()
		jobID := "test"
		var job *jobs.Job

		signals := make(chan string)
		work := func(updates chan<- string) {
			updates <- "foo"
			// wait for a signal to resume, so sub2 can unsubscribe
			<-signals
			//job.Unsubscribe(sub2)
			updates <- "bar"
			spy.Run()
		}
		job = jobs.NewJob(jobID, work)
		sub1 := job.Subscribe()
		sub2 := job.Subscribe()
		job.Start()

		sub1Event1 := <-sub1.Updates
		sub2Event1 := <-sub2.Updates
		if sub1Event1 != "foo" && sub2Event1 != "foo" {
			t.Fatalf("unexpected event values")
		}

		job.Unsubscribe(sub2)
		signals <- "continue"
		sub3 := job.Subscribe()

		sub1Event2 := <-sub1.Updates
		sub3Event2 := <-sub3.Updates
		if sub1Event2 != "bar" && sub3Event2 != "bar" {
			t.Fatalf("unexpected event values")
		}

		if job.GetStatus() != "complete" {
			t.Fatalf("job status should be `complete`")
		}
		if spy.RunCount != 1 {
			t.Fatalf("work RunCount should be 1 but is %v", spy.RunCount)
		}
	})
}
