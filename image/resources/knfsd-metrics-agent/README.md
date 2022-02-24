# KNFSD Metrics Agent

This agent collects custom metrics about the operation of the NFS proxy. The agent can also support collecting NFS metrics from client instances, including enriching the metrics with the name of the NFS proxy instance the client is connected to.

## Pre-requisites

Before this collector can be used with Google Cloud Monitoring, you must first apply the [knfsd-cache-utils/deployment/metrics/](../../../deployment/metrics/).

## Plugins

The agent uses the OpenTelemetry Collector v0.44.0 and can support exporting metrics in several formats including Prometheus and Elasticsearch.

### Receivers

#### Connections

Reports on the number of incoming client connections to the NFS server. A connection is considered to be a client connection if the local TCP/UDP port is 2049.

See [connections/metadata.yaml](internal/connections/metadata.yaml)

* `collection_interval` (default = `1m`): This receiver collects metrics on an interval. Valid time units are ns, us, ms, s, m, h.

```yaml
receivers:
  connections:
    collection_interval: 1m
```

#### Mounts

Reports on NFS mount statistics such as round trip time between the local NFS mounts and the remote NFS server.

See [mounts/metadata.yaml](internal/mounts/metadata.yaml).

* `collection_interval` (default = 1m): This receiver collects metrics on an interval. Valid time units are ns, us, ms, s, m, h.

* `query_proxy_instances`:

  * `enabled` (default = `false`): Enables querying each source NFS server to resolved which proxy instance a client is connected to. This assumes the NFS server is running the `knfsd-agent`.

  * `timeout` (default = `10s`): HTTP timeout per source server, this timeout is the full round trip time, so includes establishing the connection, and reading the response. Valid time units are ns, us, ms, s, m, h.

```yaml
receivers:
  mounts:
    collection_interval: 1m
    query_proxy_instance:
      enabled: false
      timeout: 10s
```

#### Oldest File

Reports on the age of the oldest file in FS-Cache.

This is included to aid with diagnosing issues. The only way to find the oldest file is by recursively scanning all files under `/var/cache/fscache`. For this reason this metric is not included in the pipeline by default.

If you do add this to the pipeline, you may need to increase the `collection_interval` to reduce excessive load on the cache file system. On larger caches it might require increasing the interval to `1h`.

* `collection_interval` (default = `10m`): This receiver collects metrics on an interval. Valid time units are ns, us, ms, s, m, h.

* `cache_path` (default = `/var/cache/fscache/cache`): The path to the cachefilesd cache directory.

```yaml
receivers:
  oldestfile:
    collection_interval: 10m
    cache_path: /var/cache/fscache/cache
```

#### Slab

Reports NFS cache related slab metrics (i.e. NFS inode cache, dcache).

See [slab/metadata.yaml](internal/slab/metadata.yaml)

* `collection_interval` (default = `1m`): This receiver collects metrics on an interval. Valid time units are ns, us, ms, s, m, h.

```yaml
receivers:
  mounts:
    collection_interval: 1m
```

### Processors

* [Batch](https://pkg.go.dev/go.opentelemetry.io/collector@v0.44.0/processor/batchprocessor)
* [Memory Limiter](https://pkg.go.dev/go.opentelemetry.io/collector@v0.44.0/processor/memorylimiterprocessor)
* [Metrics Transform](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor@v0.44.0)
* [Resource Detection](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor@v0.44.0)
* [Resource](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor@v0.44.0)

### Exporters

* [Logging](https://pkg.go.dev/go.opentelemetry.io/collector@v0.44.0/exporter/loggingexporter)
* [OTLP](https://pkg.go.dev/go.opentelemetry.io/collector@v0.44.0/exporter/otlpexporter)
* [OTLP HTTP](https://pkg.go.dev/go.opentelemetry.io/collector@v0.44.0/exporter/otlphttpexporter)
* [Elastic](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticexporter@v0.44.0)
* [Elastic Search](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticsearchexporter@v0.44.0)
* [File](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter@v0.44.0)
* [Google Cloud](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter@v0.44.0)
* [Google Cloud Pub/Sub](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudpubsubexporter@v0.44.0)
* [InfluxDB](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/influxdbexporter@v0.44.0)
* [OpenCensus](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/opencensusexporter@v0.44.0)
* [Prometheus](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter@v0.44.0)
* [Prometheus Remote Write](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter@v0.44.0)
* [Stackdriver](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/exporter/stackdriverexporter@v0.44.0)

### Extensions

* [Ballast](https://pkg.go.dev/go.opentelemetry.io/collector@v0.44.0/extension/ballastextension)
* [zPages](https://pkg.go.dev/go.opentelemetry.io/collector@v0.44.0/extension/zpagesextension)

## Configuring

The agent is configured using one or more YAML config files.

```bash
knfsd-metrics-agent --config common.yaml --config proxy.yaml
```

The config files are loaded in the order specified, with values from the later config files overwriting the earlier config files. Object keys are merged, while arrays are replaced.

To configure a receiver, processor, or exporter it must first be defined in the appropriate section, then added to a pipeline.

If you want to include the same receiver, processor, or exporter with different configurations then you can use the format `type/instance`, e.g. `connections/debug`

```yaml
receivers:
  # Declare the receivers with default options.
  # Note the colon as these are objects.
  connections:
  mounts:
  slab:

  # Declare a second instance of the connections receiver, with a different
  # interval.
  connections/debug:
    collection_interval: 10s

processors:
  resourcedetection:
    detectors: [gce]

exporters:
  googlecloud:
    user_agent: knfsd-metrics-agent
    metric:
      prefix: ""
      skip_create_descriptor: true

  # Useful when running knfsd-metrics-agent from the command line, will write
  # metrics to stderr.
  logging:
    loglevel: debug

service:
  pipelines:
    metrics: # name of the pipeline
      receivers:
        - connections
        - mounts
        - slabinfo
      processors:
        - resourcedetection
      exporters:
        - googlecloud

    debug: # second pipeline
      receivers:
        - connections/debug
      processors:
        # can use receivers/processors/exporters in multiple pipelines
        - resourcedetection
      exporters:
        - logging
```
