//handles the configuration of the applications. Yaml files are mapped with the struct

package client

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

type ClientConfig struct {
	Clusters      map[string]*Cluster
	Port          int    //URL of the backend
	HttpOnly      bool   //true if we must use only http and not https for request (security not enabled!)
	Oauth2Enabled bool   //true if oauth2 is enabled
	OauthURL      string //the oauth2 endpoint to be used
	TokenURL      string //the oauth2 token info endpoint
}

type Cluster struct {
	Ip   string
	Port int
}

//shared state for configuration
var clientConf *ClientConfig

//GetConfig gets the loaded configuration
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
