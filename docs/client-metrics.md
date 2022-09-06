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

## Example

The following example demonstrates how to install the `knfsd-metrics-agent` on an existing client machine.

Steps marked with "(local)" should be executed from your local machine (or your Cloud Shell in GCP).

Steps marked with "(remote)" should be executed on the remote (client) machine.

### (local) Configure environment

Set the following environment variables for use with the rest of the script (change the values to match your environment).

```bash
export PROJECT=example-project
export ZONE=us-central1-a
export VM_NAME=example-vm
```

### (local) Create a client VM in GCP (Optional)

Normally these commands would be included as part of an existing process to build a client VM image.

To test these commands before including them to an existing process you can:

* Use an existing client VM (if so skip this step)
* Follow this step to create a new VM

Configure the name of the network and subnet the VM should use:

```bash
export NETWORK_NAME=default
export SUBNET_NAME=default
```

Create the VM:

```bash
gcloud compute instances create "$VM_NAME" \
    --project="$PROJECT" \
    --zone="$ZONE" \
    --network="$NETWORK_NAME" \
    --subnet="$SUBNET_NAME" \
    --image-project="ubuntu-os-cloud" \
    --image-family="ubuntu-2004-lts"
```

Once you have finished testing the KNFSD Metrics Agent, delete the VM:

```bash
gcloud compute instances delete "$VM_NAME" --project="$PROJECT" --zone="$ZONE"
```

### (local) Checkout `knfsd-cache-utils`

```bash
git clone https://github.com/GoogleCloudPlatform/knfsd-cache-utils.git
cd knfsd-cache-utils
git checkout v0.9.0
```

### (local) Copy KNFSD Metrics Agent source code to the VM

```bash
tar -cz -f knfsd-metrics-agent.tar.gz -C image/resources knfsd-metrics-agent
```

```bash
gcloud compute scp ./knfsd-metrics-agent.tar.gz "$VM_NAME:" \
    --project="$PROJECT" \
    --zone="$ZONE"
```

### (local) Connect to client VM

```bash
gcloud compute ssh "$VM_NAME" --project="$PROJECT" --zone="$ZONE"
```

### (remote) Download and install Go 1.17

There are other methods to install Go. For simplicity, this example will install Go based on the the [standard Go instructions](https://go.dev/doc/install).

```bash
curl curl -fSLO https://go.dev/dl/go1.17.8.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.17.8.linux-amd64.tar.gz
```

Add Go to your path:

```bash
export PATH="$PATH:/usr/local/go/bin"
```

Verify Go is installed:

```bash
go version
```

### (remote) Build the KNFSD Metrics Agent

```bash
tar -xzf knfsd-metrics-agent.tar.gz
cd knfsd-metrics-agent
go build -v
```

### (remote) Install the KNFSD Metrics Agent

```bash
sudo chown root:root knfsd-metrics-agent
sudo mv knfsd-metrics-agent /usr/local/bin
sudo mkdir /etc/knfsd-metrics-agent
sudo cp config/*.yaml /etc/knfsd-metrics-agent
sudo cp systemd/client.service /etc/systemd/system/knfsd-metrics-agent.service
```

### (remote) Enable the KNFSD Metrics Agent

```bash
sudo systemctl enable --now knfsd-metrics-agent.service
```

### (remote) Test the KNFSD Metrics Agent (Optional)

Check the KNFSD Metrics Agent is running:

```bash
systemctl -o cat status knfsd-metrics-agent.service
```

You should see output similar to:

```text
 knfsd-metrics-agent.service - Knfsd Metrics Agent
     Loaded: loaded (/etc/systemd/system/knfsd-metrics-agent.service; enabled; vendor preset: enabled)
     Active: active (running) since Mon 2022-03-28 11:06:44 UTC; 8s ago
   Main PID: 6220 (knfsd-metrics-a)
      Tasks: 6 (limit: 4394)
     Memory: 12.7M
     CGroup: /system.slice/knfsd-metrics-agent.service
             └─6220 /usr/local/bin/knfsd-metrics-agent --config /etc/knfsd-metrics-agent/common.yaml --config /etc/knfsd-metrics-agent/client.yaml

2022-03-28T11:06:44.392Z        info    builder/pipelines_builder.go:65 Pipeline is started.    {"name": "pipeline", "name": "metrics"}
2022-03-28T11:06:44.392Z        info    service/service.go:97   Starting receivers...
2022-03-28T11:06:44.392Z        info    builder/receivers_builder.go:68 Receiver is starting... {"kind": "receiver", "name": "mounts"}
2022-03-28T11:06:44.392Z        info    builder/receivers_builder.go:73 Receiver started.       {"kind": "receiver", "name": "mounts"}
2022-03-28T11:06:44.393Z        info    builder/receivers_builder.go:68 Receiver is starting... {"kind": "receiver", "name": "slabinfo"}
2022-03-28T11:06:44.395Z        info    builder/receivers_builder.go:73 Receiver started.       {"kind": "receiver", "name": "slabinfo"}
2022-03-28T11:06:44.395Z        info    service/telemetry.go:87 Skipping telemetry setup.       {"address": ":8889", "level": "none"}
2022-03-28T11:06:44.395Z        info    service/collector.go:229        Starting knfsd-metrics-agent... {"Version": "", "NumCPU": 1}
2022-03-28T11:06:44.395Z        info    service/collector.go:124        Everything is ready. Begin running and processing data.
```

Before viewing the metrics you will need to generate some NFS traffic. Mount an NFS share (via a KNFSD Proxy instance) and then read some data from the share.

For example, use the `dd` command to read a file from the NFS share.

```bash
dd if=/mnt/nfs/share/example.file of=/dev/null bs=1M iflag=odirect
```

To view the metrics:

* Go to [Metrics explorer](https://console.cloud.google.com/monitoring/metrics-explorer)

* Select one of the KNFSD metrics (e.g. `custom.googleapis.com/knfsd/mount/read_bytes`, "VM Instance > Custom > NFS Mount Read Bytes").

  **NOTE:** You may have to clear "Show only active resources & metrics" if no data has been reported for the custom metric. Data should be reported within a few minutes once the agent is running.

* Add a filter:
  * Label: `name`
  * Comparison: `= (equals)`
  * Value: VM_NAME

* Change the chart type from Line chart to Stacked area chart.

* Click "Save Chart" to add this to a new or existing dashboard.
