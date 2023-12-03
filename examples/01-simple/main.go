package main

import (
	"context"
	"time"

	"github.com/pavel-krush/gopv"
)

func main() {
	const total = 360

	ctx, cancel := context.WithCancel(context.Background())
	pv := gopv.New(total)
	gopv.StartCtx(pv, ctx)

	for i := 0; i < total; i++ {
		<-time.After(time.Millisecond * 100)
		pv.Add(1)
	}
	cancel()
}
