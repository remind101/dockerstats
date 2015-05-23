// Package stat is a Go program for watching the container stats endpoint in the
// docker api and shuttling the metrics somwhere else where they belong.
// Basically logspout for container metrics.
package stats // import "github.com/remind101/dockerstats"

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
)

// DefaultResolution defines the default resolution for draining stats. We
// default to 10 seconds.
var DefaultResolution = 10

// Stats represents a set of stats from a container at a given point in time.
type Stats struct {
	*docker.Stats
	Container *docker.Container
}

// Adapter is an interface for draining metrics somewhere.
type Adapter interface {
	Drain(*Stats) error
}

// Stat is a context struct that manages the lifecycle of container metrics. It
// watches for container start, restart and stop events and streams metrics to a
// Drain.
type Stat struct {
	// Adapter is the Adapter that will be used to drain Stats.
	Adapter

	// Resolution defines how often stats will be sent to the adapter to be
	// drained. Any stats received from the docker daemon before the next
	// tick will be dropped. Throttling is on a per container basis. The
	// zero value is DefaultResolution.
	Resolution int

	mu         sync.Mutex
	containers map[string]*docker.Container
	client     *docker.Client
}

// New returns a new Stat instance with a configured docker client.
func New(host string) (*Stat, error) {
	c, err := docker.NewClient(host)
	if err != nil {
		return nil, err
	}

	return &Stat{
		client:     c,
		containers: make(map[string]*docker.Container),
	}, nil
}

// Run begins starts draining metrics from all of the currently running
// containers and starts watching for new containers to drain metrics from. This
// call is blocking.
func (s *Stat) Run() error {
	containers, err := s.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}

	for _, c := range containers {
		go s.start(c.ID)
	}

	events := make(chan *docker.APIEvents)
	if err := s.client.AddEventListener(events); err != nil {
		return err
	}

	for event := range events {
		switch event.Status {
		case "start", "restart":
			go s.start(event.ID)
		case "die":
			go s.stop(event.ID)
		}
	}

	return errors.New("unexpected stop")
}

func (s *Stat) start(containerID string) {
	s.mu.Lock()

	if _, ok := s.containers[containerID]; ok {
		// We're already watching metrics from this container. Nothing
		// to do.
		s.mu.Unlock()
		return
	}

	container, err := s.client.InspectContainer(containerID)
	if err != nil {
		debug("inspect: err: %s", err)
		s.mu.Unlock()
		return
	}
	container.Name = strings.Replace(container.Name, "/", "", 1)

	s.containers[containerID] = container
	s.mu.Unlock()

	debug("draining: %s", container.Name)

	stats := make(chan *docker.Stats)
	go func() {
		if err := s.client.Stats(docker.StatsOptions{
			ID:    containerID,
			Stats: stats,
		}); err != nil {
			debug("stats: err: %s", err)
		}
	}()

	ticker := newTicker(s.Resolution)

	for stat := range stats {
		// We select on the ticker channel. If a tick event isn't ready, we'll
		// return which will drop this stats message.
		select {
		case <-ticker.C:
			if err := s.drain(container, stat); err != nil {
				debug("drain: err: %s", err)
			}
		default:
			// Drop the stat.
		}
	}
}

func (s *Stat) stop(containerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if container, ok := s.containers[containerID]; ok {
		debug("stopping: %s", container.Name)
	}
}

func (s *Stat) drain(container *docker.Container, stats *docker.Stats) error {
	if s.Adapter == nil {
		return nil
	}

	return s.Adapter.Drain(&Stats{
		Stats:     stats,
		Container: container,
	})
}

func newTicker(resolution int) *time.Ticker {
	if resolution == 0 {
		resolution = DefaultResolution
	}

	return time.NewTicker(time.Duration(resolution) * time.Second)
}

func debug(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}
