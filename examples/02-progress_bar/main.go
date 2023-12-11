package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pavel-krush/gopv"
)

func main() {
	fmt.Println("executing long task1...")
	task()

	fmt.Println("executing long task2...")
	task()
}

func task() {
	const total = 50

	ctx, cancel := context.WithCancel(context.Background())
	pv := gopv.NewTextWithLegend(total, gopv.TextReporterLegendProgressBar)
	gopv.StartChan(pv, ctx.Done())

	for i := 0; i < total; i++ {
		<-time.After(time.Millisecond * 100)
		pv.Add(1)
	}
	cancel()
	<-pv.Done()
}
