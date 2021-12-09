# Example of configuring and testing netapp-exports

This is a short example for how to extract the CA certificate from a NetApp instance, configure the netapp-exports tool and run it locally. This can be used to test that the tool can query a NetApp instance and produces the correct result.

**NOTE:** For simplicity while testing the NetApp password is stored in a file. For production use it is advised to use Cloud Secrets to store the password.

## Export the NetApp certificate

Set the NETAPP environment variable to a DNS name or IP that can be used to
contact the NetApp server.

```
chris@netapp-client:~$ export NETAPP=netapptest

chris@netapp-client:~$ openssl s_client -connect ${NETAPP?}:443 </dev/null | openssl x509 > netapp.pem
Can't use SSL_get_servername
depth=0 CN = netapptest, C = US
verify error:num=18:self signed certificate
verify return:1
depth=0 CN = netapptest, C = US
verify return:1
DONE
```

## Verify the certificate details

In this example the certificate is a self-signed certificate. The certificate only contains a common name based on the host name (`netapptest`) instead of the fully qualified domain name.

```
chris@netapp-client:~$ openssl x509 -in netapp.pem -noout -issuer -subject
issuer=CN = netapptest, C = US
subject=CN = netapptest, C = US

chris@netapp-client:~$ openssl s_client -connect ${NETAPP?}:443 </dev/null 2>/dev/null | openssl x509 -noout -subject -nameopt sname -ext subjectAltName
subject=/CN=netapptest/C=US
```

## Verify DNS name resolves

If the certificate contains subject alternative names (SANs) then you can use
any of `DNS:` or `IP:` entries.

In this example the self-signed certificate only contains a common name (CN) of
`CN=netapptest`, so the DNS name we need to use when accessing the API is
`netapptest`.

```
chris@netapp-client:~$ nslookup netapptest
Server:		169.254.169.254
Address:	169.254.169.254#53

Non-authoritative answer:
Name:	netapptest.europe-west4-c.c.appsbroker-shared-1.internal
Address: 10.164.0.116
```

## Configure environment

Create a file named `netapp-password` with the password for your NetApp user.

You will need to change `NETAPP_URL` and `NETAPP_USER` to match your NetApp
settings.

```
chris@netapp-client:~$ vim netapp-password
chris@netapp-client:~$ export NETAPP_URL=https://netapptest/
chris@netapp-client:~$ export NETAPP_USER=admin
chris@netapp-client:~$ export NETAPP_PASSWORD_FILE=netapp-password
chris@netapp-client:~$ export NETAPP_CA=netapp.pem
```

Check the required files are present:

```
chris@netapp-client:~$ ls
netapp-password  netapp.pem  netapp-exports
```

## Run NetApp mount tool

Run the tool, and direct the output to a file.

In this example because NetApp is using a self-signed certificate that only
contains a common name (CN) the command needs to include `-allow-common-name`.

```
chris@netapp-client:~$ ./netapp-exports -allow-common-name > volumes-netapp.txt
```

Check the output of the file, it should list the NFS paths of all the volumes exported from the NetApp cluster.

```
chris@netapp-client:~$ cat volumes-netapp.txt
/
/archive
/home
/hotfiles
/pipeline
```

## Run showmount

`showmount` as an alternative way to list the NFS exports, but it is not always
supported.

```
/sbin/showmount -e nfstest 2>&1 > volumes-showmount.txt
```

Check the output of the file:

```
chris@netapp-client:~$ cat volumes-showmount.txt
```

If showmount is not supported by the NFS server you will see the error:

```
clnt_create: RPC: Unknown host
```

Otherwise you will see a list of exports:

```
Export list for netapptest:
/ 10.164.0.0/20
/archive 10.164.0.0/20
/home 10.164.0.0/20
/hotfiles 10.164.0.0/20
/pipeline 10.164.0.0/20
```

## Check all the volumes are listed

Check that the `volumes-netapp.txt` file listed all the volumes, including any
nested volumes (junction points).

If the showmount command worked, then also check the `volumes-showmount.txt`
file, compare this with the `volumes-netapp.txt` file.
