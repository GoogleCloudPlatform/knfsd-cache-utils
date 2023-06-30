# Knfsd FSID Database

This module deploys a PostgreSQL database for use with the [external fsidd service](../auto-reexport.md).

NOTE: This module is deployed automatically by `terraform-module-knfsd` when `REEXPORT` is enabled (unless you disable `REEXPORT_DATABASE_DEPLOY`).

## Inputs

* `project` - (Optional) The ID of the project in which the resource belongs. If it is not provided, the provider project is used.

* `region` - (Optional) The region the instance will sit in. If a region is not provided, the provider region will be used instead.

* `zone` - (Optional) The preferred compute engine zone that the database instance will be deployed to. If a zone is not provided, a random zone will be chosen.

* `availability_type` - (Optional) This can be either high availability (`REGIONAL`) or single zone (`ZONAL`). Defaults to `ZONAL`.

* `name_prefix` - (Optional) Prefix to use when generating a Cloud SQL instance name. The name will be suffixed with a hyphen and 8 random letters/digits (similar to MIG compute instances). Defaults to `fsids`.

* `name` - (Optional) The name of the Cloud SQL instance. If the name is left blank a random name will be generated based on `name_prefix`. This is done because after a name is used, it cannot be reused for up to [one week](https://cloud.google.com/sql/docs/delete-instance).

* `tier` - (Optional) The machine type to use. Must be a supported [PostgreSQL machine type](https://cloud.google.com/sql/docs/postgres/instance-settings#machine-type-2ndgen). Defaults to `db-custom-1-3840`.

* `deletion_protection` - (Optional) Whether or not to allow Terraform to destroy the instance. Unless this field is set to false in Terraform state, a `terraform destroy` or `terraform apply` command that deletes the instance will fail. Defaults to true.

* `proxy_service_account` - (Required) The Service Account used by the knfsd proxy Compute Instances. This Service Account will be granted access to the PostgreSQL database.

* `enable_public_ip` - (Optional) Whether to deploy the database with a public IP address. At least one of `enable_public_ip` or `private_network` must be configured. Defaults to false.

* `private_network` - (Optional) The VPC network from which the Cloud SQL instance is accessible for private IP. At least one of `enable_public_ip` or `private_network` must be configured.

* `allocated_ip_range` - (Optional) The name of the allocated IP range for the private IP. If set the instance IP will be created in the allocated range.

## Outputs

* `sql_user` - The name of the database user account for the `proxy_service_account`.

* `database_name` - The name of the SQL database.

* `name` - The name of the Cloud SQL instance.

* `connection_name` - The connection name of the instance to be used in connection strings.

* `public_ip_address` - The first public (`PRIMARY`) IPv4 address assigned.

* `private_ip_address` - The first private (`PRIVATE`) IPv4 address assigned.

* `ip_address.0.ip_address` - The IPv4 address assigned.

* `ip_address.0.time_to_retire` - The time this IP address will be retired, in RFC 3339 format.

* `ip_address.0.type` - The type of this IP address.
  * A `PRIMARY` address is an address that can accept incoming connections.
  * An `OUTGOING` address is the source address of connections originating from the instance, if supported.
  * A `PRIVATE` address is an address for an instance which has been configured to use private networking see: Private IP.

* `server_ca_cert.0.cert` - The CA Certificate used to connect to the SQL Instance via SSL.

* `server_ca_cert.0.common_name` - The CN valid for the CA Cert.

* `server_ca_cert.0.create_time` - Creation time of the CA Cert.

* `server_ca_cert.0.expiration_time` - Expiration time of the CA Cert.

* `server_ca_cert.0.sha1_fingerprint` - SHA Fingerprint of the CA Cert.

## Limitations

### Error deleting database; database in use

> Error: Error when reading or editing Database: googleapi: Error 400: Invalid  request: failed to delete database "fsids". Detail: pq: database "fsids" is being accessed by other users. (Please use psql client to delete database that is not owned by "cloudsqlsuperuser")., invalid

The database still has active connections so cannot be deleted. This may occur if the MIG has not yet finished shutting down all the VMs before Terraform attempts to delete the database. Try again once the MIG has finished being removed by GCP.

If the database is used by more than one cluster, ensure all the clusters have been shutdown.

### Error deleting database user; objects depend on user

> Error: Error, failed to deleteuser nfs-proxy@ab-knfsd.iam in instance fsids-1cef303d: googleapi: Error 400: Invalid request: failed to delete user nfs-proxy@ab-knfsd.iam: . role "nfs-proxy@ab-knfsd.iam" cannot be dropped because some objects depend on it Details: 1 object in database fsids., invalid

This is because there are database tables in the fsids database that are owned by the `proxy_service_account`. This may occur because Terraform has not finished deleting the fsids database before it attempts to remove the database user. Try again once the fsids database has finished being deleted.

### Cannot change proxy_service_account; objects depend on user

> Error: Error, failed to deleteuser nfs-proxy@ab-knfsd.iam in instance fsids-1cef303d: googleapi: Error 400: Invalid request: failed to delete user nfs-proxy@ab-knfsd.iam: . role "nfs-proxy@ab-knfsd.iam" cannot be dropped because some objects depend on it Details: 1 object in database fsids., invalid

This is because there are database tables in the fisds database that are owned by the old `proxy_service_account`. Before the old `proxy_service_account` can be removed the ownership of these tables must be changed to a different user.

Once the new `proxy_service_account` database user has been created change the ownership of the tables to the new `proxy_service_account` database user.
