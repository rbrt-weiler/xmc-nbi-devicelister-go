package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"

	godotenv "github.com/joho/godotenv"
	pflag "github.com/spf13/pflag"
	consolehelper "gitlab.com/rbrt-weiler/go-module-consolehelper"
	envordef "gitlab.com/rbrt-weiler/go-module-envordef"
	xmcnbiclient "gitlab.com/rbrt-weiler/go-module-xmcnbiclient"
)

// Definitions used within the code.
const (
	toolName       string = "DeviceLister.go"
	toolVersion    string = "4.0.0"
	toolID         string = toolName + "/" + toolVersion
	toolURL        string = "https://gitlab.com/rbrt-weiler/xmc-nbi-devicelister-go"
	envFileName    string = ".xmcenv"
	gqlDeviceQuery string = "query { network { devices { up ip sysName nickName deviceData { vendor family subFamily } } } }"
)

// Error codes.
const (
	errSuccess          int = 0  // No error
	errGeneric          int = 1  // Generic error
	errUsage            int = 2  // Usage error
	errClientSetup      int = 10 // Error creating an API client
	errXMCCommunication int = 11 // Error while querying XMC
	errXMCResult        int = 12 // Error parsing the rsult returned by XMC
)

// appConfig stores the application configuration once parsed by flags.
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

// deviceList mimics the data structure returned by XMC for easy parsing into Go variables.
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

// Global variables used throughout the program.
var (
	config appConfig
	cons   consolehelper.ConsoleHelper
)

// parseCLIOptions parses all options passed by env or CLI into the config variable.
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
		cons.Fprintf(os.Stderr, "%s\n", toolID)
		cons.Fprintf(os.Stderr, "%s\n", toolURL)
		cons.Fprintf(os.Stderr, "\n")
		cons.Fprintf(os.Stderr, "This tool lists all devices managed with XMC along with up/down information.\n")
		cons.Fprintf(os.Stderr, "\n")
		cons.Fprintf(os.Stderr, "Usage: %s [options]\n", path.Base(os.Args[0]))
		cons.Fprintf(os.Stderr, "\n")
		cons.Fprintf(os.Stderr, "Available options:\n")
		pflag.PrintDefaults()
		cons.Fprintf(os.Stderr, "\n")
		cons.Fprintf(os.Stderr, "All options that take a value can be set via environment variables:\n")
		cons.Fprintf(os.Stderr, "  XMCHOST           -->  --host\n")
		cons.Fprintf(os.Stderr, "  XMCPORT           -->  --port\n")
		cons.Fprintf(os.Stderr, "  XMCPATH           -->  --path\n")
		cons.Fprintf(os.Stderr, "  XMCTIMEOUT        -->  --timeout\n")
		cons.Fprintf(os.Stderr, "  XMCNOHTTPS        -->  --nohttps\n")
		cons.Fprintf(os.Stderr, "  XMCINSECUREHTTPS  -->  --insecurehttps\n")
		cons.Fprintf(os.Stderr, "  XMCUSERID         -->  --userid\n")
		cons.Fprintf(os.Stderr, "  XMCSECRET         -->  --secret\n")
		cons.Fprintf(os.Stderr, "  XMCBASICAUTH      -->  --basicauth\n")
		cons.Fprintf(os.Stderr, "\n")
		cons.Fprintf(os.Stderr, "Environment variables can also be configured via a file called %s, located in the current directory or in the home directory of the current user.\n", envFileName)
		os.Exit(errUsage)
	}
	pflag.Parse()
}

// init sets up consoleHelper and loads environment files if available.
func init() {
	// initialize console size
	cons.UpdateDimensions()

	// if envFileName exists in the current directory, load it
	localEnvFile := fmt.Sprintf("./%s", envFileName)
	if _, localEnvErr := os.Stat(localEnvFile); localEnvErr == nil {
		if loadErr := godotenv.Load(localEnvFile); loadErr != nil {
			cons.Fprintf(os.Stderr, "Could not load env file <%s>: %s\n", localEnvFile, loadErr)
		}
	}

	// if envFileName exists in the user's home directory, load it
	if homeDir, homeErr := os.UserHomeDir(); homeErr == nil {
		homeEnvFile := fmt.Sprintf("%s/%s", homeDir, ".xmcenv")
		if _, homeEnvErr := os.Stat(homeEnvFile); homeEnvErr == nil {
			if loadErr := godotenv.Load(homeEnvFile); loadErr != nil {
				cons.Fprintf(os.Stderr, "Could not load env file <%s>: %s\n", homeEnvFile, loadErr)
			}
		}
	}
}

// main ties everything together.
func main() {
	// Parse all valid CLI options into variables.
	parseCLIOptions()

	// Print version information and exit.
	if config.PrintVersion {
		cons.Println(toolID)
		os.Exit(errSuccess)
	}
	// Check that the option "host" has been set.
	if config.XMCHost == "" {
		cons.Fprintln(os.Stderr, "Variable --host must be defined. Use --help to get help.")
		os.Exit(errUsage)
	}

	// Set up a NBI client
	client := xmcnbiclient.New(config.XMCHost)
	client.SetUserAgent(toolID)
	timeoutErr := client.SetTimeout(config.HTTPTimeout)
	if timeoutErr != nil {
		cons.Fprintf(os.Stderr, "Could not set HTTP timeout: %s\n", timeoutErr)
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
		cons.Fprintf(os.Stderr, "Could not set port: %s\n", portErr)
		os.Exit(errClientSetup)
	}
	client.SetBasePath(config.XMCPath)
	client.UseOAuth(config.XMCUserID, config.XMCSecret)
	if config.BasicAuth {
		client.UseBasicAuth(config.XMCUserID, config.XMCSecret)
	}

	// Call the API.
	res, resErr := client.QueryAPI(gqlDeviceQuery)
	if resErr != nil {
		cons.Fprintf(os.Stderr, "Could not query XMC: %s\n", resErr)
		os.Exit(errXMCCommunication)
	}

	// Parse the API result into a Go sructure.
	devices := deviceList{}
	jsonErr := json.Unmarshal(res, &devices)
	if jsonErr != nil {
		cons.Fprintf(os.Stderr, "Could not read result: %s\n", jsonErr)
		os.Exit(errXMCResult)
	}

	// Work on the data and print the result.
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
		cons.Printf("%s %s (%s %s \"%s\") is %s.\n", stateSym, d.IP, d.DeviceData.Vendor, family, devName, stateText)
	}

	// Exit with an appropriate exit code.
	os.Exit(errSuccess)
}
