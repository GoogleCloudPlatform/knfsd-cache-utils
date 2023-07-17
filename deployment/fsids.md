# Filesystem Identifiers (FSIDs)

Almost all NFS operations use an opaque file handle to identify the file or directory that is the target of the operations (such as for reads and writes). The exact format depends upon the NFS server, for Linux's kernel NFS server (knfsd) the handle consists of two main parts, a filesystem identifier (FSID) and an inode number.

Every exported filesystem requires a unique filesystem identifier (FSID). In a standard NFS server the NFS service can automatically derive the FSID number for an export from the underlying filesystem's UUID or the hardware's device ID.

This option is not available for a knfsd proxy when re-exporting. The knfsd proxy cannot re-use the source server's FSID value as this could conflict with local filesystems on the proxy (or other source services if the proxy has multiple sources).

As such each export on the proxy has to be explicitly allocated a unique FSID number. This number should ideally be the same for all proxy instances in the cluster to avoid I/O errors or data corruption if a client switches proxy instance (e.g. when using the load balancer).

## Importance of unique FSID numbers

When a client needs to access a file or directory the client uses the LOOKUP operation to convert a file's name to a file handle. After that the client will continue to use the file handle it received for any READ or WRITE operations until the client closes the file.

The NFS protocol (and NFS client) assumes that this handle will remain stable in the face of communication issues. If TCP connection is interrupted, the client will reconnect and carry on using the same file handle.

This can lead to two possible issues if we're not careful:

1. The FSID and inode is valid, but maps to the wrong file resulting in data corruption.
2. The FSID or inode becomes invalid resulting in an I/O error

As an example of how this can occur assume we have two filesystems on a proxy:

* `/files` with FSID 1
* `/archive` with FSID 2

The client has connected and is reading data on `/files/`. The NFS server crashes and reboots, but on start up the allocates different FSIDs to the exports.

* `/files` with FSID 2
* `/archive` with FSID 1

The client retries the TCP connection until NFS server is available, then carries on its previous READ operation. However, the file handle for the READ operation has FSID 1, which was previously `/files` but is now `/archive`.

The NFS server will look for a file on `/archive` with the inode from the file handle. At this point one of two things may happen:

1. `/archive` contains a file with the inode number. The NFS server replies with data from the wrong file in the middle of the data that was previously read.

2. `/archive` does not contain that inode number resulting in an I/O error.

With a knfsd proxy cluster there are two main ways this issue with FSIDs can occur:

1. When a knfsd proxy instance is replaced (e.g. due to a failed health check) the new instance does not assign the same FSIDs to each export. This can be due to:

   * The MIG template was updated with a change to the exports in `EXPORT_MAP`.

   * Using auto-discovery, and the source server's exports have changed (e.g. due to volumes being added or removed).

   * Using auto re-export, as order FSIDs are assigned are based on the order clients access the nested volumes.

2. A client re-connects to a different knfsd proxy instance (e.g. due to load balancing) and knfsd proxy cluster has inconsistent FSIDs.

These issues can be avoided by storing the mappings between FSID and export path in an external database by using `FSID_MODE="external"`.

## FISD Mode

The `FSID_MODE` variable controls how FSID numbers are allocated to exports. There are two main options; `static` or using an FSID service. For the FSID service there are two options available. The standard NFS `fsidd` service, or the `knfsd-fsidd` service.

* `static` - FSID numbers are explicitly allocated to exports on start-up.
* `local` - FSID numbers are automatically allocated to exports by the standard NFS `fsidd` service.
* `external` - FSID numbers are automatically allocated to exports by the `knfsd-fsidd` service.

The main difference between the standard NFS `fsidd` service and the `knfsd-fsidd` service is that the standard `fsidd` service uses a local sqlite database, while the `knfsd-fsidd` service uses a Cloud SQL PostgreSQL instance.

### Static

FSID numbers are explicitly allocated to exports on start-up by the `proxy-startup.sh` script.

This mode is only recommended when using `EXPORT_MAP` to explicitly define which exports are re-exported (and in which order) to ensure all the knfsd proxies in the cluster assign the same FSID number to each export.

If using auto-discovery there is a risk that if a proxy instance is rebooted or replaced the exports might have changed (new exports added or exports removed). If this happens the same export will have different FSID numbers on different proxy instances.

### Local

Each export is automatically allocated an FSID number by the `mountd` service using using the standard NFS `fsidd` service. The mappings between FSID and export path are stored in a local SQLite database.

Local is not recommended for production and is only intended for testing with single instance proxy clusters. Because the SQLite database is stored locally on the knfsd proxy instance, if you're using multiple proxy instances each instance could allocate a different FSID to the same export (or the same FSID to different exports).

Even with a single instance, there's still a risk if the instance is replaced due to a failing health check. When the instance is replaced the new instance will start with an empty SQLite database. When existing clients re-connect, they could see inconsistent FSIDs.

### External

Each export is automatically allocated an FSID number by the `mountd` service using using the `knfsd-fsidd` service. This uses a Cloud SQL PostgreSQL instance to store the mappings between FSID and export path. This ensures that all the instances in cluster use the same FSID for each export path.

This is the recommended deployment option for all knfsd proxy configurations as its the easiest to configure and ensures the consistency of FSIDs across the cluster.

## Providing a custom database

By default, when `FSID_MODE="external"` the deployment Terraform configuration will create a Cloud SQL PostgreSQL instance for the proxy cluster. This is the simplest, and recommended option.

However, if you want to create your database, such as to use a single Cloud SQL database for multiple clusters you can set `FSID_DATABASE_DEPLOY=false`.

Before deploying the knfsd proxy cluster create a suitable Cloud SQL PostgreSQL database (the `knfsd-fsidd` service only supports PostgreSQL).

### Using the Database Terraform module

The Database Terraform module in [deployment/database](./database/) can be used to create a Cloud SQL PostgreSQL instance suitable for use by a knfsd proxy cluster. This is the same module that the KNFSD Terraform module uses internally.

```terraform
# Need to create a Service Account for use by the knfsd proxy cluster
resource "google_service_account" "nfs_proxy" {
  project     = "my-gcp-project"
  account_id  = "nfs-proxy"
  description = "KNFSD proxy service account"
}

# Assign project IAM roles to the knfsd proxy Service Account
# to support logging and metrics.
resource "google_project_iam_member" "nfs_proxy" {
  for_each = toset([
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
  ])
  project = "my-gcp-project"
  role    = each.key
  member  = "serviceAccount:${google_service_account.proxy.email}"
}

# Create a Cloud SQL PostgreSQL database instance for use by KNFSD proxy
# clusters
module "fsid_database" {

  source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v1.0.0-beta6"

  project               = "my-gcp-project"
  region                = "us-west1"
  zone                  = "us-west1-a"

  # The final name will include a random suffix such as
  # rendercluster1-fsids-4ef0fafb. This is to avoid issues if you need to
  # destroy and re-create the database. Cloud SQL reserves instance names for
  # up two weeks.
  name_prefix           = "rendercluster1-fsids"

  proxy_service_account = google_service_account.nfs_proxy.email

  # Avoid accidentally deleting the database. Removing this database requires
  # three steps:
  # * Change this to false
  # * terraform apply
  # * terraform destroy
  # For testing, set this to false to make it easier to re-create the database.
  deletion_protection   = true
}

output "sql_user" {
  value = module.fsid_database.sql_user
}

output "database_name" {
  value = module.fsid_database.database_name
}

output "connection_name" {
  value = module.fsid_database.connection_name
}

output "private_ip_address" {
  value = module.fsid_database.private_ip_address
}
```

### Cloud SQL PostgreSQL configuration

The FSID service is not resource intensive, and does not require much storage. As such the minimum database instance type of `db-custom-1-3840` with a 10 GB SSD should be sufficient for most configurations. Smaller (shared core) instances are not recommended in production.

To allow authenticating with a Service Account using IAM, the database instance needs to be deployed with the `cloudsql.iam_authentication` flag set to `on`.

Once the database instance has been created, [assign the knfsd proxy Service Account as a user](https://cloud.google.com/sql/docs/postgres/add-manage-iam-users).

### IAM roles

The standard FSID configuration uses the Compute Instance's Service Account to authenticate with the Cloud SQL database using IAM.

For best practice you should use a dedicated Service Account for your knfsd proxy clusters (not the default Compute Engine Service Account).

The knfsd proxy Service Account requires roles:

* GCP Project roles (grant the Service Account these roles on the GCP project)

  **NOTE:** You need to assign these roles yourself, even if you're using the Database Terraform module.

  * Logs Writer (`roles/logging.logWriter`) - Allows writing to Cloud Logging.

  * Monitoring Metric Writer (`roles/monitoring.metricWriter`) - Allows writing metrics to Cloud Monitoring.

* Cloud SQL instance roles (grant the Service Account these roles on the Cloud SQL FSID database instance)

  If you're using the Database Terraform module these roles will be assigned to the service account by the Database Terraform module.

  * Cloud SQL Client (`roles/cloudsql.client`) - Allows connecting to the Cloud SQL instance.

  * Cloud SQL Instance User (`roles/cloudsql.instanceUser`) - Allows logging in to the Cloud SQL instance using IAM.

### FSID Database Configuration

The when using `FSID_MODE="external"` the `knfsd-fsidd` service can be configured by setting `FSID_DATABASE_CONFIG`. This is required

* `socket` (Optional) - The unix socket to listen on for incoming FSID requests from `mountd`. This *must* match the value configured in `/etc/nfs.conf`. Default `/run/fsidd.sock`.

* `debug` (Optional) - Enabled writing verbose debug output to `stderr`. Default `false`.

* `cache` (Optional) - Enables caching FSID mappings to avoid querying FSID database. Setting this to false can result in excessive SQL queries and slow performance and is only intended for debugging. Default `true`.

---

The `[database]` section supports:

* url (Required) - A [`pgxpool` URL](https://pkg.go.dev/github.com/jackc/pgx/v4@v4.17.2/pgxpool#ParseConfig). Normally only the `user` and `database` options need to be set. The host and authentication will be handled by the `cloudsqlconn` library.

* instance (Required) - The Cloud SQL instance to connect to. The instance argument must be the instance's connection name, which is in the format "project-name:region:instance-name".

* iam-auth (Optional) - Set to `true` to enable automatic IAM authentication. Using automatic IAM authentication is recommended. This will use the machine's service account to authenticate with Cloud SQL. Default `false`.

* private-ip (Optional) - Set to `true` to use a private IP (VPC). To access the Cloud SQL instance via its private IP you will need to [configure private service access](https://cloud.google.com/sql/docs/postgres/private-ip). Default `false`.

* table-name (Required) - The name of table to store the FSID mappings for the proxy cluster. It is recommended that each proxy cluster has its own table.

* create-table (Optional) - When `true` the `knfsd-fsidd` service will try to create its own table on start up. If set to `false` the table must already exist. See [knfsd-fsidd/schema.sql](../image/resources/knfsd-fsidd/schema.sql). Default `false`.

---

The `[metrics]` section supports:

* enabled (Optional) - Set to `true` to report metrics such as the number of requests, SQL operations, etc. Default `false`.

* endpoint (Optional) - The endpoint to report metrics using the OTLP format. The endpoint must support GRPC. Default `localhost:4317`.

* insecure (Optional) - Set to `true` to allow sending metrics via GRPC without any encryption or endpoint verification. Default `false`.

* interval (Optional) - How frequently to send metrics. Default `60s`.

### Example FSID database configuration

```ini
socket=/run/fsidd.sock

[database]
iam-auth=true
instance=my-gcp-project:us-west1:rendercluster1-fsids-4ef0fafb
url=user=fsids database=fsids
private-ip=false
table-name=fsids
create-table=true

[metrics]
enabled=true
# The custom knfsd-metrics-agent has been configured to listen on this socket
# and forward metrics to GCP Cloud Monitoring.
endpoint=unix:///run/knfsd-metrics.sock
# TLS security is not required as unix domain sockets can only be accessed on
# the local machine. This socket will be protected by standard unix permissions.
insecure=true
interval=1m
```
