package stats

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/fsouza/go-dockerclient"
)

var DefaultAdapter Adapter = &nullAdapter{}

// L2MetTemplate is a template for outputing metrics as l2met samples.
var L2MetTemplate = `{{.Type}}#{{.Name}}={{.Value}} source={{.Container.Name}}.{{.Hostname}}`

type nullAdapter struct{}

func (a *nullAdapter) Stats(_ *Stats) error { return nil }
func (a *nullAdapter) Event(_ *Event) error { return nil }

// LogAdapter is a drain that drains the metrics to stdout in l2met format.
type LogAdapter struct {
	Template *template.Template
}

// NewLogAdapter parses the template string as a text/template and returns a new
// LogAdapter instance.
func NewLogAdapter(tmpl string) (*LogAdapter, error) {
	if tmpl == "" {
		tmpl = L2MetTemplate
	}

	t, err := template.New("stat").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	return &LogAdapter{
		Template: t,
	}, nil
}

func (a *LogAdapter) Event(event *Event) error {
	w := &l2metWriter{template: a.Template, container: event.Container}
	w.incr(fmt.Sprintf("Container.%s", strings.Title(event.Status)))
	return nil
}

func (a *LogAdapter) Stats(stats *Stats) error {
	w := &l2metWriter{template: a.Template, container: stats.Container}

	// Network
	w.sample("Network.RxDropped", stats.Network.RxDropped)
	w.sample("Network.RxBytes", stats.Network.RxBytes)
	w.sample("Network.RxErrors", stats.Network.RxErrors)
	w.sample("Network.TxPackets", stats.Network.TxPackets)
	w.sample("Network.RxPackets", stats.Network.RxPackets)
	w.sample("Network.TxErrors", stats.Network.TxErrors)
	w.sample("Network.TxBytes", stats.Network.TxBytes)

	// MemoryStats
	w.sample("MemoryStats.Stats.TotalPgmafault", stats.MemoryStats.Stats.TotalPgmafault)
	w.sample("MemoryStats.Stats.Cache", stats.MemoryStats.Stats.Cache)
	w.sample("MemoryStats.Stats.MappedFile", stats.MemoryStats.Stats.MappedFile)
	w.sample("MemoryStats.Stats.TotalInactiveFile", stats.MemoryStats.Stats.TotalInactiveFile)
	w.sample("MemoryStats.Stats.Pgpgout", stats.MemoryStats.Stats.Pgpgout)
	w.sample("MemoryStats.Stats.Rss", stats.MemoryStats.Stats.Rss)
	w.sample("MemoryStats.Stats.TotalMappedFile", stats.MemoryStats.Stats.TotalMappedFile)
	w.sample("MemoryStats.Stats.Writeback", stats.MemoryStats.Stats.Writeback)
	w.sample("MemoryStats.Stats.Unevictable", stats.MemoryStats.Stats.Unevictable)
	w.sample("MemoryStats.Stats.Pgpgin", stats.MemoryStats.Stats.Pgpgin)
	w.sample("MemoryStats.Stats.TotalUnevictable", stats.MemoryStats.Stats.TotalUnevictable)
	w.sample("MemoryStats.Stats.Pgmajfault", stats.MemoryStats.Stats.Pgmajfault)
	w.sample("MemoryStats.Stats.TotalRss", stats.MemoryStats.Stats.TotalRss)
	w.sample("MemoryStats.Stats.TotalRssHuge", stats.MemoryStats.Stats.TotalRssHuge)
	w.sample("MemoryStats.Stats.TotalWriteback", stats.MemoryStats.Stats.TotalWriteback)
	w.sample("MemoryStats.Stats.TotalInactiveAnon", stats.MemoryStats.Stats.TotalInactiveAnon)
	w.sample("MemoryStats.Stats.RssHuge", stats.MemoryStats.Stats.RssHuge)
	w.sample("MemoryStats.Stats.HierarchicalMemoryLimit", stats.MemoryStats.Stats.HierarchicalMemoryLimit)
	w.sample("MemoryStats.Stats.TotalPgfault", stats.MemoryStats.Stats.TotalPgfault)
	w.sample("MemoryStats.Stats.TotalActiveFile", stats.MemoryStats.Stats.TotalActiveFile)
	w.sample("MemoryStats.Stats.ActiveAnon", stats.MemoryStats.Stats.ActiveAnon)
	w.sample("MemoryStats.Stats.TotalActiveAnon", stats.MemoryStats.Stats.TotalActiveAnon)
	w.sample("MemoryStats.Stats.TotalPgpgout", stats.MemoryStats.Stats.TotalPgpgout)
	w.sample("MemoryStats.Stats.TotalCache", stats.MemoryStats.Stats.TotalCache)
	w.sample("MemoryStats.Stats.InactiveAnon", stats.MemoryStats.Stats.InactiveAnon)
	w.sample("MemoryStats.Stats.ActiveFile", stats.MemoryStats.Stats.ActiveFile)
	w.sample("MemoryStats.Stats.Pgfault", stats.MemoryStats.Stats.Pgfault)
	w.sample("MemoryStats.Stats.InactiveFile", stats.MemoryStats.Stats.InactiveFile)
	w.sample("MemoryStats.Stats.TotalPgpgin", stats.MemoryStats.Stats.TotalPgpgin)
	w.sample("MemoryStats.MaxUsage", stats.MemoryStats.MaxUsage)
	w.sample("MemoryStats.Usage", stats.MemoryStats.Usage)
	w.sample("MemoryStats.Failcnt", stats.MemoryStats.Failcnt)
	w.sample("MemoryStats.Limit", stats.MemoryStats.Limit)

	// CPUStats
	for i, v := range stats.CPUStats.CPUUsage.PercpuUsage {
		w.sample(fmt.Sprintf("CPUStats.CPUUsage.PercpuUsage.%d", i), v)
	}
	w.sample("CPUStats.CPUUsage.UsageInUsermode", stats.CPUStats.CPUUsage.UsageInUsermode)
	w.sample("CPUStats.CPUUsage.TotalUsage", stats.CPUStats.CPUUsage.TotalUsage)
	w.sample("CPUStats.CPUUsage.UsageInKernelmode", stats.CPUStats.CPUUsage.UsageInKernelmode)
	w.sample("CPUStats.SystemCPUUsage", stats.CPUStats.SystemCPUUsage)
	w.sample("CPUStats.ThrottlingData.Periods", stats.CPUStats.ThrottlingData.Periods)
	w.sample("CPUStats.ThrottlingData.ThrottledPeriods", stats.CPUStats.ThrottlingData.ThrottledPeriods)
	w.sample("CPUStats.ThrottlingData.ThrottledTime", stats.CPUStats.ThrottlingData.ThrottledTime)

	return nil
}

type l2metWriter struct {
	template  *template.Template
	container *docker.Container
}

func (w *l2metWriter) sample(name string, value interface{}) {
	w.write("sample", name, value)
}

func (w *l2metWriter) incr(name string) {
	w.write("count", name, 1)
}

func (w *l2metWriter) write(typ, name string, value interface{}) {
	b := new(bytes.Buffer)
	data := stat{
		Container: w.container,
		Type:      typ,
		Name:      name,
		Value:     value,
	}

	if err := w.template.Execute(b, data); err != nil {
		panic(err)
	}

	fmt.Println(b.String())
}
