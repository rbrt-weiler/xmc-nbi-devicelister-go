/*
Copyright (c) 2019,2020 Robert Weiler <https://robert.weiler.one/>
Copyright (c) 2019 BELL Computer-Netzwerke GmbH <https://www.bell.de/>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	godotenv "github.com/joho/godotenv"
	envordef "gitlab.com/rbrt-weiler/go-module-envordef"
	xmcnbiclient "gitlab.com/rbrt-weiler/go-module-xmcnbiclient"
)

const (
	toolName       string = "DeviceLister.go"
	toolVersion    string = "3.0.0-dev"
	toolID         string = toolName + "/" + toolVersion
	envFileName    string = ".xmcenv"
	gqlDeviceQuery string = "query { network { devices { up ip sysName nickName deviceData { vendor family subFamily } } } }"
	errSuccess     int    = 0
	errGeneric     int    = 1
	errUsage       int    = 2
)

type appConfig struct {
	XMCHost       string
	XMCPort       uint
	XMCPath       string
	HTTPTimeout   uint
	NoHTTPS       bool
	InsecureHTTPS bool
	BasicAuth     bool
	XMCUserID     string
	XMCSecret     string
	PrintVersion  bool
}

// created with https://mholt.github.io/json-to-go/
type deviceList struct {
	Data struct {
		Network struct {
			Devices []struct {
				Up         bool   `json:"up"`
				IP         string `json:"ip"`
				SysName    string `json:"sysName"`
				NickName   string `json:"nickName"`
				DeviceData struct {
					Vendor    string `json:"vendor"`
					Family    string `json:"family"`
					SubFamily string `json:"subFamily"`
				} `json:"deviceData"`
			} `json:"devices"`
		} `json:"network"`
	} `json:"data"`
}

var (
	config appConfig
)

func parseCLIOptions() {
	flag.StringVar(&config.XMCHost, "host", envordef.StringVal("XMCHOST", ""), "XMC Hostname / IP")
	flag.UintVar(&config.XMCPort, "port", envordef.UintVal("XMCPORT", 8443), "HTTP port where XMC is listening")
	flag.StringVar(&config.XMCPath, "path", envordef.StringVal("XMCPATH", ""), "Path where XMC is reachable")
	flag.UintVar(&config.HTTPTimeout, "timeout", envordef.UintVal("XMCTIMEOUT", 5), "Timeout for HTTP(S) connections")
	flag.BoolVar(&config.NoHTTPS, "nohttps", envordef.BoolVal("XMCNOHTTPS", false), "Use HTTP instead of HTTPS")
	flag.BoolVar(&config.InsecureHTTPS, "insecurehttps", envordef.BoolVal("XMCINSECUREHTTPS", false), "Do not validate HTTPS certificates")
	flag.StringVar(&config.XMCUserID, "userid", envordef.StringVal("XMCUSERID", ""), "Client ID (OAuth) or username (Basic Auth) for authentication")
	flag.StringVar(&config.XMCSecret, "secret", envordef.StringVal("XMCSECRET", ""), "Client Secret (OAuth) or password (Basic Auth) for authentication")
	flag.BoolVar(&config.BasicAuth, "basicauth", envordef.BoolVal("XMCBASICAUTH", false), "Use HTTP Basic Auth instead of OAuth")
	flag.BoolVar(&config.PrintVersion, "version", false, "Print version information and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "This tool queries the XMC API and prints the raw reply (JSON) to stdout.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", path.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Available options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "All options that take a value can be set via environment variables:\n")
		fmt.Fprintf(os.Stderr, "  XMCHOST           -->  -host\n")
		fmt.Fprintf(os.Stderr, "  XMCPORT           -->  -port\n")
		fmt.Fprintf(os.Stderr, "  XMCPATH           -->  -path\n")
		fmt.Fprintf(os.Stderr, "  XMCTIMEOUT        -->  -timeout\n")
		fmt.Fprintf(os.Stderr, "  XMCNOHTTPS        -->  -nohttps\n")
		fmt.Fprintf(os.Stderr, "  XMCINSECUREHTTPS  -->  -insecurehttps\n")
		fmt.Fprintf(os.Stderr, "  XMCUSERID         -->  -userid\n")
		fmt.Fprintf(os.Stderr, "  XMCSECRET         -->  -secret\n")
		fmt.Fprintf(os.Stderr, "  XMCBASICAUTH      -->  -basicauth\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Environment variables can also be configured via a file called %s,\n", envFileName)
		fmt.Fprintf(os.Stderr, "located in the current directory or in the home directory of the current\n")
		fmt.Fprintf(os.Stderr, "user.\n")
		os.Exit(errUsage)
	}
	flag.Parse()
}

func init() {
	// if envFileName exists in the current directory, load it
	localEnvFile := fmt.Sprintf("./%s", envFileName)
	if _, localEnvErr := os.Stat(localEnvFile); localEnvErr == nil {
		if loadErr := godotenv.Load(localEnvFile); loadErr != nil {
			fmt.Fprintf(os.Stderr, "Could not load env file <%s>: %s", localEnvFile, loadErr)
		}
	}

	// if envFileName exists in the user's home directory, load it
	if homeDir, homeErr := os.UserHomeDir(); homeErr == nil {
		homeEnvFile := fmt.Sprintf("%s/%s", homeDir, ".xmcenv")
		if _, homeEnvErr := os.Stat(homeEnvFile); homeEnvErr == nil {
			if loadErr := godotenv.Load(homeEnvFile); loadErr != nil {
				fmt.Fprintf(os.Stderr, "Could not load env file <%s>: %s", homeEnvFile, loadErr)
			}
		}
	}
}

func main() {
	parseCLIOptions()

	if config.PrintVersion {
		fmt.Println(toolID)
		os.Exit(errSuccess)
	}
	if config.XMCHost == "" {
		fmt.Fprintln(os.Stderr, "Variable -host must be defined. Use -h to get help.")
		os.Exit(errUsage)
	}

	client := xmcnbiclient.New(config.XMCHost)
	client.SetUserAgent(toolID)
	timeoutErr := client.SetTimeout(config.HTTPTimeout)
	if timeoutErr != nil {
		log.Fatalf("Could not set HTTP timeout: %s\n", timeoutErr)
	}
	client.UseSecureHTTPS()
	if config.InsecureHTTPS {
		client.UseInsecureHTTPS()
	}
	client.UseHTTPS()
	if config.NoHTTPS {
		client.UseHTTP()
	}
	portErr := client.SetPort(config.XMCPort)
	if portErr != nil {
		fmt.Fprintf(os.Stderr, "Port could not be set: %s\n", portErr)
		os.Exit(errGeneric)
	}
	client.SetBasePath(config.XMCPath)
	client.UseOAuth(config.XMCUserID, config.XMCSecret)
	if config.BasicAuth {
		client.UseBasicAuth(config.XMCUserID, config.XMCSecret)
	}

	res, resErr := client.QueryAPI(gqlDeviceQuery)
	if resErr != nil {
		log.Fatal(resErr)
	}

	devices := deviceList{}
	jsonErr := json.Unmarshal(res, &devices)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	var family string
	var devName string
	for _, d := range devices.Data.Network.Devices {
		family = d.DeviceData.Family
		devName = d.SysName
		if d.DeviceData.SubFamily != "" {
			family = family + " " + d.DeviceData.SubFamily
		}
		if devName == "" && d.NickName != "" {
			devName = d.NickName
		}
		switch d.Up {
		case true:
			fmt.Printf("+ %s (%s %s \"%s\") is up.\n", d.IP, d.DeviceData.Vendor, family, devName)
		default:
			fmt.Printf("- %s (%s %s \"%s\") is down.\n", d.IP, d.DeviceData.Vendor, family, devName)
		}
	}

	os.Exit(errSuccess)
}
