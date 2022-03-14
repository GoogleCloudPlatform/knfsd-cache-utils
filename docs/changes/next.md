# Next

* Abort mounting export after 3 attempts

## Abort mounting export after 3 attempts

Only try to mount the same export a maximum of 3 times (with 60 seconds between each attempt).

If the attempts fail the startup script will be aborted and the NFS server will not be started.

When the health check is enabled, after 10 minutes the proxy instance will be replaced.
