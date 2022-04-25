# Next

* Increase health check interval to 60 seconds
* Add parameters to configure health checks
* Support deploying metrics as a Terraform module

## Increase health check interval to 60 seconds

This allows 2 minutes (with the default health check values) to reboot a knfsd proxy instance without the managed instance group replacing the instance.

## Add parameters to configure health checks

This allows overriding various parameters used by the health checks. For example, if you do not encounter the culling issue you might want to reduce the `HEALTHCHECK_INTERVAL_SECONDS` so that failed instances are detected more quickly.

If you have a lot of volumes, or high latency between the source and the proxy causing a startup time slower than 10 minutes (600 seconds), you might want to increase the `HEALTHCHECK_INITIAL_DELAY_SECONDS`. Conversely, if you know your proxy starts up in less than 5 minutes, you can reduce the initial delay so that instances that fail to start up correctly are detected and replaced more quickly.

## Support deploying metrics as a Terraform module

Support deploying the metrics as a Terraform module so that the metrics can be deployed without needing to clone the Terraform configuration from git.
