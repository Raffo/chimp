package api

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/zalando-techmonkeys/chimp/conf"
	"github.com/zalando-techmonkeys/gin-glog"
	"github.com/zalando-techmonkeys/gin-gomonitor"
	"github.com/zalando-techmonkeys/gin-gomonitor/aspects"
	"github.com/zalando-techmonkeys/gin-oauth2"
	"github.com/zalando-techmonkeys/gin-oauth2/zalando"
	"golang.org/x/oauth2"
	"gopkg.in/mcuadros/go-monitor.v1/aspects"
)

type ServerSettings struct {
	Configuration *conf.Config
	CertKeyPair   tls.Certificate
	Httponly      bool
}

// global data, p.e. Debug
var config ServerSettings

func dummyAuthorization(tc *ginoauth2.TokenContainer, ctx *gin.Context) bool {
	return true
}

// Main Service Struct
type Service struct{}

func (svc *Service) Run(cfg ServerSettings) error {
	config = cfg // save config in global

	// init gin
	if !config.Configuration.DebugEnabled {
		gin.SetMode(gin.ReleaseMode)
	}

	var oauth2Endpoint = oauth2.Endpoint{
		AuthURL:  config.Configuration.AuthURL,
		TokenURL: config.Configuration.TokenURL,
	}

	// Middleware
	router := gin.New()
	// use glog for logging
	router.Use(ginglog.Logger(config.Configuration.LogFlushInterval))
	// monitoring GO internals and counter middleware
	counterAspect := &ginmon.CounterAspect{0}
	asps := []aspects.Aspect{counterAspect}
	router.Use(ginmon.CounterHandler(counterAspect))
	router.Use(gomonitor.Metrics(9000, asps))
	router.Use(ginoauth2.RequestLogger([]string{"uid", "team"}, "data"))
	// last middleware
	router.Use(gin.Recovery())

	// OAuth2 secured if conf.Oauth2Enabled is set
	var private *gin.RouterGroup
	//ATM team or user auth is mutually exclusive, we have to look for a better solution
	if config.Configuration.Oauth2Enabled {
		private = router.Group("")
		if config.Configuration.TeamAuthorization == conf.TEAM_AUTH {
			var accessTuple []zalando.AccessTuple = make([]zalando.AccessTuple, len(config.Configuration.AuthorizedTeams))
			for i, v := range config.Configuration.AuthorizedTeams {
				accessTuple[i] = zalando.AccessTuple{Realm: v.Realm, Uid: v.Uid, Cn: v.Cn}
			}
			zalando.AccessTuples = accessTuple
			private.Use(ginoauth2.Auth(zalando.GroupCheck, oauth2Endpoint))
		} else if config.Configuration.TeamAuthorization == conf.INDIVIDUAL_AUTH {
			var accessTuple []zalando.AccessTuple = make([]zalando.AccessTuple, len(config.Configuration.AuthorizedUsers))
			for i, v := range config.Configuration.AuthorizedUsers {
				accessTuple[i] = zalando.AccessTuple{Realm: v.Realm, Uid: v.Uid, Cn: v.Cn}
			}
			private.Use(ginoauth2.Auth(zalando.UidCheck, oauth2Endpoint))
		} else { //NO_AUTH
			private.Use(ginoauth2.Auth(dummyAuthorization, oauth2Endpoint))
		}
	}

	//non authenticated routes
	router.GET("/", rootHandler)
	router.GET("/health", healthHandler)
	//authenticated routes
	if config.Configuration.Oauth2Enabled {
		private.GET("/deployments", deployList)
		private.GET("/deployments/:name", deployInfo)
		private.POST("/deployments", deployCreate)
		private.PUT("/deployments/:name", deployUpsert)
		private.DELETE("/deployments/:name", deployDelete)
		private.PATCH("/deployments/:name/replicas/:num", deployReplicasModify)
	} else {
		router.GET("/deployments", deployList)
		router.GET("/deployments/:name", deployInfo)
		router.POST("/deployments", deployCreate)
		router.PUT("/deployments/:name", deployUpsert)
		router.DELETE("/deployments/:name", deployDelete)
		router.PATCH("/deployments/:name/replicas/:num", deployReplicasModify)
	}

	// TLS config
	var tls_config tls.Config = tls.Config{}
	if !config.Httponly {
		tls_config.Certificates = []tls.Certificate{config.CertKeyPair}
		tls_config.NextProtos = []string{"http/1.1"}
		tls_config.Rand = rand.Reader // Strictly not necessary, should be default
	}

	// run backend
	Start()

	// run frontend server
	serve := &http.Server{
		Addr:      fmt.Sprintf(":%d", config.Configuration.Port),
		Handler:   router,
		TLSConfig: &tls_config,
	}
	if config.Httponly {
		serve.ListenAndServe()
	} else {
		conn, err := net.Listen("tcp", serve.Addr)
		if err != nil {
			panic(err)
		}
		tlsListener := tls.NewListener(conn, &tls_config)
		err = serve.Serve(tlsListener)
		if err != nil {
			glog.Fatalf("Can not Serve TLS, caused by: %s\n", err)
		}
	}
	return nil
}
