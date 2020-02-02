# XMC NBI DeviceLister (Go)

DeviceLister uses the GraphQL-based API provided by the Northbound Interface (NBI) of [Extreme Management Center](https://www.extremenetworks.com/product/extreme-management-center/) (XMC; formerly known as NetSight) to fetch and display a list of all devices. The list includes the following pieces of information for each device:

  * up/down (as +/- and text)
  * IP address
  * vendor and model
  * system name or nick name

The tool is intended to provide a quick overview of the managed devices and to serve as a starting point for other tools.

## Branches

This project uses two defined branches:

  * `master` is the primary development branch. Code within `master` may be broken at any time.
  * `stable` is reserved for code that compiles without errors and is tested. Track `stable` if you just want to use the software.

Other branches, for example for developing specific features, may be created and deleted at any time.

## Dependencies

DeviceLister uses the [module xmcnbiclient](https://gitlab.com/rbrt-weiler/go-module-xmcnbiclient). This module has to be installed with `go get gitlab.com/rbrt-weiler/go-module-xmcnbiclient` or updated with `go get -u gitlab.com/rbrt-weiler/go-module-xmcnbiclient` before running or compiling DeviceLister. All other dependencies are included in a standard Go installation.

## Running / Compiling

Use `go run DeviceLister.go` to run the tool directly or `go build DeviceLister.go` to compile a binary.

Tested with [go1.13](https://golang.org/doc/go1.13).

## Usage

`DeviceLister -h`:

<pre>
Available options:
  -clientid string
    	Client ID for OAuth
  -clientsecret string
    	Client Secret for OAuth
  -host string
    	XMC Hostname / IP (default "localhost")
  -insecurehttps
    	Do not validate HTTPS certificates
  -nohttps
    	Use HTTP instead of HTTPS
  -password string
    	Password for HTTP Basic Auth
  -timeout uint
    	Timeout for HTTP(S) connections (default 5)
  -username string
    	Username for HTTP Basic Auth (default "admin")
  -version
    	Print version information and exit

OAuth will be preferred over username/password.
</pre>

## Source

The original project is [hosted at GitLab](https://gitlab.com/rbrt-weiler/xmc-nbi-devicelister-go), with a [copy over at GitHub](https://github.com/rbrt-weiler/xmc-nbi-devicelister-go) for the folks over there. Additionally, there is a project at GitLab which [collects all available clients](https://gitlab.com/rbrt-weiler/xmc-nbi-clients).
