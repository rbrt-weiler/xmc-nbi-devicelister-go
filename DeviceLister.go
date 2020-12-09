package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"

	godotenv "github.com/joho/godotenv"
	pflag "github.com/spf13/pflag"
	envordef "gitlab.com/rbrt-weiler/go-module-envordef"
	xmcnbiclient "gitlab.com/rbrt-weiler/go-module-xmcnbiclient"
)

const (
	toolName            string = "DeviceLister.go"
	toolVersion         string = "4.0.0-dev"
	toolID              string = toolName + "/" + toolVersion
	toolURL             string = "https://gitlab.com/rbrt-weiler/xmc-nbi-devicelister-go"
	envFileName         string = ".xmcenv"
	gqlDeviceQuery      string = "query { network { devices { up ip sysName nickName deviceData { vendor family subFamily } } } }"
	errSuccess          int    = 0
	errGeneric          int    = 1
	errUsage            int    = 2
	errClientSetup      int    = 10
	errXMCCommunication int    = 11
	errXMCResult        int    = 12
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
	pflag.CommandLine.SortFlags = false
	pflag.StringVarP(&config.XMCHost, "host", "h", envordef.StringVal("XMCHOST", ""), "XMC Hostname / IP")
	pflag.UintVar(&config.XMCPort, "port", envordef.UintVal("XMCPORT", 8443), "HTTP port where XMC is listening")
	pflag.StringVar(&config.XMCPath, "path", envordef.StringVal("XMCPATH", ""), "Path where XMC is reachable")
	pflag.UintVar(&config.HTTPTimeout, "timeout", envordef.UintVal("XMCTIMEOUT", 5), "Timeout for HTTP(S) connections")
	pflag.BoolVar(&config.NoHTTPS, "nohttps", envordef.BoolVal("XMCNOHTTPS", false), "Use HTTP instead of HTTPS")
	pflag.BoolVar(&config.InsecureHTTPS, "insecurehttps", envordef.BoolVal("XMCINSECUREHTTPS", false), "Do not validate HTTPS certificates")
	pflag.StringVarP(&config.XMCUserID, "userid", "u", envordef.StringVal("XMCUSERID", ""), "Client ID (OAuth) or username (Basic Auth) for authentication")
	pflag.StringVarP(&config.XMCSecret, "secret", "s", envordef.StringVal("XMCSECRET", ""), "Client Secret (OAuth) or password (Basic Auth) for authentication")
	pflag.BoolVar(&config.BasicAuth, "basicauth", envordef.BoolVal("XMCBASICAUTH", false), "Use HTTP Basic Auth instead of OAuth")
	pflag.BoolVar(&config.PrintVersion, "version", false, "Print version information and exit")
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", toolID)
		fmt.Fprintf(os.Stderr, "%s\n", toolURL)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "This tool lists all devices managed with XMC along with up/down information.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", path.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Available options:\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "All options that take a value can be set via environment variables:\n")
		fmt.Fprintf(os.Stderr, "  XMCHOST           -->  --host\n")
		fmt.Fprintf(os.Stderr, "  XMCPORT           -->  --port\n")
		fmt.Fprintf(os.Stderr, "  XMCPATH           -->  --path\n")
		fmt.Fprintf(os.Stderr, "  XMCTIMEOUT        -->  --timeout\n")
		fmt.Fprintf(os.Stderr, "  XMCNOHTTPS        -->  --nohttps\n")
		fmt.Fprintf(os.Stderr, "  XMCINSECUREHTTPS  -->  --insecurehttps\n")
		fmt.Fprintf(os.Stderr, "  XMCUSERID         -->  --userid\n")
		fmt.Fprintf(os.Stderr, "  XMCSECRET         -->  --secret\n")
		fmt.Fprintf(os.Stderr, "  XMCBASICAUTH      -->  --basicauth\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Environment variables can also be configured via a file called %s,\n", envFileName)
		fmt.Fprintf(os.Stderr, "located in the current directory or in the home directory of the current\n")
		fmt.Fprintf(os.Stderr, "user.\n")
		os.Exit(errUsage)
	}
	pflag.Parse()
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
		fmt.Fprintln(os.Stderr, "Variable --host must be defined. Use --help to get help.")
		os.Exit(errUsage)
	}

	client := xmcnbiclient.New(config.XMCHost)
	client.SetUserAgent(toolID)
	timeoutErr := client.SetTimeout(config.HTTPTimeout)
	if timeoutErr != nil {
		fmt.Fprintf(os.Stderr, "Could not set HTTP timeout: %s\n", timeoutErr)
		os.Exit(errClientSetup)
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
		fmt.Fprintf(os.Stderr, "Could not set port: %s\n", portErr)
		os.Exit(errClientSetup)
	}
	client.SetBasePath(config.XMCPath)
	client.UseOAuth(config.XMCUserID, config.XMCSecret)
	if config.BasicAuth {
		client.UseBasicAuth(config.XMCUserID, config.XMCSecret)
	}

	res, resErr := client.QueryAPI(gqlDeviceQuery)
	if resErr != nil {
		fmt.Fprintf(os.Stderr, "Could not query XMC: %s\n", resErr)
		os.Exit(errXMCCommunication)
	}

	devices := deviceList{}
	jsonErr := json.Unmarshal(res, &devices)
	if jsonErr != nil {
		fmt.Fprintf(os.Stderr, "Could not read result: %s\n", jsonErr)
		os.Exit(errXMCResult)
	}

	var family string
	var devName string
	var stateSym string
	var stateText string
	sort.Slice(devices.Data.Network.Devices[:], func(i int, j int) bool {
		return devices.Data.Network.Devices[i].IP < devices.Data.Network.Devices[j].IP
	})
	for _, d := range devices.Data.Network.Devices {
		family = d.DeviceData.Family
		if d.DeviceData.SubFamily != "" {
			family = family + " " + d.DeviceData.SubFamily
		}
		devName = d.SysName
		if devName == "" && d.NickName != "" {
			devName = d.NickName
		}
		stateSym = "-"
		stateText = "down"
		if d.Up {
			stateSym = "+"
			stateText = "up"
		}
		fmt.Printf("%s %s (%s %s \"%s\") is %s.\n", stateSym, d.IP, d.DeviceData.Vendor, family, devName, stateText)
	}

	os.Exit(errSuccess)
}
