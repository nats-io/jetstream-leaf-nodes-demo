# JetStream-LeafNodes-Demo

This repository contains the configuration for the [Persistence at the Edge == JetStream in Leaf Nodes demo](https://www.youtube.com/watch?v=0MkS_S7lyHk).
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
They contain already pushed account JWT so you are ready to go. 
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

To create the contexts used execute the commands below. The context will function in the current directory only.

```
nats context save sys --creds ./keys/creds/OP/SYS/sys.creds   --server "nats://127.0.0.1:4222,nats://127.0.0.1:4232,nats://127.0.0.1:4242,nats://127.0.0.1:4252,nats://127.0.0.1:4262,nats://127.0.0.1:4272,nats://127.0.0.1:4282,nats://127.0.0.1:4292,nats://127.0.0.1:4202" 
nats context save hub     --creds ./keys/creds/OP/TEST/leaf.creds --server "nats://127.0.0.1:4222,nats://127.0.0.1:4232,nats://127.0.0.1:4282" 
nats context save spoke-1 --creds ./keys/creds/OP/TEST/leaf.creds --server "nats://127.0.0.1:4242,nats://127.0.0.1:4252,nats://127.0.0.1:4293"
nats context save spoke-2 --creds ./keys/creds/OP/TEST/leaf.creds --server "nats://127.0.0.1:4262,nats://127.0.0.1:4272,nats://127.0.0.1:4203"
```

