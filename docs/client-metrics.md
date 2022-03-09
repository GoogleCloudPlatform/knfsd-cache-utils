# Client Metrics

The KNFSD Metrics Agent is designed so that it can be used on clients to collect additional metrics to indicate the health of the NFS proxies.

The primary metrics from the clients is the round trip (RTT) and execution (exe) times of the read/write requests. These indicate the total latency of client requests. At first the latency will be high as all requests need to go back to the source server. As more data is cached the latency of read request should reduce as more requests are answered by the proxy.

These instructions only include the custom NFS metrics collected by the KNFSD Metrics Agent. For standard metrics, including network throughput, install the standard metrics agent for your system. On GCP the [Cloud Ops Agent](https://cloud.google.com/stackdriver/docs/solutions/agents/ops-agent) is recommended.

## Prerequisites

To build the metrics agent you will need [go 1.17](https://go.dev/).

## Building the agent

```bash
cd image/resources/knfsd-metrics-agent
go build
```

Go will automatically download all the modules required by the package.

The build will produce a single binary named `knfsd-metrics-agent`.

### Building from secure networks

If you cannot access the Internet to fetch packages there are two main options:

* Use [GOPROXY](https://go.dev/ref/mod#module-proxy) a with a private package repository.

* Use `go mod vendor` from a machine that does have Internet access, see https://go.dev/ref/mod#vendoring. This will create a vendor directory with all the source code required that can then be checked into source control, or copied to the build machine.

Configuring a private package repository, or setting up vendoring is beyond the scope of this document and depends upon your specific environment.

## Configuring

The agent is configured using YAML config files, see the [KNFSD Metric Agent documentation](../image/resources/knfsd-metrics-agent/README.md) for the available configuration.

On GCP a sample configuration is provided to get started, this is split into two config files, [common.yaml](../image/resources/knfsd-metrics-agent/config/common.yaml) and [client.yaml](../image/resources/knfsd-metrics-agent/config/client.yaml).

The Agent can be started from the terminal with following command:

```bash
sudo ./knfsd-metrics-agent --config config/common.yaml --config config/client.yaml
```

When running from the terminal to test the configuration it can be useful to change the `collection_interval` to `10s` and add the `logging` exporter to the pipeline.

### Other Environments

`common.yaml` and `client.yaml` are intended for running on GCP and reporting to GCP Cloud Monitoring.

If you reporting to other systems such as Elasticsearch, or wish to run on other platforms such as on-prem then you can use these files as a template to get started.

The main elements you will need to reconfigure are:

* `resourcedetection.detectors` if you're running on a different platform. Alternatively, remove `resourcedetection` from the pipeline completely if your platform is not supported by the `resourcedetection` processor.

* Add a new exporter such as `prometheus`:
  * Add config for the exporter to the `exporters` section
  * Remove `googlecloud` from the list of exporters in the pipeline
  * Add the new exporter to the pipeline

* `metricstransform` processor. This processor can be used to rename the metrics and/or attributes to match the naming convention of your platform.

## Installing

Installing the agent on a client assumes the client uses Linux operating system with systemd.

* Copy `knfsd-metrics-agent` to `/usr/local/bin`.
* Create the config directory `/etc/knfsd-metrics-agent`
* Copy the config files to `/etc/knfsd-metrics-agent`
* Copy and rename the [systemd/client.service](../image/resources/knfsd-metrics-agent/systemd/client.service) file to `/etc/systemd/system/knfsd-metrics-agent.service`
* Enable the agent: `systemctl enable knfsd-metrics-agent.service`

These installation steps should be included into your client image building process.

If running the commands by hand to test the process before building an image remember to start the service:

* `systemctl start knfsd-metrics-agent.service`

To check the agent started successfully:

* `systemctl status knfsd-metrics-agent.service`
* `journalctl -o cat -u knfsd-metrics-agent.service`
