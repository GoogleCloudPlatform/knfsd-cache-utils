# Recovering from failure between source and proxy

This test plan investigates the behaviour of the clients when the source server is unavailable.

A firewall rule will be used to block all traffic to the source server (Filestore). This simulates both the source server crashing, or a network interruption. From the perspective of the proxy or the clients, both of these errors are identical; the source server is unreachable.

## Procedure

* Setup
* Create NFS proxy cluster
* Create deny firewall rule (disabled)
* SSH to client and run FIO
* Enable deny firewall rule
* Wait
* Disable deny firewall rule
* Clean up

### Setup

```bash
PROJECT=my-project
PREFIX=smoke-tests
ZONE=us-central1-a
```

### Create NFS proxy cluster

Use the smoke-tests Terraform to create a basic 1 node cluster with a client using Filestore as the source.

### Create deny firewall rule (disabled)

```bash
SOURCE_IP="$(terraform output --raw source_ip)"
gcloud compute firewall-rules create \
    --project="${PROJECT}" \
    --network="${PREFIX}" \
    --priority=100 \
    --action=DENY \
    --direction=EGRESS \
    --rules=all \
    --destination-ranges="${SOURCE_IP}" \
    --disabled \
    "${PREFIX}-deny-nfs-to-source"
```

### SSH to client and run FIO

```bash
gcloud compute ssh --project="${PROJECT}" --zone="${ZONE}" smoke-tests-client
```

```bash
sudo apt update
sudo apt install nfs-common fio
sudo mkdir /mnt/proxy
sudo mount 10.0.0.2:/files /mnt/proxy -o vers=3,rw,sync,hard,noatime,proto=tcp,mountproto=tcp
sudo mkdir -m 777 /mnt/proxy/test
```

For this test we will use fio to perform a constant write via the proxy. You can repeat this test using `--rw=read` to test the read behaviour.

When running the read test you can also set `--runtime=1200` to increase the  test time to 20 minutes.

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
test: Laying out IO file (1 file / 1024MiB)
Jobs: 1 (f=1): [R(1)][2.0%][r=1781MiB/s][r=1780 IOPS][eta 09m:48s]
Jobs: 1 (f=1): [R(1)][3.8%][r=1818MiB/s][r=1818 IOPS][eta 09m:37s]
Jobs: 1 (f=1): [R(1)][5.7%][r=1772MiB/s][r=1771 IOPS][eta 09m:26s]
```

Every 10 seconds (approximate) fio will write a new status line. This will make it easier to see the interruption in the console output at the end.

### Interrupt connection

Once fio begins, run the following commands as a single batch (copy the '{}' to ensure bash handles it as a single input, even though it's over multiple lines).

The sleep commands will ensure fio is interrupted half way through. This will simulate either losing network to the source, or the source failing (either scenario is the same, the source is unreachable).

When running a read test with `--runtime=1200` increase the interruption time from `300` seconds (5 minutes) to `900` seconds (15 minutes). This assumes the proxy is configured with `actimeo=600` (10 minutes).

```bash
{
echo "Waiting 2 minutes"
sleep 120

# enabled the firewall rule to simulate the loss of the source
echo "Denying traffic to source"
gcloud compute firewall-rules update \
    --project="${PROJECT}" \
    --no-disabled \
     "${PREFIX}-deny-nfs-to-source"

echo "Waiting 5 minutes"
sleep 300

# disable the firewall rule to simulate the source being recovered
echo "Recovering source"
gcloud compute firewall-rules update \
    --project="${PROJECT}" \
    --disabled \
    "${PREFIX}-deny-nfs-to-source"
}
```

### Wait

Wait for fio to complete.

You should see something like:

```text
test: (groupid=0, jobs=1): err= 0: pid=4289: Thu Nov 11 17:01:19 2021
  read: IOPS=839, BW=840MiB/s (880MB/s)(984GiB/1200024msec)
...
Run status group 0 (all jobs):
   READ: bw=840MiB/s (880MB/s), 840MiB/s-840MiB/s (880MB/s-880MB/s), io=984GiB (1057GB), run=1200024-1200024msec
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

### Clean up

First remove the firewall rule that was created manually:

```bash
gcloud compute firewall-rules delete \
  --project="${PROJECT}" \
  "${PREFIX}-deny-nfs-to-source"
```

Destroy the remaining resources using Terraform.

## Results

### Write test

Write test runs for 10 minutes, with a 5 minute interruption after 2 minutes.

For the first 2 minutes, fio should show steady progress:

```text
Jobs: 1 (f=1): [W(1)][3.8%][w=84.1MiB/s][w=84 IOPS][eta 09m:37s]
```

When the traffic is blocked fio will stop showing any write throughput or IOPS:

```text
Jobs: 1 (f=1): [W(1)][20.3%][w=95.0MiB/s][w=95 IOPS][eta 07m:58s]
Jobs: 1 (f=1): [W(1)][22.2%][eta 07m:47s]
Jobs: 1 (f=1): [W(1)][24.0%][eta 07m:36s]
Jobs: 1 (f=1): [W(1)][25.8%][eta 07m:25s]
```

The percentage will continue to increment as the eta decrements.
Once the traffic is allowed again, fio should continue until it completes:

```text
Jobs: 1 (f=1): [W(1)][68.0%][eta 03m:12s]
Jobs: 1 (f=1): [W(1)][69.8%][eta 03m:01s]
Jobs: 1 (f=1): [W(1)][71.7%][eta 02m:50s]
Jobs: 1 (f=1): [W(1)][73.5%][w=86.0MiB/s][w=86 IOPS][eta 02m:39s]
Jobs: 1 (f=1): [W(1)][75.3%][w=94.0MiB/s][w=94 IOPS][eta 02m:28s]
...
Jobs: 1 (f=1): [W(1)][97.5%][w=96.1MiB/s][w=96 IOPS][eta 00m:15s]
Jobs: 1 (f=1): [W(1)][99.2%][w=93.1MiB/s][w=93 IOPS][eta 00m:05s]
Jobs: 1 (f=1): [W(1)][100.0%][w=91.0MiB/s][w=91 IOPS][eta 00m:00s]
test: (groupid=0, jobs=1): err= 0: pid=4259: Thu Nov 11 16:11:58 2021
  write: IOPS=44, BW=44.7MiB/s (46.8MB/s)(26.2GiB/600681msec); 0 zone resets
```

### Read test (10 minutes)

Read test runs for 10 minutes with a 5 minute interruption after 2 minutes.

After two minutes the file should be fully cached in the proxy, as such the client is able to continue reading the file during the interruption.

### Read test (20 minutes)

Read test runs for 20 minutes with a 15 minute interruption after 2 minutes.

Similar to the 10 minute test, initially the client will continue reading after the interruption. However, because the interruption is longer than the proxy's metadata cache time (`actimeo`, or `acmaxreg`) the metadata eventually becomes invalid.

Once the metadata has become invalid the read operation will block the same as observed in the write test until the connection is resumed.

```text
Jobs: 1 (f=1): [R(1)][31.1%][r=1807MiB/s][r=1807 IOPS][eta 13m:47s]
Jobs: 1 (f=1): [R(1)][32.0%][r=1751MiB/s][r=1750 IOPS][eta 13m:36s]
Jobs: 1 (f=1): [R(1)][32.9%][r=1792MiB/s][r=1792 IOPS][eta 13m:25s]
Jobs: 1 (f=1): [R(1)][33.8%][eta 13m:14s]
Jobs: 1 (f=1): [R(1)][34.8%][eta 13m:03s]
Jobs: 1 (f=1): [R(1)][35.7%][eta 12m:52s]
...
Jobs: 1 (f=1): [R(1)][86.1%][eta 02m:47s]
Jobs: 1 (f=1): [R(1)][87.0%][eta 02m:36s]
Jobs: 1 (f=1): [R(1)][87.9%][eta 02m:25s]
Jobs: 1 (f=1): [R(1)][88.8%][r=1837MiB/s][r=1836 IOPS][eta 02m:14s]
Jobs: 1 (f=1): [R(1)][89.2%][r=2030MiB/s][r=2030 IOPS][eta 02m:10s]
Jobs: 1 (f=1): [R(1)][90.6%][r=1919MiB/s][r=1918 IOPS][eta 01m:53s]
...
Jobs: 1 (f=1): [R(1)][98.8%][r=1828MiB/s][r=1828 IOPS][eta 00m:14s]
Jobs: 1 (f=1): [R(1)][99.8%][r=1781MiB/s][r=1780 IOPS][eta 00m:03s]
Jobs: 1 (f=1): [R(1)][100.0%][r=1876MiB/s][r=1876 IOPS][eta 00m:00s]
test: (groupid=0, jobs=1): err= 0: pid=4289: Thu Nov 11 17:01:19 2021
  read: IOPS=839, BW=840MiB/s (880MB/s)(984GiB/1200024msec)
```

The connection resumed reasonably quickly when running this test. However, it might take up to 10 minutes to recover, if so the runtime for fio may need to be increased from 20 minutes to 30 minutes.

## Conclusion

When both the proxy and the client are mounted using the `hard` option (recommended) any NFS operations between the proxy and the source will wait for the source to recover.

No manual intervention on the proxy or the clients is required.

The connection does not recover immediately. NFS has a linear back off on retries based on the `timeo` setting, with a maximum retry time of 600 seconds.

The source server *MUST* have the same IP when it recovers. Even if the source is mounted by DNS name, the NFS mount and RPC requests have already resolved the IP address of the source server.

If the source server changes IP address the proxy will need to be restarted. The clients will be unaffected as clients are connected to the proxy via the load balancer. When the proxy restarts, the clients will reconnect automatically and resume.
