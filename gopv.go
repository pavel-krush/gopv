package gopv

import (
	"context"
	"sync/atomic"
	"time"
)

type Progress struct {
	total            int64
	done             int64
	startedAt        time.Time
	reportTime       time.Duration
	lastReportedDone int64
	lastReportedAt   time.Time

	reporter Reporter
	doneCh   chan struct{}
}

var DefaultReportTime = time.Second

// New creates new progress tracker
func New(total int) *Progress {
	if total <= 0 {
		panic("total should be greater than 0")
	}

	return &Progress{
		total:      int64(total),
		reportTime: DefaultReportTime,
		reporter:   NewTextReporter(),
		doneCh:     make(chan struct{}),
	}
}

// NewTextWithLegend is just a shortcut for
// New(total).WithReporter(NewTextReporter().WithLegend(legend))
func NewTextWithLegend(total int, legend string) *Progress {
	return New(total).WithReporter(NewTextReporter().WithLegend(legend))
}

// WithReporter returns a new instance of progress tracker with custom reporter
func (p *Progress) WithReporter(r Reporter) *Progress {
	cp := *p
	cp.reporter = r
	return &cp
}

// StartCtx starts progress tracker using context
func StartCtx(p *Progress, ctx context.Context) {
	StartChan(p, ctx.Done())
}

// StartChan starts progress tracker using done channel
func StartChan[T any](p *Progress, done <-chan T) {
	p.startedAt = time.Now()
	p.lastReportedAt = p.startedAt
	go func() {
		defer func() {
			p.reporter.Finalize()
			defer close(p.doneCh)
		}()
		p.reporter.Report(p.Report())
		for {
			select {
			case <-done:
				return
			case <-time.After(p.reportTime):
				p.reporter.Report(p.Report())
			}
		}
	}()
}

// Add reports done items to the progress tracker
func (p *Progress) Add(done int) {
	atomic.AddInt64(&p.done, int64(done))
}

// Report returns current progress report
func (p *Progress) Report() Report {
	if p.total == 0 {
		return Report{}
	}

	now := time.Now()
	dt := now.Sub(p.lastReportedAt)
	done := atomic.LoadInt64(&p.done)
	ratio := float64(done) / float64(p.total)
	elapsed := now.Sub(p.startedAt)
	rps := float64(done) / now.Sub(p.startedAt).Seconds()
	var eta time.Duration
	if rps != 0 {
		eta = time.Duration(float64(p.total-done)/rps) * time.Second
	}

	defer func() {
		p.lastReportedDone = done
		p.lastReportedAt = now
	}()

	return Report{
		Now:          now,
		StartedAt:    p.startedAt,
		DT:           dt,
		Total:        int(p.total),
		Done:         int(done),
		Left:         int(p.total) - int(done),
		Ratio:        ratio,
		PercentInt:   int(ratio * 100),
		PercentFloat: ratio * 100,
		Elapsed:      elapsed,
		ETA:          eta,
		RPSAvg:       rps,
		RPSInst:      float64(done-p.lastReportedDone) / dt.Seconds(),
		RPMAvg:       float64(done) / now.Sub(p.startedAt).Minutes(),
	}
}

func (p *Progress) Done() chan struct{} {
	return p.doneCh
}
