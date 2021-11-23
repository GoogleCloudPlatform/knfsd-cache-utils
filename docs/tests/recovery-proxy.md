# Recovering from failure of the proxy cluster

This test plan investigates the behaviour of the clients when the proxy cluster is unavailable.

Some of the reasons the entire cluster could fail are:

* Incorrect configuration scales down (or deletes) the MIG
* All the instances restart at the same time (e.g. due to update)
* After update instances fail to serve NFS traffic
* Instances restart while the source server is unavailable

The proxy instance group will be resized to zero instances to simulate the entire cluster failing.

## Procedure

* Set variables
* Create proxy cluster
* SSH to client and run FIO
* Resize instance group
* Wait
* Clean up

### Set variables

Update settings in brackets <> below and set variables

```bash
PROJECT=<my-project>
NETWORK=<my-network>
ZONE=<my-zone>
PROXY=<instance-group-name>
```

### Create proxy cluster

* [Build the NFS proxy image](../../image/README.md)
* [Deploy the proxy](../../deployment/README.md)

### SSH to client and run FIO

```bash
gcloud compute ssh --project="${PROJECT}" --zone="${ZONE}" client-1
```

```bash
sudo apt update
sudo apt install nfs-common fio
sudo mkdir /mnt/proxy
sudo mount 10.164.15.195:/mnt /mnt/proxy -o vers=3,rw,sync,hard,noatime,proto=tcp,mountproto=tcp
sudo mkdir -m 777 /mnt/proxy/test
```

For this test we will use fio to perform a constant write via the proxy. You can repeat this test using `--rw=read` to test the read behaviour.

When running the read test you can also set `--runtime=1200` to increase the test time to 20 minutes.

```bash
fio \
    --name=test \
    --directory=/mnt/proxy/test \
    --ioengine=libaio \
    --iodepth=64 \
    --direct=1 \
    --rw=write \
    --bs=1Mi \
    --size=1Gi \
    --time_based \
    --runtime=600 \
    --eta-newline=10
```

When running a read test you need to wait until fio has finished laying out the file and the read test begins. Fio may skip this step if the test file already exists and is the correct size.

```text
Starting 1 process
test: Laying out IO file (1 file / 1024MiB)
Jobs: 1 (f=1): [W(1)][2.0%][w=2048KiB/s][w=2 IOPS][eta 09m:48s]
Jobs: 1 (f=1): [W(1)][3.8%][w=1025KiB/s][w=1 IOPS][eta 09m:37s]
Jobs: 1 (f=1): [W(1)][4.3%][w=1025KiB/s][w=1 IOPS][eta 09m:34s]
```

Every 10 seconds (approximate) fio will write a new status line. This will make it easier to see the interruption in the console output at the end.

## Resize instance group

Once fio begins, resize the cluster to zero. The resize command will simulate either losing connection to the proxy, or the proxy failing.

When running a read test with `--runtime=1200` increase the interruption time from 300 seconds (5 minutes) to 900 seconds (15 minutes). This assumes the proxy is configured with `actimeo=600` (10 minutes).

```bash
gcloud compute instance-groups managed resize \
      "${PROXY}" \
      --size=0 \
      --zone="${ZONE}"
```

Once the instance group is resized to zero, scale the instance group back up to one to simulate the proxy recovering. This will test if the client will reconnect to the new instance.

```bash
gcloud compute instance-groups managed resize \
      "${PROXY}" \
      --size=1 \
      --zone="${ZONE}"
```

### Wait

Wait for fio to complete.

You should see something like:

```text
test: (groupid=0, jobs=1): err= 0: pid=63676: Wed Nov 17 09:28:30 2021
  write: IOPS=0, BW=936KiB/s (959kB/s)(656MiB/717505msec); 0 zone resets
```

```bash
Run status group 0 (all jobs):
  WRITE: bw=936KiB/s (959kB/s), 936KiB/s-936KiB/s (959kB/s-959kB/s), io=656MiB (688MB), run=717505-717505msec
```

### Clean up

Destroy the remaining resources using Terraform.

### Conclusion

When both the proxy and the client are mounted using the `hard` option (recommended) any NFS operations between the client and the proxy will wait for the proxy to recover.

The load balancer provides a static IP for the clients, so the proxy instances can be recreated with a different IP address without affecting the clients or requiring manual intervention.

The connection does not recover immediately. NFS has a linear back off on retries based on the `timeo` setting, with a maximum retry time of 600 seconds.
