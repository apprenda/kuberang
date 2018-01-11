# Kuberang
[![CircleCI](https://circleci.com/gh/apprenda/kuberang.svg?style=svg)](https://circleci.com/gh/apprenda/kuberang)

`kuberang` is a command-line utility for smoke testing a Kubernetes install.

It scales out a pair of services, checks their connectivity, and then scales back in again. Sort of like a boomerang, you know?

[Latest Build for Linux x86](https://kismatic-installer.s3-accelerate.amazonaws.com/kuberang/latest/kuberang-linux-amd64)

[Latest Build for Darwin x86](https://kismatic-installer.s3-accelerate.amazonaws.com/kuberang/latest/kuberang-darwin-amd64)

It will tell you if the machine and account from which you run it:
* Has kubectl installed correctly with access controls
* Has active kubernetes namespace (if specified)
* Has available workers
* Has working pod & service networks
* Has working pod <-> pod DNS
* Has working master(s)
* Has the ability to access pods and services from the node you run it on.

It's suggested that you run `kuberang` from a node OTHER than a worker.

`kuberang` will exit with a code of 0 even if some of the above are not possible. It's up to the user to parse the output.

Adding -o json will return a parsable json blob instead of a pretty string report.

### Pre-requisites
* A working kubectl (or all you'll get is a message complaining about kubectl)
* Access to a Docker registry with 
  busybox and nginx images

### Usage

```
$ kuberang --help
kuberang tests your kubernetes cluster using kubectl

Usage:
  kuberang [flags]
  kuberang [command]

Available Commands:
  version     display the Kismatic CLI version

Flags:
  -n, --namespace string      Kubernetes namespace in which kuberang will operate. Defaults to 'default' if not specified.
      --registry-url string   Override the default Docker Hub URL to use a local offline registry for required Docker images.
      --skip-cleanup          Don't clean up. Leave all deployed artifacts running on the cluster.

Use "kuberang [command] --help" for more information about a command.
```

# Developer notes
### Pre-requisites
- Go 1.7 installed

### Build using make
We use `make` to clean, build, and produce our distribution package. Take a look at the Makefile for more details.

In order to build the Go binaries (e.g. Kismatic CLI):
```
make build
```

In order to clean:
```
make clean
```

In order to produce a docker image:
```
make docker
```

In order to produce the distribution package:
```
make dist
```
This will produce an `./out` directory, which contains the bits, and a tarball.

You may pass build options as necessary:
```
GOOS=linux make build
```
