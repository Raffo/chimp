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
	. "github.com/zalando-techmonkeys/chimp/types"
)

//Buildstamp and Githash are used to set information at build time regarding
//the version of the build.
//Buildstamp is used for storing the timestamp of the build
var Buildstamp = "Not set"

//Githash is used for storing the commit hash of the build
var Githash = "Not set"

//DEBUG enables debug mode in the cli. Used for verbose printing //TODO: currently brokend
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
  chimp create (<filename> | <name> <url> --port=<port> --memory=<memory> --cpu=<cpu-number> --replicas=<replicas>) [options]
  chimp update (<filename> | <name> <url> --port=<port> --memory=<memory> --cpu=<cpu-number> --replicas=<replicas> ) [--cluster=<cluster>] [options]
  chimp scale (<name>) (<replicas>) [--cluster=<cluster>] [options]
  chimp delete (<name>) [--cluster=<cluster>] [options]
  chimp info (<name>) [--cluster=<cluster>] [options]
  chimp list [--all] [--cluster=<cluster>] [options]
  chimp login (<username>) [options]


Options:
  --label=<k=v>  Labels of the deploy artifact, has to be a dict like k=v
  --env=<k=v>  Environment variables of the deploy artifact, has to be a dict like k=v
  --http-only  If not set we use https as default to query deploy requests
  --oauth2  OAuth2 enable
  --oauth2-token=<access_token>  OAuth2 AccessToken (no user, password required)
  --oauth2-authurl=<oauth2_authurl>  OAuth2 endpoint that issue AccessTokens
  --debug  Debug
  --verbose  Verbose logging
  --force  Force deployment
  --cluster=<cluster> The endpoint of the cluster. "all" means deployed on every cluster in the config.
`)

	arguments, err := docopt.Parse(usage, nil, true, fmt.Sprintf("%s Build Time: %s - Git Commit Hash: %s", os.Args[0], Buildstamp, Githash), false)
	if err != nil {
		panic("Could not parse CLI")
	}

	DEBUG = arguments["--debug"].(bool)
	var verbose = arguments["--verbose"].(bool)
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
		cli.CreateDeploy(&cmdReq.DeployRequest[0])
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
		cli.UpdateDeploy(&cmdReq.DeployRequest[0])
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
	clusterName := (arguments["--cluster"])
	configuration := konfig.New()
	clusters := []string{} //array of endpoints
	if clusterName == nil || clusterName == "all" || clusterName == "ALL" {
		//deploying to all clusters
		for k := range configuration.Clusters {
			clusters = append(clusters, k)
		}
	} else {
		if configuration.Clusters[clusterName.(string)] == nil {
			fmt.Printf("Cluster name is invalid.\n")
			os.Exit(-1)
		}
		clusters = append(clusters, clusterName.(string))
	}

	//FIXME issue with CLI params because of multi-cluster behaviour
	port := configuration.Port
	_server := arguments["--reqserver"]
	_port := arguments["--reqport"]
	//defaulting to some values if this is present
	if _server != nil {
		clusters = []string{_server.(string)}
	} else if len(clusters) == 0 {
		clusters = []string{"127.0.0.1"}
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
	var scheme = "https"
	httpOnly := konfig.New().HTTPOnly
	if httpOnly {
		scheme = "http"
	}
	//this overwrites the setting in the config file
	if arguments["--http-only"].(bool) {
		scheme = "http"
	}

	accessToken := GetStringFromArgs(arguments, "--oauth2-token", "")

	return client.Client{
		Clusters:    clusters,
		Config:      configuration,
		Scheme:      scheme,
		AccessToken: accessToken,
	}
}

func buildRequest(arguments map[string]interface{}) (*ChimpDefinition, error) {
	//reading configuration file
	var c ChimpDefinition
	fileName := GetStringFromArgs(arguments, "<filename>", "")
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
		svcport := GetIntFromArgs(arguments, "--port", 8080)
		cpuNumber := GetIntFromArgs(arguments, "--cpu", 0)            //unlimited or backend decided
		memoryLimit := GetStringFromArgs(arguments, "--memory", "0M") //unlimited or backend decided

		ports := []int{svcport}
		c.DeployRequest = append(c.DeployRequest, CmdClientRequest{})
		c.DeployRequest[0].Labels = labels
		c.DeployRequest[0].Env = envVars
		c.DeployRequest[0].Replicas = replicas
		c.DeployRequest[0].ImageURL = imageURL
		c.DeployRequest[0].CPULimit = cpuNumber
		c.DeployRequest[0].MemoryLimit = memoryLimit
		c.DeployRequest[0].Ports = ports
		c.DeployRequest[0].Name = name
	}

	return &c, nil

}
