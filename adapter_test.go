package stats_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/remind101/dockerstats"
)

func TestLogAdapter(t *testing.T) {
	b := new(bytes.Buffer)
	a, err := stats.NewLogAdapter(`{{.Type}}#{{.Name}}={{.Value}} source={{.Container.Name}}`, b)
	if err != nil {
		t.Fatal(err)
	}

	c := &docker.Container{
		Name: "dummy",
	}

	a.Sample(c, "foo.bar", 1)
	if got, want := b.String(), "sample#foo.bar=1 source=dummy\n"; got != want {
		t.Errorf("Sample() => %q; want %q", got, want)
	}
}

type fakeStatsdClient struct {
	Stats []string
}

func (c *fakeStatsdClient) Incr(name string, value int64) error {
	return c.write(name, value, "c")
}

func (c *fakeStatsdClient) Gauge(name string, value int64) error {
	return c.write(name, value, "g")
}

func (c *fakeStatsdClient) write(name string, value int64, typ string) error {
	c.Stats = append(c.Stats, fmt.Sprintf("%s:%d|%s", name, value, typ))
	return nil
}

func TestStatsdAdapter(t *testing.T) {
	client := &fakeStatsdClient{Stats: []string{}}
	a, err := stats.NewStatsdAdapter(client, `tests.source__{{.Env "SOURCE"}}__`)
	if err != nil {
		t.Fatal(err)
	}

	c := &docker.Container{
		Name: "dummy",
		Config: &docker.Config{
			Env: []string{"SOURCE=dockerstats.tests.statsd"},
		},
	}

	a.Incr(c, "foo.bar", 1)
	if got, want := client.Stats[0], "tests.source__dockerstats.tests.statsd__:1|c"; got != want {
		t.Errorf("Incr() => %q; want %q", got, want)
	}
}
