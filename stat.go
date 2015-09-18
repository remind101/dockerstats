package dockerstats

// stat represents a single stat and is provided as the context to a
import (
	"os"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

// template.
type stat struct {
	Container *docker.Container
	Type      string
	Name      string
	Value     interface{}
}

func (s stat) Hostname() string {
	return hostname
}

// Returns the first 12 characters of the container ID.
func (s stat) ID() string {
	return s.Container.ID[:12]
}

func (s stat) Env(key string) string {
	for _, env := range s.Container.Config.Env {
		if strings.HasPrefix(env, key+"=") {
			return env[len(key)+1:]
		}
	}
	return ""
}
