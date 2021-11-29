# Difference in caching of directory listings between proxy and source

A difference was observed between clients connected directly to the source vs clients connected via the proxy when performing the `ls` command.

A client connected directly to the source would invalidate its cache immediately upon a directory's contents changing while clients connected through the proxy would continue to see the original directory listing.

## Setup

Create a single node NFS proxy and two clients. The clients need to be able to connect to the NFS proxy and directly to the source server.

Client 1 will have local caching of directory metadata enabled. To give time for the test, and to fully highlight the difference an aggressive cache time of 10 minutes will be used; `ac,actimeo=600`. The proxy also has `actimeo=600`.

* Client 1 mounts a volume via the proxy
* Client 1 mounts the same volume directly from the source
* Client 2 mounts the same volume directly from the source

This allows Client 1 to view the same directory via the proxy and directly from the source. Client 2 will be used to modify the directory by creating new files and directories.

## Procedure

* Client 1: `ls` a directory through the proxy
* Client 1: `ls` the same directory directly from the source
* Client 2: Use `touch` and `mkdir` to create new files and directories directly on the source.
* Client 1: `ls` the directory through the proxy
* Client 1: `ls` the directory directly from the source

## Observed Behaviour

Client 1 immediately shows the new files and directories when listing the source, but when listing via the proxy Client 1 will continue to show the original cached directory listing until the cache expires (after 10 minutes).

This difference in behaviour seems odd, because the client should be caching the directory metadata for 10 minutes, the same as the proxy.

## Further Analysis

To further understand the difference in behaviour `nfstrace` was used on both Client 1 and the proxy.

### `ls` Direct to Source (1st)

The initial request from the client to the source has to look up a handle for the test directory and then list the directory's contents.

```text
client -> source LOOKUP      [ dir   : /files       name: test ]
client -> source GETATTR     [ object: /files/test             ]
client -> source READDIRPLUS [ dir   : /files/test             ]
```

Running a second `ls` command without modifying the directory shows that the client always issues a `GETATTR` request, even when the metadata for the directory is cached. This is to check the directory's `mtime` to see if the client's cache is still valid.

```text
client -> source GETATTR     [ object: /files/test             ]
```

### `ls` via Proxy (1st)

The initial request from the client via the proxy is similar to accessing direct to source, only the proxy has to forward the request to the source.

```text
client -> proxy  LOOKUP      [ dir   : /files       name: test ]
proxy  -> source LOOKUP      [ dir   : /files       name: test ]
client -> proxy  GETATTR     [ object: /files                  ]
client -> proxy  GETATTR     [ object: /files/test             ]
client -> proxy  READDIRPLUS [ dir: /files/test                ]
proxy  -> source READDIRPLUS [ dir: /files/test                ]
```

Running a second `ls` command without modifying the directory shows the same behaviour as before. The client issues a `GETATTR` to the proxy to check the directory's `mtime`.

Note however that the proxy *does not* forward the `GETATTR` request to the source. The proxy already has the metadata cached and cannot tell the difference between this special behaviour to validate the client's cache, and a standard metadata lookup on a directory (e.g. `stat`).

```text
client -> proxy  GETATTR     [ object: /files/test             ]
```

### `ls` Direct to Source (2nd)

After modifying the test directory, running an `ls` command from the client directly to the source shows the client sends a `GETATTR`. This allows the client to detect that the test directory's `mtime` has changed, invalidating the client's cache.

```text
client -> source GETATTR     [ object: /files/test             ]
client -> source READDIRPLUS [ dir   : /files/test             ]
```

### `ls` via Proxy (2nd)

After modifying the test directory, running an `ls` command from the client via the proxy shows the same initial behaviour on the client. The client sends the `GETATTR` command.

However, this time the proxy answers with the cached metadata. As such the client does not detect a change to the test directory's `mtime` and thinks the client's local cache is still valid.

This causes the client to skip the `READDIRPLUS` command to get the latest directory listing. Even if the client's cache has expired, if the client were to issue a `READDIRPLUS` to the proxy, the proxy may answer with stale cached data until the proxy's cached metadata for the test directory expires.

```text
client -> proxy  GETATTR     [ object: /files/test             ]
```

## Conclusion

The `ls` command bypasses the cache for the directory's attributes and always issues a `GETATTR` (assuming this is to check the directory's `mtime`) even if the client still has the directory info cached.

The `ls` command doesn't actually have special logic for NFS. The handling for this is in the NFS client when handling the [`readdir` function call](https://github.com/torvalds/linux/blob/89d714ab6043bca7356b5c823f5335f5dce1f930/fs/nfs/dir.c#L1094-L1098).

This differs from other commands such as `stat`. Running a stat a directory will only issue a `GETATTR` if the cached info has expired.

When querying directly to the source server, the source will reply to the `GETATTR` with the latest metadata.

When querying via the a proxy, the proxy only sees a `GETATTR` request. The proxy does not know the client's context. So the proxy does not know if the `GETATTR` is for an `ls` command or a `stat` command. As such the proxy always responds to a `GETATTR` request from the cache.

Lowering `acdirmin` and `acdirmax` will cause the directory info to expire sooner. This will fix the listing issue by forcing the proxy to go back to the
source when responding to the client's `GETATTR`.

However, this will also result in extra metadata traffic back to the source. This is because other requests such as `stat` on the directory that previously would have been answered by the cache will also have to issue a `GETATTR` to the source.

This does not result in extra `READDIR` operations. If `GETATTR` indicates that
proxy's cache is still valid (based on the directory's `mtime`), the proxy can respond to the `READDIR` from the proxy's cache.
