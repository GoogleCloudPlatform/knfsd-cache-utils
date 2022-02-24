# Next

* Replaced Stackdriver agent with Cloud Ops Agent
* Changed KNFSD Metrics Agent to use OpenTelemetry
* Changed custom metric types

## Replaced Stackdriver agent with Cloud Ops Agent

The Stackdriver Agent is obsolete. The last supported Ubuntu version is 20.04 LTS.

## Changed KNFSD Metrics Agent to use OpenTelemetry

The previous KNFSD Metrics Agent relied on reporting metrics via collectd (using the Stackdriver Agent).

Update the KNFSD metrics agent to use same OpenTelemetry Collector as the Cloud Ops Agent. This allows including additional metadata such as separating out the server and path for NFS mounts.

## Changed custom metric types

A limitation of the old collectd based KNFSD Metrics Agent is that all gauges had to be floats.

The new KNFSD Metrics Agent can now report gauges such as `knfsd/nfs_connections` using the correct data type (integers).

You will need to apply the [knfsd-cache-utils/deployment/metrics/](../../deployment/metrics/) Terraform to update the custom metrics.
