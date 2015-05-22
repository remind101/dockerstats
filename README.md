# Dockerstats

Dockerstats is a simple docker container for collecting metrics from the [Docker Stats](https://docs.docker.com/reference/api/docker_remote_api_v1.17/#get-container-stats-based-on-resource-usage) API and sending them elsewhere. It supports adapters called `Drains` for configuring where you want metrics to be sent.

Currently, the following drains are provided:

* **L2Met**: An adapter that logs metrics in l2met format. This is useful if you're using something like logspout to collect all logs from containers and send them to l2met.

## Usage

Simply run the container and mount the docker socket:

```console
$ docker run --name="dockerstats" \
    --volume=/var/run/docker.sock:/tmp/docker.sock \
    remind101/dockerstats
```

## Metrics

The following metrics will be created:

```
Network.RxDropped
Network.RxBytes
Network.RxErrors
Network.TxPackets
Network.RxPackets
Network.TxErrors
Network.TxBytes

MemoryStats.Stats.TotalPgmafault
MemoryStats.Stats.Cache
MemoryStats.Stats.MappedFile
MemoryStats.Stats.TotalInactiveFile
MemoryStats.Stats.Pgpgout
MemoryStats.Stats.Rss
MemoryStats.Stats.TotalMappedFile
MemoryStats.Stats.Writeback
MemoryStats.Stats.Unevictable
MemoryStats.Stats.Pgpgin
MemoryStats.Stats.TotalUnevictable
MemoryStats.Stats.Pgmajfault
MemoryStats.Stats.TotalRss
MemoryStats.Stats.TotalRssHuge
MemoryStats.Stats.TotalWriteback
MemoryStats.Stats.TotalInactiveAnon
MemoryStats.Stats.RssHuge
MemoryStats.Stats.HierarchicalMemoryLimit
MemoryStats.Stats.TotalPgfault
MemoryStats.Stats.TotalActiveFile
MemoryStats.Stats.ActiveAnon
MemoryStats.Stats.TotalActiveAnon
MemoryStats.Stats.TotalPgpgout
MemoryStats.Stats.TotalCache
MemoryStats.Stats.InactiveAnon
MemoryStats.Stats.ActiveFile
MemoryStats.Stats.Pgfault
MemoryStats.Stats.InactiveFile
MemoryStats.Stats.TotalPgpgin
MemoryStats.MaxUsage
MemoryStats.Usage
MemoryStats.Failcnt
MemoryStats.Limit

CPUStats.CPUUsage.UsageInUsermode
CPUStats.CPUUsage.TotalUsage
CPUStats.CPUUsage.UsageInKernelmode
CPUStats.SystemCPUUsage
CPUStats.ThrottlingData.Periods
CPUStats.ThrottlingData.ThrottledPeriods
CPUStats.ThrottlingData.ThrottledTime
```

## Roadmap

* Add a statsd drain.
* BlkioStats
