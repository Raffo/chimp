package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/golang/glog"
	"github.com/zalando-techmonkeys/chimp/api"
	"github.com/zalando-techmonkeys/chimp/conf"
)

//Buildstamp and Githash are used to set information at build time regarding
//the version of the build.
//Buildstamp is used for storing the timestamp of the build
var Buildstamp string = "Not set"

//Githash is used for storing the commit hash of the build
var Githash string = "Not set"

var serverConfig *conf.Config

func init() {
	bin := path.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Usage of %s
================
Example:
  %% %s
`, bin, bin)
		flag.PrintDefaults()
	}
	serverConfig = conf.New()
	serverConfig.VersionBuildStamp = Buildstamp
	serverConfig.VersionGitHash = Githash
	//config from file is loaded.
	//the values will be overwritten by command line flags
	flag.BoolVar(&serverConfig.DebugEnabled, "debug", serverConfig.DebugEnabled, "Enable debug output")
	flag.BoolVar(&serverConfig.Oauth2Enabled, "oauth", serverConfig.Oauth2Enabled, "Enable OAuth2")
	flag.IntVar(&serverConfig.TeamAuthorization, "team-auth", serverConfig.TeamAuthorization, "Enable team based authorization")
	flag.StringVar(&serverConfig.AuthURL, "oauth-authurl", serverConfig.AuthURL, "OAuth2 Auth URL")
	flag.StringVar(&serverConfig.TokenURL, "oauth-tokeninfourl", serverConfig.TokenURL, "OAuth2 Auth URL")
	flag.StringVar(&serverConfig.TlsCertfilePath, "tls-cert", serverConfig.TlsCertfilePath, "TLS Certfile")
	flag.StringVar(&serverConfig.TlsKeyfilePath, "tls-key", serverConfig.TlsKeyfilePath, "TLS Keyfile")
	flag.IntVar(&serverConfig.Port, "port", serverConfig.Port, "Listening TCP Port of the service.")
	if serverConfig.Port == 0 {
		serverConfig.Port = 8082 //default port when no option is provided
	}
	flag.DurationVar(&serverConfig.LogFlushInterval, "flush-interval", time.Second*5, "Interval to flush Logs to disk.")
}

func main() {
	flag.Parse()

	// default https, if cert and key are found
	var err error
	httpOnly := false
	if _, err = os.Stat(serverConfig.TlsCertfilePath); os.IsNotExist(err) {
		glog.Warningf("WARN: No Certfile found %s\n", serverConfig.TlsCertfilePath)
		httpOnly = true
	} else if _, err = os.Stat(serverConfig.TlsKeyfilePath); os.IsNotExist(err) {
		glog.Warningf("WARN: No Keyfile found %s\n", serverConfig.TlsKeyfilePath)
		httpOnly = true
	}
	var keypair tls.Certificate
	if httpOnly {
		keypair = tls.Certificate{}
	} else {
		keypair, err = tls.LoadX509KeyPair(serverConfig.TlsCertfilePath, serverConfig.TlsKeyfilePath)
		if err != nil {
			fmt.Printf("ERR: Could not load X509 KeyPair, caused by: %s\n", err)
			os.Exit(1)
		}
	}

	// configure service
	cfg := api.ServerSettings{
		Configuration: serverConfig,
		CertKeyPair:   keypair,
		Httponly:      httpOnly,
	}
	svc := api.Service{}
	svc.Run(cfg)
}
