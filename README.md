# XMC API Clients - Go - DeviceLister

A simple tool that fetches all known devices from XMC.

## Compiling

`go build DeviceLister.go`

Tested with go1.13.

## Usage

`DeviceLister -h`:

<pre>
  -host string
        XMC Hostname / IP (default "localhost")
  -httptimeout int
        Timeout for HTTP(S) connections (default 5)
  -password string
        Password for HTTP auth
  -username string
        Username for HTTP auth (default "admin")
</pre>
