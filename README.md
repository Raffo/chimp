# CHIMP
[![Build Status](https://travis-ci.org/zalando-techmonkeys/chimp.svg?branch=master)](https://travis-ci.org/zalando-techmonkeys/chimp)

Chimp is the deployment tool for our cloud solution. It is designed to be the front facing layer for any API for deployment
This project currently contains chimp's server (chimp-server) and our CLI (chimp).

## Project Status
The project is still in active development and we could introduce breaking changes. The master branch is still to be considered as "stable" and most breaking changes will be developed in short lived branches.
NOTE: In the current version, the kubernetes support is not working.

## CHIMP Server

### Prerequisites

To build the source you need to have [Go](https://golang.org/) installed. You
can install it in any way you like, we strongly suggest you to use version 1.5.X as this is our current target. Feel free to use any other version and report any possible bug.

You then need to configure Go:

````shell
#configure GOPATH
mkdir $HOME/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
````
Please remember that the exported variables will only be available for your
current session. Add them to your bashrc/zshrc if you want to make them
persistent.

The source code of this repository must be put into the ```$GOPATH/src folder```.
This can be done with the following commands:

````shell
mkdir -p $GOPATH/src/github.com/zalando-techmonkeys/
cd $GOPATH/src/github.com/zalando-techmonkeys/
git clone REPO_URL
cd chimp
````

### Install

Install [godep](https://github.com/tools/godep) for dependency management.

````shell
#install godep if you don't have it
go get github.com/tools/godep

#install required dependencies
godep restore

#for tagging the build, both server and cli:
godep go install  -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`"  github.com/zalando-techmonkeys/chimp/...
````

### Configuration
To choose a backend, put a yaml configuration
file named [config.yaml](chimp/docs/configurations/chimp-server/config.yaml) into: ```/etc/chimp-server/``` or ```$HOME/.config/chimp-server/```.
The endpoint of the chosen backend system is also specified in the ```config.yaml``` file. Please refer to the example for an overview of supported options.

### Usage
Once chimp is installed, the API server can be simply run as:

````shell
# run service
chimp-server -logtostderr
````

## CHIMP CLI

The cli allows you to do four operations:

- LOGIN: used to get a valid OAuth2 token, optional if no OAuth is required .
- CREATE: used to deploy an application.
- DELETE: used to stop a running application.
- INFO: used to get info for a particular application.
- LIST: to list all the apps running on the cluster.
- UPDATE: to update an application definition. The app will be restarted.
- SCALE: to scale the application to a number of instances.

## Install
#### From Source

Install [godep](https://github.com/tools/godep) for dependency management.

````shell
#install godep if you don't have it
go get github.com/tools/godep

#install required dependencies
godep restore

#for tagging the build, both server and cli:
godep go install  -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" github.com/zalando-techmonkeys/chimp/...
````

### Configuration
To setup the chimp cli put a yaml configuration file named [config.yaml](chimp/docs/configurations/chimp/config.yaml) into: ```/etc/chimp/``` or ```$HOME/.config/chimp/``` In ```docs/configurations/chimp/config.yaml``` you  find an example of such a file where you can set server and port for chimp's server.
Please note that command arguments passed to chimp command line will override the configuration set in the ```config.yaml``` file.


### Multi cluster support
A very rough multi-cluster support has been implemented.
You can set in chimp (cli) the endpoints of multiple clusters. This allows the user to specify which cluster wants to use with the option ```--cluster=CLUSTERNAME```. If the cluster "ALL" is selected or no option is specified, chimp will deploy the app on every available cluster as the DeployRequest has to be considered "per cluster". This means that if `3` instances are specified and two clusters are currently available, 3 instances will be deployed on the first and 3 instances on the second.

## Commands

### Login
You can use chimp login to obtain a valid token for the built-in OAuth2 support. Please note that this is not required if
the server is not configured to use OAuth2.
````shell
#this will ask for your password
chimp login USERNAME
````

### Create
Please populate the svcport with the port you want to expose for your application and the name for the name of your app. Names must be exclusive for now.

````shell
# example params --url=pierone.stups.zalan.do/cat/cat-hello-aws:0.0.1 --name=test
chimp create YOURAPPNAME YOUR_PIERONE_URL --port=YOUR_PORT --cpu=NUM_CORES --memory=MEMORY --replicas=NUM_REPLICAS --http-only 
````

Create also supports a ```--file``` option that allow you to pass the parameters in a yaml file.
The following is an example file:

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

### Delete
````
chimp delete YOUR_APP_NAME
````

### Info
````
chimp info YOUR_APP_NAME
````

### List
Lists all the applications for your team (when OAuth2 is on). It has a ```--all``` option that can be used to get the full list of every application.
````
chimp list
````

### Scale
````
chimp scale YOUR_APP_NAME NUMBER_OF_REPLICAS
````

## Development
* Issues: Just create issues on github
* Enhancements/Bugfixes: Pull requests are welcome
* get in contact: team-techmonkeys@zalando.de
* see [MAINTAINERS](MAINTAINERS)
file.

## License

See [LICENSE](LICENSE) file.
