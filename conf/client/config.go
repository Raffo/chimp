//handles the configuration of the applications. Yaml files are mapped with the struct

package client

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

//ClientConfig is the configuration from the client. Usually loaded from config files.
type ClientConfig struct {
	Clusters      map[string]*Cluster //map of clusters by name.
	Port          int                 //URL of the backend
	HTTPOnly      bool                //true if we must use only http and not https for request (security not enabled!)
	Oauth2Enabled bool                //true if oauth2 is enabled
	OauthURL      string              //the oauth2 endpoint to be used
	TokenURL      string              //the oauth2 token info endpoint
}

//Cluster is used to represent the main endpoint of a chimp server, used to target a specific cluster
type Cluster struct {
	IP   string
	Port int
}

//shared state for configuration
var clientConf *ClientConfig

//New gets the ClientConfiguration
func New() *ClientConfig {
	if clientConf == nil {
		c, err := clientConfigInit("config.yaml")
		clientConf = c
		if err != nil {
			glog.Errorf("could not load configuration. Reason: %s", err)
			panic("Cannot load configuration. Exiting.")
		}
	}

	return clientConf
}

func clientConfigInit(filename string) (*ClientConfig, error) {
	viper := viper.New()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/chimp")
	viper.AddConfigPath(fmt.Sprintf("%s/.config/chimp", os.ExpandEnv("$HOME")))
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Can not read config, caused by: %s", err)
		return nil, err
	}
	var c ClientConfig
	err = viper.Marshal(&c)
	if err != nil {
		fmt.Printf("Can not marshal config, caused by: %s", err)
		return nil, err
	}

	return &c, nil
}
