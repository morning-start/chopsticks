package output

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type ProgressBar struct {
	mu        sync.Mutex
	prefix    string
	total     int64
	current   int64
	startTime time.Time
	writer    io.Writer
	width     int
	spinnerPos int
	showSpeed bool
	done      bool
	spinner   bool
}

var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func NewProgressBar(total int64) *ProgressBar {
	return &ProgressBar{
		total:     total,
		width:     50,
		showSpeed: true,
		startTime: time.Now(),
		writer:    stdout,
	}
}

func (p *ProgressBar) SetPrefix(prefix string) *ProgressBar {
	p.prefix = prefix
	return p
}

func (p *ProgressBar) SetWidth(width int) *ProgressBar {
	p.width = width
	return p
}

func (p *ProgressBar) SetWriter(w io.Writer) *ProgressBar {
	p.writer = w
	return p
}

func (p *ProgressBar) ShowSpeed(show bool) *ProgressBar {
	p.showSpeed = show
	return p
}

func (p *ProgressBar) UseSpinner(use bool) *ProgressBar {
	p.spinner = use
	return p
}

func (p *ProgressBar) Add(n int64) *ProgressBar {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current += n
	return p
}

func (p *ProgressBar) Set(n int64) *ProgressBar {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = n
	return p
}

func (p *ProgressBar) Current() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

func (p *ProgressBar) Total() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.total
}

func (p *ProgressBar) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done = true
	p.current = p.total
	p.render()
}

func (p *ProgressBar) render() {
	if p.writer == nil {
		return
	}

	if p.spinner {
		p.renderSpinner()
		return
	}
	p.renderBar()
}

func (p *ProgressBar) renderSpinner() {
	elapsed := time.Since(p.startTime).Seconds()
	rate := float64(p.current) / elapsed

	var bar strings.Builder
	if p.prefix != "" {
		bar.WriteString(p.prefix)
		bar.WriteString(" ")
	}

	bar.WriteString(spinnerChars[p.spinnerPos])
	p.spinnerPos = (p.spinnerPos + 1) % len(spinnerChars)

	bar.WriteString(" ")
	bar.WriteString(fmt.Sprintf("%d", p.current))

	if p.total > 0 {
		bar.WriteString("/")
		bar.WriteString(fmt.Sprintf("%d", p.total))
	}

	if p.showSpeed && rate > 0 {
		bar.WriteString(" (")
		bar.WriteString(formatBytes(int64(rate)))
		bar.WriteString("/s)")
	}

	bar.WriteString("\r")
	fmt.Fprint(p.writer, bar.String())
}

func (p *ProgressBar) renderBar() {
	elapsed := time.Since(p.startTime).Seconds()
	rate := float64(p.current) / elapsed

	var bar strings.Builder

	if p.prefix != "" {
		bar.WriteString(p.prefix)
		bar.WriteString(" ")
	}

	filled := float64(p.current) / float64(p.total) * float64(p.width)
	filledInt := int(filled)

	bar.WriteString("[")
	for i := 0; i < p.width; i++ {
		if i < filledInt {
			bar.WriteString("=")
		} else if i == filledInt {
			bar.WriteString(">")
		} else {
			bar.WriteString(" ")
		}
	}
	bar.WriteString("] ")

	percent := float64(p.current) / float64(p.total) * 100
	bar.WriteString(fmt.Sprintf("%.1f%%", percent))

	bar.WriteString(" ")
	bar.WriteString(fmt.Sprintf("%d/%d", p.current, p.total))

	if p.showSpeed && rate > 0 {
		bar.WriteString(" (")
		bar.WriteString(formatBytes(int64(rate)))
		bar.WriteString("/s)")
	}

	if p.done {
		elapsedStr := formatDuration(elapsed)
		bar.WriteString(" ")
		bar.WriteString(elapsedStr)
	}

	bar.WriteString("\r")
	fmt.Fprint(p.writer, bar.String())
}

func (p *ProgressBar) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.total == 0 {
		return ""
	}

	elapsed := time.Since(p.startTime).Seconds()
	rate := float64(p.current) / elapsed

	var sb strings.Builder
	if p.prefix != "" {
		sb.WriteString(p.prefix)
		sb.WriteString(" ")
	}

	filled := float64(p.current) / float64(p.total) * float64(p.width)
	filledInt := int(filled)

	sb.WriteString("[")
	for i := 0; i < p.width; i++ {
		if i < filledInt {
			sb.WriteString("=")
		} else if i == filledInt {
			sb.WriteString(">")
		} else {
			sb.WriteString(" ")
		}
	}
	sb.WriteString("] ")

	percent := float64(p.current) / float64(p.total) * 100
	sb.WriteString(fmt.Sprintf("%.1f%% ", percent))

	sb.WriteString(fmt.Sprintf("%d/%d", p.current, p.total))

	if p.showSpeed && rate > 0 {
		sb.WriteString(fmt.Sprintf(" (%.2f/s)", rate))
	}

	return sb.String()
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n >= unit*10 {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func formatDuration(seconds float64) string {
	d := time.Duration(seconds * float64(time.Second))
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", h, m)
}

type DownloadProgress struct {
	*ProgressBar
	filename string
	url      string
}

func NewDownloadProgress(filename string, total int64) *DownloadProgress {
	return &DownloadProgress{
		ProgressBar: NewProgressBar(total).SetPrefix("下载"),
		filename:    filename,
	}
}

func (p *DownloadProgress) SetURL(url string) *DownloadProgress {
	p.url = url
	return p
}

func (p *DownloadProgress) Filename() string {
	return p.filename
}

func (p *DownloadProgress) String() string {
	var sb strings.Builder
	sb.WriteString("下载: ")
	sb.WriteString(p.filename)
	sb.WriteString(" ")
	sb.WriteString(p.ProgressBar.String())
	return sb.String()
}
