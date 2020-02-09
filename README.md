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

Use `go run DeviceLister.go` to run the tool directly or `go build DeviceLister.go` to compile a binary. Prebuilt binaries may be available as artifacts from the GitLab CI/CD [pipeline for tagged releases](https://gitlab.com/rbrt-weiler/xmc-nbi-devicelister-go/pipelines?scope=tags).

Tested with [go1.13](https://golang.org/doc/go1.13).

## Usage

`DeviceLister -h`:

<pre>
Available options:
  -basicauth
        Use HTTP Basic Auth instead of OAuth
  -host string
        XMC Hostname / IP
  -insecurehttps
        Do not validate HTTPS certificates
  -nohttps
        Use HTTP instead of HTTPS
  -path string
        Path where XMC is reachable
  -port uint
        HTTP port where XMC is listening (default 8443)
  -secret string
        Client Secret (OAuth) or password (Basic Auth) for authentication
  -timeout uint
        Timeout for HTTP(S) connections (default 5)
  -userid string
        Client ID (OAuth) or username (Basic Auth) for authentication
  -version
        Print version information and exit
</pre>

## Authentication

DeviceLister supports two methods of authentication: OAuth2 and HTTP Basic Auth.

* OAuth2: To use OAuth2, provide the parameters `userid` and `secret`. DeviceLister will attempt to obtain a OAuth2 token from XMC with the supplied credentials and, if successful, submit only that token with each API request as part of the HTTP header.
* HTTP Basic Auth: To use HTTP Basic Auth, provide the parameters `userid` and `secret` as well as `basicauth`. DeviceLister will transmit the supplied credentials with each API request as part of the HTTP request header.

As all interactions between DeviceLister and XMC are secured with HTTPS by default both methods should be safe for transmission over networks. It is strongly recommended to use OAuth2 though. Should the credentials ever be compromised, for example when using them on the CLI on a shared workstation, remediation will be much easier with OAuth2. When using unencrypted HTTP transfer (`nohttps`), Basic Auth should never be used.

In order to use OAuth2 you will need to create a Client API Access client. To create such a client, visit the _Administration_ -> _Client API Access_ tab within XMC and click on _Add_. Make sure to note the returned credentials, as they will never be shown again.

## Authorization

Any user or API client who wants to access the Northbound Interface needs the appropriate access rights. In general, checking the full _Northbound API_ section within rights management will suffice. Depending on the use case, it may be feasible to go into detail and restrict the rights to the bare minimum required.

For API clients (OAuth2) the rights are defined when creating an API client and can later be adjusted in the same tab. For regular users (HTTP Basic Auth) the rights are managed via _Authorization Groups_ found in the _Administration_ -> _Users_ tab within XMC.

## Source

The original project is [hosted at GitLab](https://gitlab.com/rbrt-weiler/xmc-nbi-devicelister-go), with a [copy over at GitHub](https://github.com/rbrt-weiler/xmc-nbi-devicelister-go) for the folks over there. Additionally, there is a project at GitLab which [collects all available clients](https://gitlab.com/rbrt-weiler/xmc-nbi-clients).
