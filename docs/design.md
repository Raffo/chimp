# CHIMP

Chimp was designed to support many cloud APIs as backends from the beginning, initially starting with Marathon and Kubernetes in mind. The goal was to have a common API to deploy applications while having to deal with different cloud solutions.
The API was designed to be a common API on top of heterogenous APIs and it somehow reminds Kubernetes' API.

## Assumptions
Chimp has been built to be as modular as possible and there are not a lot of assumptions made. Most of the things you will find are optonal and configurable, both on server side (chimp-server) and client side (chimp).
Please have a look at the configurations in [config.yaml](chimp/docs/configurations/chimp-server/config.yaml) and [config.yaml](chimp/docs/configurations/chimp/config.yaml) which can be used as examples on how to tweak your current setup.
We, at [Zalando SE](https://tech.zalando.com/),  have to support multiple heterogeneous and autonomous teams. This means you will find the concept of "team" in some part of the chimp. This should be in no way a problem for you to use chimp as everything was designed to be pluggable, but feel free to open an issue otherwise.

## Adding a new backend
To add a new Backend to the existing ones, it must implement the interface in the package backend. You can have a look at how the Marathon backend has been implemented as a reference.

## Adding a new Validator
Validators are introduced to force some constraints on the user requests. Using build tags you can specify at compile time which validators you want to use. Not specifying any build tags will remove the validation.
By design, you can add as many validators as you want, but only one can be used at the same time and has to be specified at build time.
