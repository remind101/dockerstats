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
	"github.com/remind101/empire/pkg/dockerutil"
)

// DefaultResolution defines the default resolution for draining stats. We
// default to 10 seconds.
var DefaultResolution = 10

var eventList = map[string]bool{
	// Events that should be handled.
	"create":      true,
	"destroy":     true,
	"die":         true,
	"exec_create": true,
	"exec_start":  true,
	"export":      true,
	"kill":        true,
	"oom":         true,
	"pause":       true,
	"restart":     true,
	"start":       true,
	"stop":        true,
	"unpause":     true,

	// Events that should not be handled
	// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#monitor-docker-s-events
	"untag":  false,
	"delete": false,
}

// Stats represents a set of stats from a container at a given point in time.
type Stats struct {
	*docker.Stats
	Container *docker.Container
}

// Event represents an event for a container.
type Event struct {
	*docker.APIEvents
	Container *docker.Container
}

// Adapter is an interface for draining stats and events somewhere.
type Adapter interface {
	Stats(*Stats) error
	Event(*Event) error
}

// Stat is a context struct that manages the lifecycle of container metrics. It
// watches for container start, restart and stop events and streams metrics to a
// Drain.
type Stat struct {
	// Adapter is the Adapter that will be used to drain Stats and Events.
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
func New() (*Stat, error) {
	c, err := dockerutil.NewDockerClientFromEnv()
	if err != nil {
		return nil, err
	}

	return &Stat{
		client:     c,
		containers: make(map[string]*docker.Container),
	}, nil
}

// Run begins starts draining metrics and events from all of the currently
// running containers and starts watching for new containers to drain metrics
// and events from. This call is blocking.
func (s *Stat) Run() error {
	containers, err := s.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}

	for _, c := range containers {
		container, err := s.addContainer(c.ID)
		if err != nil {
			return err
		}

		go s.attachMetrics(container)
	}

	events := make(chan *docker.APIEvents)
	if err := s.client.AddEventListener(events); err != nil {
		return err
	}

	for event := range events {
		// Ignore events that are not whitelisted.
		if !eventList[event.Status] {
			continue
		}

		container, err := s.addContainer(event.ID)
		if err != nil {
			debug("add container: err: %s", err)
			continue
		}

		go s.event(container, event)

		switch event.Status {
		case "start", "restart":
			go s.attachMetrics(container)
		}
	}

	return errors.New("unexpected stop")
}

// addContainer adds the container to the internal map of known containers.
func (s *Stat) addContainer(containerID string) (*docker.Container, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if container, ok := s.containers[containerID]; ok {
		// We already know about this container.
		return container, nil
	}

	container, err := s.client.InspectContainer(containerID)
	if err != nil {
		debug("inspect: err: %s", err)
		return container, err
	}
	container.Name = strings.Replace(container.Name, "/", "", 1)

	s.containers[containerID] = container

	return container, nil
}

func (s *Stat) attachMetrics(container *docker.Container) {
	defer func() {
		if v := recover(); v != nil {
			debug("recovered panic in attachMetrics: %v", v)
		}
	}()

	debug("draining: %s", container.Name)

	stats := make(chan *docker.Stats)
	go func() {
		if err := s.client.Stats(docker.StatsOptions{
			ID:    container.ID,
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
			if err := s.stats(container, stat); err != nil {
				debug("stats: err: %s", err)
			}
		default:
			// Drop the stat.
		}
	}
}

func (s *Stat) stats(container *docker.Container, stats *docker.Stats) error {
	return s.adapter().Stats(&Stats{
		Stats:     stats,
		Container: container,
	})
}

func (s *Stat) event(container *docker.Container, event *docker.APIEvents) error {
	return s.adapter().Event(&Event{
		APIEvents: event,
		Container: container,
	})
}

func (s *Stat) adapter() Adapter {
	if s.Adapter == nil {
		return DefaultAdapter
	}

	return s.Adapter
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
