package main

import (
	"context"
	"time"

	"github.com/pavel-krush/gopv"
)

func main() {
	const total = 360

	ctx, cancel := context.WithCancel(context.Background())
	pg := gopv.NewTextWithLegend(total, gopv.TextReporterLegendProgressBar)
	pg.Start(ctx)

	for i := 0; i < total; i++ {
		<-time.After(time.Millisecond * 100)
		pg.Add(1)
	}
	cancel()
}
