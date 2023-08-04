# Network

## Cloud SQL (FSID database)

All connections to the Cloud SQL instance is authenticated using IAM permissions and encrypted using SSL.

The encryption and authentication is handled externally to the Cloud SQL database by the [Cloud SQL Auth Proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy).

The database connection requires two levels of authentication:

1. Cloud SQL Auth Proxy

   This uses GCP's IAM permissions to authenticate that the connection is allowed to access the Cloud SQL instance. For the knfsd proxy instances this normally uses the Service Account that is assigned to the Compute Instance.

2. PostgreSQL Authentication

   Once connected the PostgreSQL instance the connection needs to be authenticated to the PostgreSQL instance.

   The Cloud SQL instance is deployed with the `cloudsql.iam_authentication` flag enabled. This allows the knfsd proxy instances to authenticate using IAM using the Service Account that is assigned to the Compute Instance.

   **NOTE:** Other authentication methods supported by PostgreSQL such as username and password can be used, but IAM is recommended when using Cloud SQL.

The advantage of authenticating using IAM with the Compute Instance's Service Account is that no long-term credentials (e.g. private key, access token, password) are stored on the knfsd proxy instances. The GCP infrastructure handles the credentials and only provides individual short-term OAuth2 access tokens as required.

### Cloud SQL with private IP address

Using a private IP address to access the Cloud SQL instance is the recommended option. When deploying Cloud SQL with a private IP address you will need to [configure private service access](https://cloud.google.com/sql/docs/postgres/configure-private-ip).

As an example, the following Terraform configuration can be used to deploy private service access:

```terraform
resource "google_project_service" "servicenetworking" {
  project            = "my-gcp-project"
  service            = "servicenetworking.googleapis.com"
  disable_on_destroy = false
}

resource "google_compute_global_address" "private_ip_address" {
  project       = "my-gcp-project"
  name          = "private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 20
  network       = "my-vpc"
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network = "my-vpc"
  service = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [
    google_compute_global_address.private_ip_address.name
  ]
  depends_on = [
    google_project_service.servicenetworking,
  ]
}
```

### Cloud SQL with public IP address

The GCP infrastructure prevents arbitrary connections to the Cloud SQL instance, even when it has a public IP address. The FSID database is deployed without any authorized networks; this prevents any direct access to the Cloud SQL instance.

Connections to the Cloud SQL proxy are authenticated using IAM and encrypted over an SSL connection. See [Cloud SQL Auth Proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy) for more details.

Because the Cloud SQL instance is using a public IP address the knfsd proxy instances will require access to the internet. Either using a public IP (not recommended), or Cloud NAT.
