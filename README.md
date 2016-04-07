# CHIMP
[![Build Status](https://travis-ci.org/zalando-techmonkeys/chimp.svg?branch=master)](https://travis-ci.org/zalando-techmonkeys/chimp) [![Go Report Card](http://goreportcard.com/badge/zalando-techmonkeys/chimp)](http://goreportcard.com/report//zalando-techmonkeys/chimp)

Chimp is a command line interface/server application created at [Zalando Tech](https://tech.zalando.com/), but useful out-of-the-box for anyone who is deploying applications in a Mesos/Marathon environment. This repository currently contains Chimp's server (chimp-server) and CLI (chimp).

###Potential Uses
Here are some ways you can use Chimp:
- To serve as an OAuth2 proxy for Marathon and/or Kubernetes. Neither technology supported OAuth2 when we started this project.
- To have an opinionated, flexible and simple way to deploy applications. 
- To switch easily between a Kubernetes cluster and a Mesos/Marathon cluster without having to change your tooling.
- To function as the front-facing layer for any API during deployment.

Chimp is designed to support replaceable backends and enable users to opt-out of OAuth2, SSL, etc.

### Project Status
Chimp is in active development. You can consider the master branch "stable," but breaking changes are still possible. Breaking changes will be developed in short-lived branches.

The current version doesn't offer functioning support for Kubernetes. Let us know if you'd like to become a contributor and work on this.

## CHIMP Server

### Prerequisites

You'll need to install [Go](https://golang.org/). Any version should work, but we strongly recommend using 1.5.X. If you use a different version, please submit an issue reporting any bugs that you find.

After installing Go, configure it:

````shell
#configure GOPATH
mkdir $HOME/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
````
To make exported variables persistent, add them to your bashrc/zshrc. Otherwise, they'll only be available for your current session. 

Place the source code of this repository in the ```$GOPATH/src folder```. Use these commands:

````shell
mkdir -p $GOPATH/src/github.com/zalando-techmonkeys/
cd $GOPATH/src/github.com/zalando-techmonkeys/
git clone REPO_URL
cd chimp
````

### Installation

For dependency management, install [godep](https://github.com/tools/godep):

````shell
#install godep if you don't have it
go get github.com/tools/godep

#install required dependencies
godep restore

#for tagging the build, both server and cli:

godep go install  -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`"  -tags "zalandoValidation zalando" github.com/zalando-techmonkeys/chimp/...
````

### Configuration

To choose a backend, add a yaml configuration file named [config.yaml](docs/configurations/chimp-server/config.yaml) into: ```/etc/chimp-server/``` or ```$HOME/.config/chimp-server/```. Note that Chimp only supports Marathon for now; see the section "Potential Next Steps" below for more info.

The endpoint of the chosen backend system is also specified in the ```config.yaml``` file. Please refer to the example for an overview of supported options.

### Using Chimp
After you've installed Chimp successfully, you can run the API server as:

````shell
# run service
chimp-server -logtostderr
````

###The Chimp Command Line Interface

Chimp's CLI offers the following operations:
- LOGIN: get a valid OAuth2 token. Optional, if no OAuth2 is required.
- CREATE: deploy an application.
- DELETE: stop a running application.
- INFO: get info for a particular application.
- LIST: list all the apps running on the cluster.
- UPDATE: update an application definition. The app will be restarted.
- SCALE: scale the application to a number of instances.

#### Installing the Chimp CLI from Source

Install [godep](https://github.com/tools/godep) for dependency management:

````shell
#install godep if you don't have it
go get github.com/tools/godep

#install required dependencies
godep restore

#for tagging the build, both server and cli:
godep go install  -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`"  -tags "zalandoValidation zalando" github.com/zalando-techmonkeys/chimp/...
````

#### CLI Configuration
To set up the Chimp CLI, add a yaml configuration file named [config.yaml](docs/configurations/chimp/config.yaml) to ```/etc/chimp/``` or ```$HOME/.config/chimp/```. You'll find an example of such a file in ```docs/configurations/chimp/config.yaml```, where you can set the server and port for Chimp's server. Note that command arguments passed to the Chimp CLI will override the configuration set in the ```config.yaml``` file.

#### Multi-Cluster Support
(Note: This support is currently in a rough state.) The Chimp CLI enables you to set endpoints for multiple clusters. You can also specify which cluster you want to use, via the option ```--cluster=CLUSTERNAME```. 

If you select the cluster "ALL" — or don't specify an option — Chimp will deploy the app on every available cluster. The DeployRequest must be considered "per cluster." This means that, if `3` instances are specified, but only two clusters are currently available, three instances will be deployed on the first and three instances on the second.

####Chimp Commands
**Login**: Use this to obtain a valid token for built-in OAuth2 support. Unnecessary if your server is not configured to use OAuth2.

````shell
#this will ask for your password
chimp login USERNAME
````
**Create**: Please populate the svcport with the port you want to expose for your application and your app's name. For now, names must be exclusive.

````shell
# example params --url=pierone.stups.zalan.do/cat/cat-hello-aws:0.0.1 --name=test
chimp create YOURAPPNAME YOUR_PIERONE_URL --port=YOUR_PORT --cpu=NUM_CORES --memory=MEMORY --replicas=NUM_REPLICAS --http-only
````
**Create** also supports a ```--file``` option that allows you to pass the parameters in a yaml file. An example file:

````yaml
---
DeployRequest:
  - name: demo
    imageURL: YOUR_IMAGE
    replicas: 3
    ports:
      - 8080
    CPULimit: 1
    MemoryLimit: 4000MB
    force: true
    env:
      MYENVVAR: "test"
    volumes:
        - hostPath: /etc/chimp-server/config.yaml
          containerPath: /etc/chimp-server/config.yaml
          mode: "RO"
````

**Delete**
````
chimp delete YOUR_APP_NAME
````

**Info**
````
chimp info YOUR_APP_NAME
````

**List**: Lists all the applications for your team, when OAuth2 is on. Includes a ```--all``` option so you can get the full list of every application.
````
chimp list
````

**Scale**
````
chimp scale YOUR_APP_NAME NUMBER_OF_REPLICAS
````

###Contributing
- Issues: Just post a GitHub issue.
- Enhancements/Bug fixes: Pull requests are welcome.
- Contact: team-techmonkeys@zalando.de.
- See [MAINTAINERS](MAINTAINERS) file.

###Potential Next Steps
We've considered several possibilities for the Chimp project:
- Evolve it into a package manager that abstracts away Kubernetes or Marathon — making it possible to deploy the same packages in the same way across multiple, different clusters.
- Redesign it as a client-only app. This is in response to the direction Marathon has recently taken. OAuth2 would remain, but the new plugin architecture would allow users to write their own. This would simplify usage by getting rid of the server.

Share your thoughts on these (or any other options) by submitting an issue.

### License

See [LICENSE](LICENSE) file.
