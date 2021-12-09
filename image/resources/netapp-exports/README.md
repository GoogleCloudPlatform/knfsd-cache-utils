# Netapp Mount tool prototype

Uses the NetApp REST API to list volumes on a NetApp filer. This is similar to the `showmount` command and is useful for auto-discovery when `showmount` is disabled.

## Options

* `-config path`\
  Path to a config file. This can be used as an alternative to specify the host, password, etc as command line options. Config files also support listing multiple servers.

* `-host string`\
  DNS or IP of the NetApp server. This is the DNS or IP name clients use when mounting the NFS shares.

* `-url string`\
  URL of the NetApp REST API. This *must* include the API version and end with a slash, for example `https://netapp.example/api/v1/`.

* `-user string`\
  The username used to authenticate with the NetApp REST API.

* `-password string`\
  The password used to authenticate with the NetApp REST API. This option is not secure and is only intended for testing. For more secure options, use `-secret-name` or `-password-file`.

* `-password-file path`\
  A file a password to authenticate with the NetApp REST API.

* `-secret-project string`\
  The GCP project containing the secret. Defaults to the current project when running on a GCP Compute Instance.

* `-secret-name string`\
  The name of a GCP Secret containing the NetApp REST API password.

* `-secret-version string`\
  The version of the secret. Defaults to `latest`.

* `-ca path`\
  Path to PEM encoded certificate file containing the root certificate for the NetApp REST API. This can also include intermediate certificates to provide the full certificate chain.

* `-insecure`\
  Allow insecure connections. This permits the use of the unencrypted `http` connections and ignores any certificate errors. This is only intended for testing as it can expose the password over an unencrypted connection, and encrypted connections will be vulnerable to man in the middle attacks.

* `-allow-common-name`\
  Allows using the Common Name (CN) field of the certificate as a DNS name when the certificate does not include a Subject Alternate Name (SAN) field. Use of the CN field is now deprecated as CN is ambiguous and only intended to provide a human readable name. However, some self-signed NetApp certificates still rely on the CN field. If the certificate contains a SAN then the CN will be ignored.

## Environment Variables

Alternatively, most of the options can be set using environment variables.

| Option               | Environment Variable       |
|----------------------|----------------------------|
| `-host`              | `NETAPP_HOST`              |
| `-url`               | `NETAPP_URL`               |
| `-user`              | `NETAPP_USER`              |
| `-password`          | `NETAPP_PASSWORD`          |
| `-password-file`     | `NETAPP_PASSWORD_FILE`     |
| `-secret-project`    | `NETAPP_SECRET_PROJECT`    |
| `-secret-name`       | `NETAPP_SECRET`            |
| `-secret-version`    | `NETAPP_SECRET_VERSION`    |
| `-ca`                | `NETAPP_CA`                |
| `-allow-common-name` | `NETAPP_ALLOW_COMMON_NAME` |

## Prerequisites

### NetApp CA certificate

First you need the root CA certificate for NetApp.

If NetApp is using a self-signed certificate you can fetch this using:

```bash
openssl s_client -connect ${NETAPP?}:443 </dev/null | openssl x509 > netapp.pem
```

Verify that the certificate was downloaded:

```bash
openssl x509 -in netapp.pem -noout -issuer -subject
```

### Check NetApp's SSL certificate names

```bash
openssl s_client -connect ${NETAPP?}:443 </dev/null 2>/dev/null | openssl x509 -noout -subject -nameopt sname -ext subjectAltName
```

If the output only contains a common name (`/CN=netapp.example`) you will need
to use this as the DNS name or IP and include the `-allow-common-name` option.

If the output includes `X509v3 Subject Alternative Name` you can use any of the
`DNS:` or `IP:` entries to access the NetApp REST API.

## Running the tool

Create a file with your NetApp user password, e.g. `netapp-password`.

Set the environment variables for the tool:

```bash
export NETAPP_HOST=netapp.example
export NETAPP_URL=https://netapp.example/api/v1/
export NETAPP_USER=admin
export NETAPP_PASSWORD_FILE=netapp-password
export NETAPP_CA=netapp.pem
```

Run the command:

```bash
./netapp-exports
```

If the SSL certificate only contains a common name:

```bash
./netapp-exports -allow-common-name
```

## Troubleshooting

### The NetApp certificate is not valid

If the NetApp certificate is not valid because:

* The DNS name of the certificate does not match the DNS name used to access NetApp.
* The certificate has expired

You can ignore any SSL certificate errors using the `-insecure` option.

**WARNING:** Using this option is vulnerable to man in the middle attacks and should not be used in production.

```bash
./netapp-exports -insecure
```
