package stats

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/fsouza/go-dockerclient"
)

// L2MetTemplate is a template for outputing metrics as l2met samples.
var L2MetTemplate = "sample#{{.Name}}={{.Value}} source={{.Container.Name}}"

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

func (a *LogAdapter) Drain(stats *Stats) error {
	w := &l2metWriter{template: a.Template, container: stats.Container}

	// Network
	w.write("Network.RxDropped", stats.Network.RxDropped)
	w.write("Network.RxBytes", stats.Network.RxBytes)
	w.write("Network.RxErrors", stats.Network.RxErrors)
	w.write("Network.TxPackets", stats.Network.TxPackets)
	w.write("Network.RxPackets", stats.Network.RxPackets)
	w.write("Network.TxErrors", stats.Network.TxErrors)
	w.write("Network.TxBytes", stats.Network.TxBytes)

	// MemoryStats
	w.write("MemoryStats.Stats.TotalPgmafault", stats.MemoryStats.Stats.TotalPgmafault)
	w.write("MemoryStats.Stats.Cache", stats.MemoryStats.Stats.Cache)
	w.write("MemoryStats.Stats.MappedFile", stats.MemoryStats.Stats.MappedFile)
	w.write("MemoryStats.Stats.TotalInactiveFile", stats.MemoryStats.Stats.TotalInactiveFile)
	w.write("MemoryStats.Stats.Pgpgout", stats.MemoryStats.Stats.Pgpgout)
	w.write("MemoryStats.Stats.Rss", stats.MemoryStats.Stats.Rss)
	w.write("MemoryStats.Stats.TotalMappedFile", stats.MemoryStats.Stats.TotalMappedFile)
	w.write("MemoryStats.Stats.Writeback", stats.MemoryStats.Stats.Writeback)
	w.write("MemoryStats.Stats.Unevictable", stats.MemoryStats.Stats.Unevictable)
	w.write("MemoryStats.Stats.Pgpgin", stats.MemoryStats.Stats.Pgpgin)
	w.write("MemoryStats.Stats.TotalUnevictable", stats.MemoryStats.Stats.TotalUnevictable)
	w.write("MemoryStats.Stats.Pgmajfault", stats.MemoryStats.Stats.Pgmajfault)
	w.write("MemoryStats.Stats.TotalRss", stats.MemoryStats.Stats.TotalRss)
	w.write("MemoryStats.Stats.TotalRssHuge", stats.MemoryStats.Stats.TotalRssHuge)
	w.write("MemoryStats.Stats.TotalWriteback", stats.MemoryStats.Stats.TotalWriteback)
	w.write("MemoryStats.Stats.TotalInactiveAnon", stats.MemoryStats.Stats.TotalInactiveAnon)
	w.write("MemoryStats.Stats.RssHuge", stats.MemoryStats.Stats.RssHuge)
	w.write("MemoryStats.Stats.HierarchicalMemoryLimit", stats.MemoryStats.Stats.HierarchicalMemoryLimit)
	w.write("MemoryStats.Stats.TotalPgfault", stats.MemoryStats.Stats.TotalPgfault)
	w.write("MemoryStats.Stats.TotalActiveFile", stats.MemoryStats.Stats.TotalActiveFile)
	w.write("MemoryStats.Stats.ActiveAnon", stats.MemoryStats.Stats.ActiveAnon)
	w.write("MemoryStats.Stats.TotalActiveAnon", stats.MemoryStats.Stats.TotalActiveAnon)
	w.write("MemoryStats.Stats.TotalPgpgout", stats.MemoryStats.Stats.TotalPgpgout)
	w.write("MemoryStats.Stats.TotalCache", stats.MemoryStats.Stats.TotalCache)
	w.write("MemoryStats.Stats.InactiveAnon", stats.MemoryStats.Stats.InactiveAnon)
	w.write("MemoryStats.Stats.ActiveFile", stats.MemoryStats.Stats.ActiveFile)
	w.write("MemoryStats.Stats.Pgfault", stats.MemoryStats.Stats.Pgfault)
	w.write("MemoryStats.Stats.InactiveFile", stats.MemoryStats.Stats.InactiveFile)
	w.write("MemoryStats.Stats.TotalPgpgin", stats.MemoryStats.Stats.TotalPgpgin)
	w.write("MemoryStats.MaxUsage", stats.MemoryStats.MaxUsage)
	w.write("MemoryStats.Usage", stats.MemoryStats.Usage)
	w.write("MemoryStats.Failcnt", stats.MemoryStats.Failcnt)
	w.write("MemoryStats.Limit", stats.MemoryStats.Limit)

	// CPUStats
	for i, v := range stats.CPUStats.CPUUsage.PercpuUsage {
		w.write(fmt.Sprintf("CPUStats.CPUUsage.PercpuUsage.%d", i), v)
	}
	w.write("CPUStats.CPUUsage.UsageInUsermode", stats.CPUStats.CPUUsage.UsageInUsermode)
	w.write("CPUStats.CPUUsage.TotalUsage", stats.CPUStats.CPUUsage.TotalUsage)
	w.write("CPUStats.CPUUsage.UsageInKernelmode", stats.CPUStats.CPUUsage.UsageInKernelmode)
	w.write("CPUStats.SystemCPUUsage", stats.CPUStats.SystemCPUUsage)
	w.write("CPUStats.ThrottlingData.Periods", stats.CPUStats.ThrottlingData.Periods)
	w.write("CPUStats.ThrottlingData.ThrottledPeriods", stats.CPUStats.ThrottlingData.ThrottledPeriods)
	w.write("CPUStats.ThrottlingData.ThrottledTime", stats.CPUStats.ThrottlingData.ThrottledTime)

	return nil
}

// stat represents a single stat and is provided as the context to a
// template.
type stat struct {
	Container *docker.Container
	Name      string
	Value     interface{}
}

type l2metWriter struct {
	template  *template.Template
	container *docker.Container
}

func (w *l2metWriter) write(name string, value interface{}) {
	b := new(bytes.Buffer)
	data := stat{
		Container: w.container,
		Name:      name,
		Value:     value,
	}

	if err := w.template.Execute(b, data); err != nil {
		panic(err)
	}

	fmt.Println(b.String())
}
