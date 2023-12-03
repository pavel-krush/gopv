# gopv
**gopv** is a go library that helps track the progress of long-running tasks.
It gives information such as time elapsed, percentage completed, current rps, eta and lots more.

# Installation
```bash
go get github.com/pavel-krush/gopv
```

# Usage

```go
// Initialize library and start reporting the progress
pv := gopv.New(total)
gopv.StartCtx(ctx) // or gopv.StartChan(ctx.Done())

// Execute long-running procedure
for i := 0; i < total; i++ {
    <-time.After(time.Millisecond * 100)
    pv.Add(1)
}
```

# Stopping
Progress tracker can be stopped by cancelling the context which was passed to `StartCtx()` or by closing(orwriting to) the channel passed to `StartChan`.

# Customizing
By default, gopv generates reports in the following format:
```text
[2023-12-02 08:52:24] - working (29/360) done 8%, RPS 9.65, elapsed 3s, ETA 34s
```

You can customize it by giving format to TextReporter, or on initialization by specialized constructor `NewTextWithLegend`:
```go
pv := gopv.NewTextWithLegend(498, "{progress_bar} {now} {percent_int} {rps_avg} ")
```

Which produces the following output:
```text
[##----------------------------------------------------------------------------] 2023-12-03 01:45:00 3 8.99
```

# Legend placeholders
There are many placeholders available for TextReporter:
- {now} - current time
- {started_at} - time when progress was started
- {dt} - time since last report
- {total} - total number of items
- {done} - number of items done
- {left} - number of items left
- {ratio} - ratio of done items to total
- {percent_int} - integer percent of done items to total
- {percent_float} - percent of done items to total
- {elapsed} - time elapsed since start
- {eta} - estimated time to finish
- {rps_avg} - average done items per second
- {rps_inst} - instant RPS(rps since last report)
- {rpm} - average done items per minute
- {progress_bar} - text-based progress bar
