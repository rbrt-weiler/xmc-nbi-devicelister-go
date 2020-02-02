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

	xmcnbiclient "gitlab.com/rbrt-weiler/go-module-xmcnbiclient"
)

const (
	toolName       string = "DeviceLister.go"
	toolVersion    string = "2.0.0-dev"
	httpUserAgent  string = toolName + "/" + toolVersion
	gqlDeviceQuery string = "query { network { devices { up ip sysName nickName deviceData { vendor family subFamily } } } }"
	errSuccess     int    = 0
)

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

func main() {
	var httpHost string
	var httpTimeoutSecs uint
	var noHTTPS bool
	var insecureHTTPS bool
	var username string
	var password string
	var clientID string
	var clientSecret string
	var printVersion bool

	flag.StringVar(&httpHost, "host", "localhost", "XMC Hostname / IP")
	flag.UintVar(&httpTimeoutSecs, "timeout", 5, "Timeout for HTTP(S) connections")
	flag.BoolVar(&noHTTPS, "nohttps", false, "Use HTTP instead of HTTPS")
	flag.BoolVar(&insecureHTTPS, "insecurehttps", false, "Do not validate HTTPS certificates")
	flag.StringVar(&username, "username", "admin", "Username for HTTP Basic Auth")
	flag.StringVar(&password, "password", "", "Password for HTTP Basic Auth")
	flag.StringVar(&clientID, "clientid", "", "Client ID for OAuth")
	flag.StringVar(&clientSecret, "clientsecret", "", "Client Secret for OAuth")
	flag.BoolVar(&printVersion, "version", false, "Print version information and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "This tool fetches lists all devices XMC knows about with up/down information.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", path.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Available options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "OAuth will be preferred over username/password.\n")
	}
	flag.Parse()

	if printVersion {
		fmt.Println(httpUserAgent)
		os.Exit(errSuccess)
	}

	client := xmcnbiclient.New(httpHost)
	client.SetUserAgent(httpUserAgent)
	client.UseHTTPS()
	if noHTTPS {
		client.UseHTTP()
	}
	client.UseBasicAuth(username, password)
	if clientID != "" && clientSecret != "" {
		client.UseOAuth(clientID, clientSecret)
	}
	client.UseSecureHTTPS()
	if insecureHTTPS {
		client.UseInsecureHTTPS()
	}
	timeoutErr := client.SetTimeout(httpTimeoutSecs)
	if timeoutErr != nil {
		log.Fatalf("Could not set HTTP timeout: %s\n", timeoutErr)
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
}
