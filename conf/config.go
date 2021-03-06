//handles the configuration of the applications. Yaml files are mapped with the struct

package conf

import (
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

//constants for authorization
const (
	NoAuth         = iota //0
	IndividualAuth        //1
	TeamAuth              //2
)

//Config is the current configuration for the server. It's mapped to a yaml file
type Config struct {
	BackendType       string //"marathon" or "kubernetes"
	Endpoint          string //URL of the backend
	FluentdEnabled    bool   //true if fluentd is enabled, will be ON for each container
	DebugEnabled      bool
	Oauth2Enabled     bool //true if authentication is enabled
	AuthorizationType int
	AuthURL           string
	TokenURL          string
	TLSCertfilePath   string
	TLSKeyfilePath    string
	LogFlushInterval  time.Duration
	Port              int
	AuthorizedTeams   []AccessTuple
	AuthorizedUsers   []AccessTuple
	VersionBuildStamp string
	VersionGitHash    string
	MarathonAuth      MarathonAuth
	EndpointPattern   string
}

//AccessTuple reprsent an entry for Auth
type AccessTuple struct {
	Realm string
	UID   string
	Cn    string
}

//ConfigError contains the error while unmarshalling the config file
type ConfigError struct {
	Message string
}

//MarathonAuth is used to enable/disable marathon api auth and configure user/password used for that
type MarathonAuth struct {
	Enabled              bool
	MarathonHttpUser     string
	MarathonHttpPassword string
}

//shared state for configuration
var conf *Config

//New gets an instance of the loaded configuration
func New() *Config {
	var err *ConfigError
	if conf == nil {
		conf, err = configInit("config.yaml")
		if err != nil {
			glog.Errorf("could not load configuration. Reason: %s", err.Message)
		}
	}
	return conf
}

func configInit(filename string) (*Config, *ConfigError) {
	var config Config
	viper := viper.New()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/chimp-server")
	viper.AddConfigPath(fmt.Sprintf("%s/.config/chimp-server", os.ExpandEnv("$HOME")))
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Can not read config, caused by: %s", err)
		return &config, &ConfigError{"configuration format is not correct."}
	}
	err = viper.Marshal(&config)
	if err != nil {
		fmt.Printf("Can not marshal config, caused by: %s", err)
		return &config, &ConfigError{"cannot read configuration, something must be wrong."}
	}
	return &config, nil
}
