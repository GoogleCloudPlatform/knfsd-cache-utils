# NetApp Exports Auto-Discovery Configuration

For instructions on how to get the NetApp root CA certificate, and verify the `netapp-exports` command works see the [netapp-exports README](../image/resources/netapp-exports/README.md).

**NOTE:** A service account must be assigned to allow the proxy access to the GCP secret. The service account will need to be granted the following permissions on the NetApp secret:

* Secret Manager Viewer (`roles/secretmanager.viewer`)
* Secret Manager Secret Accessor (`roles/secretmanager.secretAccessor`)

**IMPORTANT:** *Do not* assign the permissions at the project level as this will allow the NetApp proxy to read any secret in the project. Assign the IAM permissions directly on the NetApp secret.

#### NetApp Self-Signed Certificates

Modern SSL certificates use the Subject Alternate Name (SAN) field to provide a list of DNS names and IPs that are valid for the certificate.

However, older certificates relied on the Common Name (CN) field. This use has been deprecated and is no longer supported by default as the Common Name field was ambiguous.

If you have a certificate that does not contain a Subject Alternate Name then you can set `NETAPP_ALLOW_COMMON_NAME=true`. When this is enabled the Common Name *must* be the DNS name or IP address of the NetApp cluster. This DNS name or IP address *must* be used for the `NETAPP_URL` host.

If the certificate contains a Subject Alternate Name then the Common Name will be ignored.

#### Updating NetApp secret

Normally using the `latest` version for secrets in Terraform is discouraged because Terraform will not detect when a new version is added to the secret. However, in this case using `latest` does not cause any issues because the secret is only used when a proxy instance is starting up.

To update the NetApp secret, just add a new version and disable the old version. Once the new version has been verified as valid the old version can be destroyed.

Changing the password and updating the secret will not affect any running instances as the password is only required to generate the list of exports when the instance starts.
