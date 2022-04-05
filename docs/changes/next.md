# Next

* Abort mounting export after 3 attempts
* Custom GCP labels for proxy VM instances

## Abort mounting export after 3 attempts

Only try to mount the same export a maximum of 3 times (with 60 seconds between each attempt).

If the attempts fail the startup script will be aborted and the NFS server will not be started.

When the health check is enabled, after 10 minutes the proxy instance will be replaced.

## Custom GCP labels for proxy VM instances

Added a new `PROXY_LABELS` variable to set custom labels on the proxy VM instances. This can aid with filtering metrics and logs when running multiple proxy clusters in a single project.
