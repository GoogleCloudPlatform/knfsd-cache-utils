# Check the knfsd proxy instance is starting correctly

To check if a knfsd proxy instance is starting correctly, check the output from the serial port. In the output wait for "Finished running startup scripts.". This indicates that the start up scripts have completed.

Once the start up scripts have completed, look for either of these messages:

* "startup-script: Reached Proxy Startup Exit. Happy caching!".
* "startup-script: Error starting proxy"

## GCP Console

Using the GCP Console:

1. In the Google Cloud console go to the [Instance Groups](https://console.cloud.google.com/compute/instanceGroups/list) page.

2. Select the knfsd proxy instance group.

3. Select one of the knfsd proxy instances from the **Instance group members** list.

4. In the **Logs** section, select **Serial port 1 (console)**.

5. Search for "Finished running startup scripts.".

If the proxy instance is still starting, then you will need to refresh the serial port output until the start up script finishes running.

## gcloud

Using the `gcloud` command line gcloud:

Find an instance in the Managed Instance Group.

```sh
gcloud compute instance-groups managed list-instances INSTANCE_GROUP_NAME
```

Get the serial port output of the Compute Instance.

```sh
gcloud compute instances get-serial-port-output INSTANCE_NAME  | grep "google_metadata_script_runner"
```

If the instance is still starting up you can watch the serial port output in real time by running:

```sh
gcloud compute instances tail-serial-port-output INSTANCE_NAME | grep "google_metadata_script_runner"
```

## Common startup issues

There are three main reasons:

1. Invalid configuration (startup failed)
2. Firewall blocking GCP's health check services (startup successful, instance unhealthy)
3. Startup takes longer than 10 minutes (startup never finishes)

### Invalid configuration

If the start up shows the message "startup-script: Error starting proxy" then look at the last from "starup-script" to see what the error was.

The most likely errors are:

* Source NFS server could not be contacted:

  * Check the `EXPORT_MAP` or `EXPORT_HOST_AUTO_DETECT` has the correct IPs or DNS names.

  * If using DNS, check that the DNS names can be resolved by the knfsd proxy instances.

  * If GCP egress firewall rules exist, check these allow the knfsd proxy instances to access the source server.

  * Check on-prem firewall rules allow the NFS traffic from the knfsd proxy instances.

  * Check the VPN or interconnect is working.

  * Check if traffic from the knfsd proxy's subnet can be routed to the VPN or interconnect.

* Access denied by server while mounting share. This is either permission issues, or share does not exist. Check source export paths specified in `EXPORT_MAP`.

### Firewall blocking GCP's health check servers

If `AUTO_CREATE_FIREWALL_RULES = false` then you need to create the firewall rules to allow the GCP health check servers to access the knfsd proxy instances.

Setting `AUTO_CREATE_FIREWALL_RULES = false` is advised for production, or when deploying multiple knfsd proxy clusters to the same GCP project as the rules only need to be defined once per VPC network.

See [Firewall configuration](./firewall.md).

### Start up takes longer than 10 minutes

Normally 10 minutes is long enough for the proxy to start. However if you are re-exporting a large number of exports (e.g. 1000+) this start up time may exceed 10 minutes.

In this case the best option is to split up the exports across multiple separate knfsd proxy clusters to reduce the number of exports per knfsd proxy cluster.

If the startup time cannot be reduced below 10 minutes, then try starting the knfsd proxy cluster with `ENABLE_AUTOHEALING_HEALTHCHECKS = false` and measure how long the instances take to start. Then re-enable the health checks and configure `HEALTHCHECK_INITIAL_DELAY_SECONDS` to increase the initial grace period when starting the cluster.
