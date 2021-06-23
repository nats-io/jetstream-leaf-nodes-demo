# JetStream-LeafNodes-Demo

This repository contains the configuration for the JPersistence at the Edge == JetStream in Leaf Nodes demo.
The state is identical to the one the demo started with. 

To set up your `nsc` environment execute the followings commands in base directory:

```bash
export NKEYS_PATH="`pwd`/keys"
nsc env -s "`pwd`/store"
nsc env --operator OP
```

The context used in the demo need to be created separately using 

## Content of this repo

This directory contains `nsc` directories `store` and `keys` containing jwt and creds.
Server config files are of the format `cluster-<domain>-<server number>.cfg`
`nats-account-resolver.cfg` contains the account resolver setup shared by all server.
The directories `CACHE*` are nats account resolver directories for each server.
`main.go` contains the source code shown during the presentation.
`outline.txt` contains the outline of the presentation.
The folder `puml` contains the plant uml files used to generate the png named `topology*`
To generate install plantuml and execute `plantuml -tpng <puml file>`.

## Server Startup

To have a single nats account resolver config file each server needs the environment variable `CACHE` set.
This variable is referenced in line four of the config file `nats-account-resolver.cfg`.

To start the server execute the following commands.

```bash
export CACHE='"./cache1"'; nats-server -c cluster-hub-1.cfg
export CACHE='"./cache2"'; nats-server -c cluster-hub-2.cfg
export CACHE='"./cache3"'; nats-server -c cluster-hub-3.cfg

export CACHE='"./cache4"'; nats-server -c cluster-spoke-1-1.cfg
export CACHE='"./cache5"'; nats-server -c cluster-spoke-1-2.cfg
export CACHE='"./cache6"'; nats-server -c cluster-spoke-1-3.cfg

export CACHE='"./cache7"'; nats-server -c cluster-spoke-2-1.cfg
export CACHE='"./cache8"'; nats-server -c cluster-spoke-2-2.cfg
export CACHE='"./cache9"'; nats-server -c cluster-spoke-2-3.cfg
```

## Nats cli contexts

### Context sys

When editing the context, make sure to set credentials to the absolute path of `./keys/creds/OP/SYS/sys.creds` and Server URLs to:
nats://127.0.0.1:4222,nats://127.0.0.1:4232,nats://127.0.0.1:4242,nats://127.0.0.1:4252,nats://127.0.0.1:4262,nats://127.0.0.1:4272,nats://127.0.0.1:4282,nats://127.0.0.1:4292,nats://127.0.0.1:4202

```bash
nats context add sys
mats context edit sys
```

### Context hub

When editing the context, make sure to set credentials to the absolute path of `./keys/creds/OP/TEST/leaf.creds` and Server URLs to:
nats://127.0.0.1:4222,nats://127.0.0.1:4232,nats://127.0.0.1:4282

```bash
nats context add sys
mats context edit sys
```

### Context spoke-1

When editing the context, make sure to set credentials to the absolute path of `./keys/creds/OP/TEST/leaf.cred` and Server URLs to:
nats://127.0.0.1:4242,nats://127.0.0.1:4252,nats://127.0.0.1:4293

```bash
nats context add spoke-1
mats context edit spoke-1
```

### Context spoke-1

When editing the context, make sure to set credentials to the absolute path of `./keys/creds/OP/TEST/leaf.cred` and Server URLs to:
nats://127.0.0.1:4262,nats://127.0.0.1:4272,nats://127.0.0.1:4203

```bash
nats context add spoke-2
mats context edit spoke-2
```

