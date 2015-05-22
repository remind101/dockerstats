// package stat is a Go program for watching the container stats endpoint in the
// docker api and shuttling the metrics somwhere else where they belong.
// Basically logspout for container metrics.
package stat

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/fsouza/go-dockerclient"
)

// Drain is an interface for draining metrics somewhere.
type Drain interface {
	Drain(*docker.Container, *docker.Stats) error
}

// Stat is a context struct that manages the lifecycle of container metrics. It
// watches for container start, restart and stop events and streams metrics to a
// Drain.
type Stat struct {
	Drain

	mu         sync.Mutex
	containers map[string]chan *docker.Stats
	client     *docker.Client
}

// New returns a new Stat instance with a configured docker client.
func New(host string) (*Stat, error) {
	c, err := docker.NewClient(host)
	if err != nil {
		return nil, err
	}

	return &Stat{
		client: c,
	}, nil
}

// Start begins starts draining metrics from all of the currently running
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
	defer s.mu.Unlock()

	if _, ok := s.containers[containerID]; ok {
		// We're already watching metrics from this container. Nothing
		// to do.
		return
	}

	container, err := s.client.InspectContainer(containerID)
	if err != nil {
		debug("inspect: err: %s", err)
		return
	}

	stats := make(chan *docker.Stats)
	go func() {
		if err := s.client.Stats(docker.StatsOptions{
			ID:    containerID,
			Stats: stats,
		}); err != nil {
			debug("stats: err: %s", err)
		}
		close(stats)
	}()

	for stat := range stats {
		if err := s.drain(container, stat); err != nil {
			debug("drain: err: %s", err)
		}
	}
}

func (s *Stat) stop(containerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.containers[containerID]; ok {
		close(ch)
	}
}

func (s *Stat) drain(container *docker.Container, stats *docker.Stats) error {
	if s.Drain == nil {
		return nil
	}

	return s.Drain.Drain(container, stats)
}

func debug(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
}
