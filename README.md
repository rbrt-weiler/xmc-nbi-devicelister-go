# XMC NBI DeviceLister (Go)

DeviceLister uses the GraphQL-based API provided by the Northbound Interface (NBI) of [Extreme Management Center (XMC)](https://www.extremenetworks.com/product/extreme-management-center/) to fetch and display a list of all devices. The list includes the following pieces of information for each device:

  * up/down (as +/- and text)
  * IP address
  * vendor and model
  * system name or nick name

The tool is intended to provide a quick overview of the managed devices and to serve as a starting point for other tools.

## Running / Compiling

Use `go run DeviceLister.go` to run the tool directly or `go build DeviceLister.go` to compile a binary.

Tested with go1.11 and go1.13.

## Usage

`DeviceLister -h`:

<pre>
  -host string
    	XMC Hostname / IP (default "localhost")
  -httptimeout uint
    	Timeout for HTTP(S) connections (default 5)
  -insecurehttps
    	Do not validate HTTPS certificates
  -password string
    	Password for HTTP auth
  -username string
    	Username for HTTP auth (default "admin")
</pre>

## Source

The original project is [hosted at GitLab](https://gitlab.com/rbrt-weiler/xmc-nbi-devicelister-go), with a [copy over at GitHub](https://github.com/rbrt-weiler/xmc-nbi-devicelister-go) for the folks over there. Additionally, there is a project at GitLab which [collects all available clients](https://gitlab.com/rbrt-weiler/xmc-nbi-clients).
