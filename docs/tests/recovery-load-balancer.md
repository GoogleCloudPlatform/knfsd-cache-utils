# Recovering from failure between client and Load Balancer

This test plan investigates the behaviour of the clients when the proxy cluster is unavailable.

Some of the reasons the network could fail are:

* No firewall rule permitting traffic:
  * Client to Load Balancer
  * Load Balancer to proxy
* Network failure:
  * Client to Load Balancer
  * Load Balancer to proxy

To simulate a network error a firewall rule will be used to block all traffic between the clients and the Load Balancer.

## Procedure

* Set variables
* Create proxy cluster
* Create a firewall rule (disabled)
* SSH to client and run FIO
* Interrupt connection
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

```bash
Create deny firewall rule (disabled)
LOAD_BALANCER_IP="$(terraform output --raw load_balancer_ip_address)"
gcloud compute firewall-rules create \
    --project="${PROJECT}" \
    --network="${NETWORK}" \
    --priority=100 \
    --action=DENY \
    --direction=INGRESS \
    --rules=all \
    --destination-ranges="${LOAD_BALANCER_IP}" \
    --disabled \
    "${ZONE}-deny-nfs-to-lb"
```

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

When running the read test you can also set `--runtime=1200`  to increase the test time to 20 minutes.

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

```text
Starting 1 process
test: Laying out IO file (1 file / 1024MiB)
Jobs: 1 (f=1): [W(1)][2.0%][w=2048KiB/s][w=2 IOPS][eta 09m:48s]
Jobs: 1 (f=1): [W(1)][3.8%][w=1025KiB/s][w=1 IOPS][eta 09m:37s]
Jobs: 1 (f=1): [W(1)][4.3%][w=1025KiB/s][w=1 IOPS][eta 09m:34s]
```

### Interrupt connection

Once fio begins, run the following commands as a single batch (copy the '{}' to ensure bash handles it as a single input, even though it's over multiple lines).

The sleep commands will ensure fio is interrupted half way through. This will simulate either losing network to the source, or the source failing (either scenario is the same, the Load Balancer is unreachable).

When running a read test with `--runtime=1200` increase the interruption time from 300 seconds (5 minutes) to 900 seconds (15 minutes). This assumes the proxy is configured with `actimeo=600` (10 minutes).

```bash
{
echo "Waiting 2 minutes"
sleep 120

# enabled the firewall rule to simulate the loss of the source
echo "Denying traffic to source"
gcloud compute firewall-rules update \
    --project="${PROJECT}" \
    --no-disabled \
     "${ZONE}-deny-nfs-to-lb"

echo "Waiting 5 minutes"
sleep 300

# disable the firewall rule to simulate the source being recovered
echo "Recovering source"
gcloud compute firewall-rules update \
    --project="${PROJECT}" \
    --disabled \
    "${ZONE}-deny-nfs-to-lb"
}
```

### Wait

Wait for fio to complete.

You should see something like:

```text
test: (groupid=0, jobs=1): err= 0: pid=63676: Wed Nov 17 09:28:30 2021
  write: IOPS=0, BW=936KiB/s (959kB/s)(656MiB/717505msec); 0 zone resets


Run status group 0 (all jobs):
  WRITE: bw=936KiB/s (959kB/s), 936KiB/s-936KiB/s (959kB/s-959kB/s), io=656MiB (688MB), run=717505-717505msec
```

In the console running the interrupt command you should see:

```text
Waiting 2 minutes
Denying traffic to source
Updated [https://www.googleapis.com/compute/v1/projects/my-project/global/firewalls/smoke-tests-deny-nfs-to-source].
Waiting 5 minutes
Recovering source
Updated [https://www.googleapis.com/compute/v1/projects/my-project/global/firewalls/smoke-tests-deny-nfs-to-source].
```

## Results

### Write test

Write test runs for 10 minutes, with a 5 minute interruption after 2 minutes.

For the first 2 minutes, fio should show steady progress:

```text
Jobs: 1 (f=1): [W(1)][2.0%][w=2048KiB/s][w=2 IOPS][eta 09m:48s]
```

When the traffic is blocked fio will stop showing any write throughput or IOPS:

```text
Jobs: 1 (f=1): [W(1)][77.2%][w=2050KiB/s][w=2 IOPS][eta 02m:17s]
Jobs: 1 (f=1): [W(1)][79.0%][eta 02m:06s]
Jobs: 1 (f=1): [W(1)][80.8%][eta 01m:55s]
Jobs: 1 (f=1): [W(1)][82.7%][eta 01m:44s]
Jobs: 1 (f=1): [W(1)][84.5%][eta 01m:33s]
```

The percentage will continue to increment as the eta decrements. Once the traffic is allowed again, fio should continue until it completes:

```text
Jobs: 1 (f=1): [W(1)][64.3%][eta 03m:34s]
Jobs: 1 (f=1): [W(1)][66.2%][w=2050KiB/s][w=2 IOPS][eta 03m:23s]
Jobs: 1 (f=1): [W(1)][68.0%][w=1025KiB/s][w=1 IOPS][eta 03m:12s]
Jobs: 1 (f=1): [W(1)][69.8%][w=1025KiB/s][w=1 IOPS][eta 03m:01s]


write: IOPS=0, BW=936KiB/s (959kB/s)(656MiB/717505msec); 0 zone resets
```

Read test (10 minutes)
Read test runs for 10 minutes with a 5 minute interruption after 2 minutes.

After two minutes the file should be fully cached in the proxy, as such the client is able to continue reading the file during the interruption.

Read test (20 minutes)
Read test runs for 20 minutes with a 15 minute interruption after 2 minutes.

Similar to the 10 minute test, initially the client will continue reading after the interruption. However, because the interruption is longer than the proxy's metadata cache time (`actimeo`, or `acmaxreg`) the metadata eventually becomes invalid.

Once the metadata has become invalid the read operation will block the same as observed in the write test until the connection is resumed.

```text
Jobs: 1 (f=1): [W(1)][0.3%][w=1025KiB/s][w=1 IOPS][eta 59m:49s]
Jobs: 1 (f=1): [W(1)][0.6%][eta 59m:38s]
Jobs: 1 (f=1): [W(1)][0.9%][eta 59m:27s]
Jobs: 1 (f=1): [W(1)][1.2%][eta 59m:16s]


Jobs: 1 (f=1): [W(1)][10.1%][eta 53m:57s]
Jobs: 1 (f=1): [W(1)][10.4%][w=46.0MiB/s][w=46 IOPS][eta 53m:46s]
Jobs: 1 (f=1): [W(1)][10.7%][w=1025KiB/s][w=1 IOPS][eta 53m:35s]
Jobs: 1 (f=1): [W(1)][11.0%][w=2048KiB/s][w=2 IOPS][eta 53m:24s]
Jobs: 1 (f=1): [W(1)][11.3%][w=1025KiB/s][w=1 IOPS][eta 53m:13s]
```

The connection resumed reasonably quickly when running this test. However, it might take up to 10 minutes to recover, if so the runtime for fio may need to be increased from 20 minutes to 30 minutes.

### Clean up

First remove the firewall rule that was created manually:

```bash
gcloud compute firewall-rules delete \
  --project="${PROJECT}" \
  "${ZONE}-deny-nfs-to-lb"
```

Destroy the remaining resources using Terraform.

### Conclusion

When both the proxy and the client are mounted using the hard option (recommended) any NFS operations between the client and the proxy will wait for the Load Balancer to recover.

No manual intervention on the Load Balancer or the clients is required.

The connection does not recover immediately. NFS has a linear back off on retries based on the `timeo` setting, with a maximum retry time of 600 seconds.
