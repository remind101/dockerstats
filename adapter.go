package stats

import (
	"bytes"
	"fmt"
	"math"
	"text/template"

	"github.com/fsouza/go-dockerclient"
	"github.com/quipo/statsd"
)

var DefaultAdapter Adapter = &nullAdapter{}

// L2MetTemplate is a template for outputing metrics as l2met samples.
var L2MetTemplate = `{{.Type}}#{{.Name}}={{.Value}} source={{.Container.Name}}.{{.Hostname}}`

type nullAdapter struct{}

func (a *nullAdapter) Incr(c *docker.Container, n string, v uint64)   {}
func (a *nullAdapter) Sample(c *docker.Container, n string, v uint64) {}

// LogAdapter is a drain that drains the metrics to stdout in l2met format.
type LogAdapter struct {
	template *template.Template
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
		template: t,
	}, nil
}

func (a *LogAdapter) Sample(c *docker.Container, name string, value uint64) {
	a.write(c, "sample", name, value)
}

func (a *LogAdapter) Incr(c *docker.Container, name string, value uint64) {
	a.write(c, "count", name, value)
}

func (a *LogAdapter) write(c *docker.Container, typ, name string, value uint64) {
	data := stat{
		Container: c,
		Type:      typ,
		Name:      name,
		Value:     value,
	}
	fmt.Println(renderTemplate(a.template, data))
}

// StatsdTemplate defines the template used to render the statsd metric name.
var StatsdTemplate = `{{.Name}}.source__{{.Container.Name}}__`

type StatsdAdapter struct {
	client   *statsd.StatsdClient
	template *template.Template
}

func NewStatsdAdapter(addr string, tmpl string) (*StatsdAdapter, error) {
	c := statsd.NewStatsdClient(addr, "")

	if tmpl == "" {
		tmpl = StatsdTemplate
	}

	t, err := template.New("stat").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	return &StatsdAdapter{
		client:   c,
		template: t,
	}, nil

}

func (a *StatsdAdapter) Incr(c *docker.Container, name string, value uint64) {
	if value <= math.MaxInt64 {
		a.client.Incr(a.name(c, name), int64(value))
	}
}

func (a *StatsdAdapter) Sample(c *docker.Container, name string, value uint64) {
	if value <= math.MaxInt64 {
		a.client.Gauge(a.name(c, name), int64(value))
	}
}

func (a *StatsdAdapter) name(c *docker.Container, name string) string {
	data := stat{
		Container: c,
		Name:      name,
	}
	return renderTemplate(a.template, data)
}

func renderTemplate(t *template.Template, data stat) string {
	b := new(bytes.Buffer)
	if err := t.Execute(b, data); err != nil {
		panic(err)
	}
	return b.String()
}
