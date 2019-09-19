/*
Copyright (c) 2019 BELL Computer-Netzwerke GmbH
Copyright (c) 2019 Robert Weiler

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
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const ToolName string = "BELL XMC NBI DeviceLister"
const ToolVersion string = "1.0"
const HttpUserAgent string = ToolName + "/" + ToolVersion
const GqlDeviceQuery string = `query {
	network {
	  devices {
		up
		ip
		sysName
		deviceData {
		  vendor
		  family
		  subFamily
		}
	  }
	}
  }`

// created with https://mholt.github.io/json-to-go/
type DeviceList struct {
	Data struct {
		Network struct {
			Devices []struct {
				Up         bool   `json:"up"`
				IP         string `json:"ip"`
				SysName    string `json:"sysName"`
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
	var host string
	var httpTimeout int
	var username string
	var password string

	flag.StringVar(&host, "host", "localhost", "XMC Hostname / IP")
	flag.IntVar(&httpTimeout, "httptimeout", 5, "Timeout for HTTP(S) connections")
	flag.StringVar(&username, "username", "admin", "Username for HTTP auth")
	flag.StringVar(&password, "password", "", "Password for HTTP auth")
	flag.Parse()

	var apiUrl string = "https://" + host + ":8443/nbi/graphql"

	nbiClient := http.Client {
		Timeout: time.Second * time.Duration(httpTimeout),
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", HttpUserAgent)
	req.SetBasicAuth(username, password)

	httpQuery := req.URL.Query()
	httpQuery.Add("query", GqlDeviceQuery)
	req.URL.RawQuery = httpQuery.Encode()

	res, getErr := nbiClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	devices := DeviceList{}
	jsonErr := json.Unmarshal(body, &devices)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	var family string
	for _, d := range devices.Data.Network.Devices {
		if d.DeviceData.SubFamily != "" {
			family = d.DeviceData.Family + " " + d.DeviceData.SubFamily
		} else {
			family = d.DeviceData.Family
		}
		switch d.Up {
		case true:
			fmt.Printf("+ %s (%s %s \"%s\") is up.\n", d.IP, d.DeviceData.Vendor, family, d.SysName)
		default:
			fmt.Printf("- %s (%s %s \"%s\") is down.\n", d.IP, d.DeviceData.Vendor, family, d.SysName)
		}
	}
}
