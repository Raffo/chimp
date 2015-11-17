package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/spf13/viper"
	"github.com/vrischmann/envconfig"
	"github.com/zalando-techmonkeys/chimp/client"
	konfig "github.com/zalando-techmonkeys/chimp/conf/client"
	"golang.org/x/oauth2"
)

//Buildstamp and Githash are used to set information at build time regarding
//the version of the build.
//Buildstamp is used for storing the timestamp of the build
var Buildstamp string = "Not set"

//Githash is used for storing the commit hash of the build
var Githash string = "Not set"

var DEBUG bool
var conf struct {
	AccessUser     string `envconfig:"optional"`
	AccessPassword string `envconfig:"optional"`
	AccessToken    string `envconfig:"optional"`
	OAuth2Endpoint struct {
		AuthURL      string `envconfig:"optional"`
		TokenInfoURL string `envconfig:"optional"`
	}
}

func main() {

	usage := fmt.Sprintf(`Usage:
  chimp -h | --help
  chimp --version
  chimp create (--file=<filename> | <name> <url> --port=<svcport> --memory=<memory> --cpu=<cpu-number> --replicas=<replicas>) [options]
  chimp update (--file=<filename> | <name> <url> --port=<svcport> --memory=<memory> --cpu=<cpu-number> --replicas=<replicas>) [options]
  chimp scale (<name>) (<replicas>) [options]
  chimp delete (<name>) [options]
  chimp info (<name>) [options]
  chimp list [--all] [options]
  chimp login (<username>) [options]


Options:
  --label=<k=v>  Labels of the deploy artifact, has to be a dict like k=v
  --env=<k=v>  Environment variables of the deploy artifact, has to be a dict like k=v
  --svcport=<svcport>  Port to listen on
  --reqserver=<reqserver>  Server to query for deploy request
  --reqport=<reqport>  Port to use in query for deploy request
  --http-only  If not set we use https as default to query deploy requests
  --oauth2  OAuth2 enable
  --oauth2-token=<access_token>  OAuth2 AccessToken (no user, password required)
  --oauth2-authurl=<oauth2_authurl>  OAuth2 endpoint that issue AccessTokens
  --debug  Debug
  --verbose  Verbose logging
  --force  Force deployment
`)

	arguments, err := docopt.Parse(usage, nil, true, fmt.Sprintf("%s Build Time: %s - Git Commit Hash: %s", os.Args[0], Buildstamp, Githash), false)
	if err != nil {
		panic("Could not parse CLI")
	}

	DEBUG = arguments["--debug"].(bool)
	var verbose bool = arguments["--verbose"].(bool)
	// Auth information from ENV and parameter
	if err := envconfig.Init(&conf); err != nil {
		fmt.Printf("ERR: envconfig failed, caused by: %s\n", err)
	}
	cli := createClient(arguments)
	name := GetStringFromArgs(arguments, "<name>", "")
	username := GetStringFromArgs(arguments, "<username>", "")
	if arguments["create"].(bool) {
		cli.GetAccessToken(username)
		cmdReq, err := buildRequest(arguments)
		if err != nil {
			fmt.Println("Cannot parse, please provide valid options.")
			os.Exit(1)
		}
		cli.CreateDeploy(cmdReq)
	} else if arguments["delete"].(bool) {
		cli.GetAccessToken(username)
		cli.DeleteDeploy(name)
	} else if arguments["info"].(bool) {
		cli.GetAccessToken(username)
		cli.InfoDeploy(name, verbose)
	} else if arguments["list"].(bool) {
		cli.GetAccessToken(username)
		all := arguments["--all"].(bool)
		cli.ListDeploy(all)
	} else if arguments["update"].(bool) {
		cli.GetAccessToken(username)
		cmdReq, err := buildRequest(arguments)
		if err != nil {
			fmt.Println("Cannot parse, please provide valid options.")
			os.Exit(1)
		}
		cli.UpdateDeploy(cmdReq)
	} else if arguments["scale"].(bool) {
		cli.GetAccessToken(username)
		replicas := GetIntFromArgs(arguments, "<replicas>", 1)
		cli.Scale(name, replicas)
	} else if arguments["login"].(bool) {
		cli.RenewAccessToken(strings.TrimSpace(username))
	}
}

func createClient(arguments map[string]interface{}) client.Client {
	//loading configuration from file. it is overridden by the command line parameters
	configuration := konfig.New()
	server := configuration.Server
	port := configuration.Port
	_server := arguments["--reqserver"]
	_port := arguments["--reqport"]
	//defaulting to some values if this is present
	if _server != nil {
		server = _server.(string)
	} else if server == "" {
		server = "127.0.0.1"
	}
	if _port != nil {
		var s string
		var _err error
		s = _port.(string)
		if port, _err = strconv.Atoi(s); _err != nil {
			fmt.Printf("ERR: int conversion failed for port: %s, err: %s", _port, _err)
		}
	} else if port == 0 {
		port = 8082
	}
	var scheme string = "https"
	httpOnly := konfig.New().HttpOnly
	if httpOnly {
		scheme = "http"
	}
	//this overwrites the setting in the config file
	if arguments["--http-only"].(bool) {
		scheme = "http"
	}

	var oauth2Enabled bool
	if arguments["--oauth2"].(bool) {
		oauth2Enabled = true
	} else {
		oauth2Enabled = configuration.Oauth2Enabled
	}

	accessToken := GetStringFromArgs(arguments, "--oauth2-token", "")
	oauth2Authurl := GetStringFromArgs(arguments, "--oauth2-authurl", configuration.OauthURL)
	oauth2Tokeninfourl := GetStringFromArgs(arguments, "--oauth2-tokeninfourl", configuration.TokenURL)
	oauth2Endpoint := oauth2.Endpoint{AuthURL: oauth2Authurl, TokenURL: oauth2Tokeninfourl}

	return client.Client{Host: server,
		Port: port, Scheme: scheme,
		Oauth2Enabled:  oauth2Enabled,
		AccessToken:    accessToken,
		OAuth2Endpoint: oauth2Endpoint,
	}
}

//FIXME dirty hack to make labels work
func buildRequest(arguments map[string]interface{}) (*client.CmdClientRequest, error) {
	//reading configuration file
	var c client.CmdClientRequest
	fileName := GetStringFromArgs(arguments, "--file", "")
	if fileName != "" {
		viper := viper.New()
		viper.SetConfigFile(fileName)
		err := viper.ReadInConfig()
		if err != nil {
			fmt.Printf("Can not read config, caused by: %s\n", err)
			return nil, err
		}
		err = viper.Marshal(&c)
	} else {
		labelStr := GetStringFromArgs(arguments, "--label", "")
		name := GetStringFromArgs(arguments, "<name>", "")
		labels := ConvertMaps(labelStr)
		envStr := GetStringFromArgs(arguments, "--env", "")
		envVars := ConvertMaps(envStr)
		replicas := GetIntFromArgs(arguments, "--replicas", 1)
		imageURL := GetStringFromArgs(arguments, "<url>", "")
		svcport := GetIntFromArgs(arguments, "--svcport", 8080)
		cpuNumber := GetIntFromArgs(arguments, "--cpu", 0)            //unlimited or backend decided
		memoryLimit := GetStringFromArgs(arguments, "--memory", "0M") //unlimited or backend decided

		ports := []int{svcport}
		c.Labels = labels
		c.Env = envVars
		c.Replicas = replicas
		c.ImageURL = imageURL
		c.CPULimit = cpuNumber
		c.MemoryLimit = memoryLimit
		c.Ports = ports
		c.Name = name
	}

	var force bool = arguments["--force"].(bool)
	c.Force = force
	return &c, nil

}
