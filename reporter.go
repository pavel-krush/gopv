package gopv

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type Reporter interface {
	Report(report Report)
	Finalize()
}

type Report struct {
	// Current time
	Now time.Time

	// Time when progress was started
	StartedAt time.Time

	// Time since last report
	DT time.Duration

	// Total number of items
	Total int

	// Number of items done
	Done int

	// Number of items left
	Left int

	// Ratio of done items to total
	Ratio float64

	// Percent of done items to total
	PercentInt int

	// Percent of done items to total
	PercentFloat float64

	// Time elapsed since start
	Elapsed time.Duration

	// Estimated time to finish
	ETA time.Duration

	// Average done items per second
	RPSAvg float64

	// Instant RPS(rps since last report)
	RPSInst float64

	// Average done items per minute
	RPMAvg float64
}

// TextReporter is a simple reporter that writes reports to given output.
//
// Default Legend:
//
//	"[{now}] - working ({done}/{total}) done {percent_int}%%, RPS {rps_avg}, elapsed {elapsed}, ETA {eta}  "
//
// Which produces messages like following:
//
//	[2023-12-02 01:01:21] - working (39/360) done 10%, RPS 9.74, elapsed 4s, ETA 32s
//
// To customize legend see WithLegend()
type TextReporter struct {
	// config - should be copied in clone()
	legend         string
	floatPrecision int
	output         io.Writer
	pbWidth        int

	// runtime vars. should not be copied in clone()
	legendCompiled   string
	writer           *bufio.Writer
	lastLegendLength int
}

const (
	// TextReporterLegendDefault is the default legend for TextReporter
	TextReporterLegendDefault = "[{now}] - working ({done}/{total}) done {percent_int}%%, RPS {rps_avg}, elapsed {elapsed}, ETA {eta}\r"
	// TextReporterLegendProgressBar TextReporter legend with progress bar
	TextReporterLegendProgressBar = "{progress_bar} {percent_int}%%, {rps_avg} RPS, {eta} ETA\r"
	// TextReporterDefaultFloatPrecision is the default float precision for ann floats in TextReporter
	TextReporterDefaultFloatPrecision = 2
	// TextReporterDefaultProgressBarWidth is the default progress bar with for TextReporter
	TextReporterDefaultProgressBarWidth = 80
)

// NewTextReporter returns a new instance of reporter
func NewTextReporter() *TextReporter {
	return &TextReporter{
		legend:         TextReporterLegendDefault,
		floatPrecision: TextReporterDefaultFloatPrecision,
		output:         os.Stderr,
		pbWidth:        TextReporterDefaultProgressBarWidth,
	}
}

// WithLegend returns a new instance of TextReporter with custom legend.
func (r *TextReporter) WithLegend(legend string) *TextReporter {
	ret := r.clone()
	ret.legend = legend
	return ret
}

// WithFloatPrecision returns a new instance of TextReporter with custom float precision
func (r *TextReporter) WithFloatPrecision(floatPrecision int) *TextReporter {
	ret := r.clone()
	ret.floatPrecision = floatPrecision
	return ret
}

// WithOutput return a new instance of TextReporter with custom output
func (r *TextReporter) WithOutput(output io.Writer) *TextReporter {
	ret := r.clone()
	ret.output = output
	return ret
}

// WithProgressBarWidth returns a new instance of TextReporter with given progress bar width
func (r *TextReporter) WithProgressBarWidth(width int) *TextReporter {
	ret := r.clone()
	ret.pbWidth = width
	return ret
}

// Report renders report
func (r *TextReporter) Report(report Report) {
	if r.legendCompiled == "" {
		r.legendCompiled = r.compileLegend(r.legend, r.floatPrecision)
		r.writer = bufio.NewWriter(r.output)
	}

	eta := report.ETA.Round(time.Second)
	if eta <= 0 {
		eta = 0
	}

	progressBar := r.renderProgressBar(report)

	legend := fmt.Sprintf(r.legendCompiled,
		report.Now.Format("2006-01-02 03:04:05"),
		report.StartedAt.Format("2006-01-02 03:04:05"),
		report.DT.Round(time.Millisecond),
		report.Total,
		report.Done,
		report.Left,
		report.Ratio,
		report.PercentInt,
		report.PercentFloat,
		report.Elapsed.Round(time.Second),
		eta,
		report.RPSAvg,
		report.RPSInst,
		report.RPMAvg,
		progressBar,
	)
	lineLength := len(legend)

	r.writeString(legend)

	if r.lastLegendLength > lineLength {
		spaces := strings.Repeat(" ", r.lastLegendLength-lineLength)
		r.writeString(spaces)
	}

	r.lastLegendLength = lineLength
	r.flush()
}

func (r *TextReporter) Finalize() {
	r.writeString("\n")
	r.flush()
}

// compileLegend replaces placeholders with corresponding format specifiers
func (r *TextReporter) compileLegend(format string, floatPrecision int) string {
	format = strings.ReplaceAll(format, "{now}", "%[1]s")
	format = strings.ReplaceAll(format, "{started_at}", "%[2]s")
	format = strings.ReplaceAll(format, "{dt}", "%[3]s")
	format = strings.ReplaceAll(format, "{total}", "%[4]d")
	format = strings.ReplaceAll(format, "{done}", "%[5]d")
	format = strings.ReplaceAll(format, "{left}", "%[6]d")
	format = strings.ReplaceAll(format, "{ratio}", "%.{float_precision}[7]f")
	format = strings.ReplaceAll(format, "{percent_int}", "%[8]d")
	format = strings.ReplaceAll(format, "{percent_float}", "%.{float_precision}[9]f")
	format = strings.ReplaceAll(format, "{elapsed}", "%[10]s")
	format = strings.ReplaceAll(format, "{eta}", "%[11]s")
	format = strings.ReplaceAll(format, "{rps_avg}", "%.{float_precision}[12]f")
	format = strings.ReplaceAll(format, "{rps_inst}", "%.{float_precision}[13]f")
	format = strings.ReplaceAll(format, "{rpm}", "%.{float_precision}[14]f")

	format = strings.ReplaceAll(format, "{progress_bar}", "%[15]s")

	format = strings.ReplaceAll(format, "{float_precision}", strconv.Itoa(floatPrecision))
	return format
}

// renderProgressBar builds and returns string containing progress bar
func (r *TextReporter) renderProgressBar(report Report) string {
	ratio := report.Ratio
	if ratio < 0 {
		ratio = 0
	}
	progressBarWidth := r.pbWidth - 2 // [ and ]
	if progressBarWidth <= 0 {
		return ""
	}

	fillChars := int(ratio * float64(progressBarWidth))
	if fillChars > progressBarWidth {
		fillChars = progressBarWidth
	}

	fillSpaces := progressBarWidth - fillChars
	if fillSpaces < 0 {
		fillSpaces = 0
	}

	progressBar := "["
	progressBar += strings.Repeat("#", fillChars)
	progressBar += strings.Repeat("-", fillSpaces)
	progressBar += "]"

	return progressBar
}

// writeString writes given string to the output. it just proxies WriteString
// call to the output and discards errors
func (r *TextReporter) writeString(str string) {
	_, _ = r.writer.WriteString(str)
}

// fLush flushes buffered output to the underlying io stream. same as writeString
// just pass Flush call to the writer and discard error
func (r *TextReporter) flush() {
	_ = r.writer.Flush()
}

func (r *TextReporter) clone() *TextReporter {
	cp := *r
	return &cp
}
