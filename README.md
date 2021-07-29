# JetStream-LeafNodes-Demo

This repository contains the configuration for the [Persistence at the Edge == JetStream in Leaf Nodes demo](https://www.youtube.com/watch?v=0MkS_S7lyHk) as well as the [script](#video-script-and-commands) it is based on.
The state is identical to the one the demo started with. 

To set up your `nsc` environment execute the following commands in the base directory:

```txt
export NKEYS_PATH="`pwd`/keys"
nsc env -s "`pwd`/store"
nsc env --operator OP
```

The context used in the demo needs to be created separately using the following;

## Content of This Repo

This directory contains `nsc` directories `store` and `keys` containing JWT and creds.
Server config files are of the format `cluster-<domain>-<server number>.cfg`
`nats-account-resolver.cfg` contains the account resolver setup shared by all servers.
The directories `CACHE*` are NATS account resolver directories for each server.
They contain already pushed account JWTs so you are ready to go. 
`main.go` contains the source code shown during the YouTube demo.
`outline.txt` contains the outline of the YouTube demo.
The folder `puml` contains the plant uml files used to generate the png named `topology*`.
To generate, install [plantuml](https://plantuml.com/) and execute `plantuml -tpng <puml file>`.

## Server Startup

To have a single NATS account resolver config file, each server needs the environment variable `CACHE` set.
This variable is referenced in line four of the config file `nats-account-resolver.cfg`.

To start the server execute the following commands:

```txt
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

Or all at once:

```txt
i=0; for c in cluster*.cfg; do ((i=i+1)); export CACHE=cache$i; nats-server -c $c  & ; done
```

## NATS CLI Contexts

To create the contexts used, execute the commands below. The context will function in the current directory only.

```
nats context save sys --creds ./keys/creds/OP/SYS/sys.creds   --server "nats://127.0.0.1:4222,nats://127.0.0.1:4232,nats://127.0.0.1:4242,nats://127.0.0.1:4252,nats://127.0.0.1:4262,nats://127.0.0.1:4272,nats://127.0.0.1:4282,nats://127.0.0.1:4292,nats://127.0.0.1:4202" 
nats context save hub     --creds ./keys/creds/OP/TEST/leaf.creds --server "nats://127.0.0.1:4222,nats://127.0.0.1:4232,nats://127.0.0.1:4282" 
nats context save spoke-1 --creds ./keys/creds/OP/TEST/leaf.creds --server "nats://127.0.0.1:4242,nats://127.0.0.1:4252,nats://127.0.0.1:4292"
nats context save spoke-2 --creds ./keys/creds/OP/TEST/leaf.creds --server "nats://127.0.0.1:4262,nats://127.0.0.1:4272,nats://127.0.0.1:4202"
```

## Video Script and Commands

NATS is the networking abstraction to finally free you and your apps from networking and the silos that plague us all.
It allows you to refocus on your data and it’s flows on a global scale.
We have recently added our newest persistence layer, JetStream in NATS Server v2.2.

This video is about JetStream at the edge.

The NATS server is lightweight and can therefore run in resource constrained environments.
This allows you to have persistence on a remote edge and have it function independently, without connectivity to the cloud.
If you want to, you can configure JetStream such that your locally persisted data is automatically uploaded to the cloud as soon as you regain connectivity.

This applies specifically to cases where the edge site itself moves in and out of network connectivity, 
or when you have a great number of sites and are guaranteed a network outage at any given time.

Depending on your needs you would install one or more JetStream enabled servers in your edge site and connect them to a server in the cloud using leaf node connections.
To demonstrate the principle, I am working with two clusters that are connected to a hub. But you can have as many as needed.

This is going to be a whirlwind tour through features and use cases.
Do not focus on every detail. Please focus on what our software can do and which use cases are relevant to you.
I will provide [additional links](#relevant-links) at the end. 

### Outline

In this video I want to:

1. [Introduce the leafnode setup used here](#setup-topology)
2. [Talk about the implications of connecting the system account as leaf node remote](#connected-system-account-implications)
3. [Introduce the concept of JetStream domains](#jetstream-domains)
4. [Use stream mirrors to connect a command and control stream across domains](#stream-mirrors-across-domains)
5. [Use stream source to aggregate streams across domains](#stream-source-across-domains)
6. [Demonstrate that domain connectivity is not tied to the underlying topology](#sourcemirror-not-dependent-on-topology)
7. [Connect streams across accounts](#connect-streams-cross-accounts) 
8. [Connect streams through accounts](#connect-streams-through-accounts)
9. [Relevant Documentation](#relevant-links)

#### Setup Topology 

This is the topology of my setup.

![`imgcat topology1-server.png`](topology1-server.png)

There is a central cluster to which two clusters `spoke-1` and `spoke-2` connect via leafnode connections.
I am using two of these to demonstrate the principle, but you can have as many as needed.
To show it's possible, each cluster consists of three servers but use fewer if you do not need the redundancy. 
Let's have a look at server one in the cluster hub.

```txt
> cat cluster-hub-1.cfg
listen: localhost:4222
server_name: srv-4222
jetstream {
    store_dir: "./s1-1"
    domain: hub
}
cluster {
    listen localhost:4223
    name cluster-hub
    routes = [
        nats-route://localhost:4223
        nats-route://localhost:4233
        nats-route://localhost:4283
    ]
}
leafnodes {
    listen localhost:4224
    no_advertise: true
}
mqtt {
    port: 4225
}
http: localhost:8080
include ./nats-account-resolver.cfg
```

This is your regular cluster and leafnode setup.
JetStream has a new property called `domain` which I'll be talking about in a moment.

```txt
> cat nats-account-resolver.cfg
operator: "./store/OP/OP.jwt"
resolver: {
    type: full
    dir: $CACHE
    interval: "2m"
    allow_delete: true
}
resolver_preload: {
	ADECCNBUEBWZ727OMBFSN7OMK2FPYRM52TJS25TFQWYS76NPOJBN3KU4:eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJMTFpUU0paUTNGTkZWNlhDVDNBRkdCSlZWU0FaSk9IN0JSNlFETUdNVUdETktVWUlQR0hRIiwiaWF0IjoxNjI0NDc5NDgyLCJpc3MiOiJPRE5FQjdESUtMNlQ0UTYyTVNFUjJEMkhDQ05OSTU1WkZMUEpVNkM0NEFRVEQ1T0lPUEhUTEQ1USIsIm5hbWUiOiJTWVMiLCJzdWIiOiJBREVDQ05CVUVCV1o3MjdPTUJGU043T01LMkZQWVJNNTJUSlMyNVRGUVdZUzc2TlBPSkJOM0tVNCIsIm5hdHMiOnsibGltaXRzIjp7InN1YnMiOi0xLCJkYXRhIjotMSwicGF5bG9hZCI6LTEsImltcG9ydHMiOi0xLCJleHBvcnRzIjotMSwid2lsZGNhcmRzIjp0cnVlLCJjb25uIjotMSwibGVhZiI6LTF9LCJzaWduaW5nX2tleXMiOlsiQUFDWUlDT0FRTVE3MkVIVDM1UjdMVjZWRldNSVZXRktXRkU1UDJKSjJUVDY3NEVPN0RKVFVITU0iXSwiZGVmYXVsdF9wZXJtaXNzaW9ucyI6eyJwdWIiOnt9LCJzdWIiOnt9fSwidHlwZSI6ImFjY291bnQiLCJ2ZXJzaW9uIjoyfX0.tHteTcshVIInToM6LQ7G2AmmWfeKYCCjJdCC4ZfJ1WUtmY1Bk0sEwwbjb6uSycEb4ljohMQnQVgkbAYZsiqZDw
}
```

Enabling multi tenancy, accounts are the secure NATS isolation context. 
Unless explicitly defined, connections belonging to one account can not communicate with connections belonging to another.
Because of this we generally recommend one account per application. However, their scope is up to you.

In addition, this is an operator setup and uses the NATS account resolver to retrieve accounts.
Instead of defining an account in every server we provide a way to obtain it.
On connect, the client provides a web token carrying user permissions as well as proof of possession of a corresponding private key.
After the account web token is downloaded, the chain of trust, operator signs account token, account signs user token is verified and all relevant limits and settings are applied.

What I will be demonstrating in this video also applies when you configure your accounts in server config files.

```txt
> cat cluster-spoke-1-1.cfg
listen: localhost:4242
server_name: srv-4242
jetstream {
    store_dir: "./s2-1"
    domain: spoke-1
}
cluster {
    listen localhost:4243
    name cluster-spoke-1
    routes = [
        nats-route://localhost:4243
        nats-route://localhost:4253
        nats-route://localhost:4293
    ]
}
mqtt {
    port: 4245
}
http: localhost:8084
include ./nats-account-resolver.cfg
include ./leafnode-remotes.cfg
```

Leaf nodes are clustered as well. 

#### Connected System Account Implications

Although leaf nodes bridge authentication domains, if you have the same operator setup on either end and connect system accounts, it will appear to you as one large authentication domain.
You could use such a setup, for example, when you can't use a super cluster due to firewall rules and have to use leaf node connections instead.

Another benefit of connecting system accounts is that you can obtain monitoring information from every server in this network.

```
> nats --context=sys server list 9
╭─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                           Server Overview                                                           │
├──────────┬─────────────────┬───────────┬────────────┬─────┬───────┬──────┬────────┬─────┬─────────┬─────┬──────┬────────┬───────────┤
│ Name     │ Cluster         │ IP        │ Version    │ JS  │ Conns │ Subs │ Routes │ GWs │ Mem     │ CPU │ Slow │ Uptime │ RTT       │
├──────────┼─────────────────┼───────────┼────────────┼─────┼───────┼──────┼────────┼─────┼─────────┼─────┼──────┼────────┼───────────┤
│ srv-4242 │ cluster-spoke-1 │ localhost │ 2.3.3-beta │ yes │ 0     │ 283  │ 2      │ 0   │ 16 MiB  │ 0.4 │ 0    │ 39.65s │ 975.35µs  │
│ srv-4262 │ cluster-spoke-2 │ localhost │ 2.3.3-beta │ yes │ 0     │ 283  │ 2      │ 0   │ 17 MiB  │ 0.3 │ 0    │ 39.65s │ 953.193µs │
│ srv-4202 │ cluster-spoke-2 │ localhost │ 2.3.3-beta │ yes │ 0     │ 283  │ 2      │ 0   │ 16 MiB  │ 0.3 │ 0    │ 39.65s │ 933.849µs │
│ srv-4232 │ cluster-hub     │ localhost │ 2.3.3-beta │ yes │ 0     │ 338  │ 2      │ 0   │ 18 MiB  │ 0.3 │ 0    │ 39.65s │ 913.57µs  │
│ srv-4272 │ cluster-spoke-2 │ localhost │ 2.3.3-beta │ yes │ 0     │ 297  │ 2      │ 0   │ 16 MiB  │ 0.4 │ 0    │ 39.65s │ 890.216µs │
│ srv-4292 │ cluster-spoke-1 │ localhost │ 2.3.3-beta │ yes │ 0     │ 283  │ 2      │ 0   │ 16 MiB  │ 0.3 │ 0    │ 39.65s │ 865.885µs │
│ srv-4252 │ cluster-spoke-1 │ localhost │ 2.3.3-beta │ yes │ 0     │ 311  │ 2      │ 0   │ 16 MiB  │ 0.2 │ 0    │ 39.65s │ 844.531µs │
│ srv-4222 │ cluster-hub     │ localhost │ 2.3.3-beta │ yes │ 0     │ 306  │ 2      │ 0   │ 17 MiB  │ 0.3 │ 0    │ 39.65s │ 819.322µs │
│ srv-4282 │ cluster-hub     │ localhost │ 2.3.3-beta │ yes │ 1     │ 279  │ 2      │ 0   │ 16 MiB  │ 0.5 │ 0    │ 39.65s │ 791.504µs │
├──────────┼─────────────────┼───────────┼────────────┼─────┼───────┼──────┼────────┼─────┼─────────┼─────┼──────┼────────┼───────────┤
│          │ 3 Clusters      │ 9 Servers │            │ 9   │ 1     │ 2663 │        │     │ 149 MiB │     │ 0    │        │           │
╰──────────┴─────────────────┴───────────┴────────────┴─────┴───────┴──────┴────────┴─────┴─────────┴─────┴──────┴────────┴───────────╯

╭────────────────────────────────────────────────────────────────────────────────────╮
│                                  Cluster Overview                                  │
├─────────────────┬────────────┬───────────────────┬───────────────────┬─────────────┤
│ Cluster         │ Node Count │ Outgoing Gateways │ Incoming Gateways │ Connections │
├─────────────────┼────────────┼───────────────────┼───────────────────┼─────────────┤
│ cluster-spoke-1 │ 3          │ 0                 │ 0                 │ 0           │
│ cluster-spoke-2 │ 3          │ 0                 │ 0                 │ 0           │
│ cluster-hub     │ 3          │ 0                 │ 0                 │ 1           │
├─────────────────┼────────────┼───────────────────┼───────────────────┼─────────────┤
│                 │ 9          │ 0                 │ 0                 │ 1           │
╰─────────────────┴────────────┴───────────────────┴───────────────────┴─────────────╯
```

Here you can see all of the leaf nodes as well.
Please note however, the decision to connect system accounts heavily depends on your security needs!

```txt
> cat leafnode-remotes.cfg
HUB-URLS=["nats-leaf://localhost:4224","nats-leaf://localhost:4234","nats-leaf://localhost:4284"]
leafnodes {
no_advertise: true
    remotes = [
		{
			urls: $HUB-URLS
			account: ADECCNBUEBWZ727OMBFSN7OMK2FPYRM52TJS25TFQWYS76NPOJBN3KU4
 			credentials: keys/creds/OP/SYS/sys.creds
		},
		{
			urls: $HUB-URLS
			account: AA5C56FAETBTUCYM7NC5BFBYFTKLOABIOIFPQDHO4RUEAPSN3FTY5R4G
			credentials: keys/creds/OP/TEST/leaf.creds
		},
	]
}
```

This is exactly what was done here. 
The first remote connects the system accounts, he second remote connects the account we will be using. If you have the system account connected but no domain specified, this is the JetStream topology you'd get.

![`imgcat topology2-js-merged.png`](topology2-js-merged.png)

This is a singe JetStream spanning all of our clusters.
JetStream has a meta data leader that is in charge of creating and placing streams and consumers.
In a hub/spoke setup like the one here, this meta data leader will be pinned to server that has no leaf node remotes specified (meaning, the hub).
This avoids having a leaf node server become a leader and thus forcing other leaf nodes to take two hops to get to it.
But, while leaf node connections are down, in the affected server, you could not create streams or consumers during that time.

#### JetStream Domains

Unlike above with a single JetStream, if you specify a JetStream `domain` in your configs, each `domain` will become an independent JetStream.
My config will result in this topology.

![`imgcat topology3-js-domains.png`](topology3-js-domains.png)

If you do not connect system accounts, this will be the resulting topology as well.
Presence of domains cuts off certain system account traffic along leaf node connections and makes JetStream addressable by domain name.

The part I want to show you now is how to connect these independent JetStreams.

On the left I am starting a watch command that repeatedly executes three `nats` commands.
This CLI command is our swiss army knife. Among other things it can be used to publish and subscribe, interact with JetStream and generate various reports.
It supports a context that can be set once and subsequently reused.
Here, each invocation of NATS uses a context that uses appropriate credentials and connects it to the corresponding cluster.

```txt
> watch -n 1 "nats --context=hub s report ; nats --context=spoke-1 s report; nats --context=spoke-2 s report"

Obtaining Stream stats

No Streams defined
Obtaining Stream stats

No Streams defined
Obtaining Stream stats

No Streams defined
```

This will be my primary tool to show you what happened.
Right now, there are no streams yet.

#### Stream Mirrors Across Domains

One way to connect streams across JetStream domains would be to have a command and control stream. 
Essentially to communicate from the hub to each spoke without loss, even when leaf nodes are disconnected.

![`imgcat topology4-js-streams-mirror.png`](topology4-js-streams-mirror.png)

We have a stream called `cnc` and it is subscribing to a subject by the same name.
This stream is located in the hub. 
It's content is then mirrored to a `recv-cnc` stream located in each node.

```txt
> nats --context=hub stream add cnc --subjects cnc --replicas 3
? Storage backend file
? Retention Policy Limits
? Discard Policy Old
? Stream Messages Limit -1
? Per Subject Messages Limit -1
? Message size limit -1
? Maximum message age limit -1
? Maximum individual message size -1
? Duplicate tracking time window 2m
Stream cnc was created

Information for Stream cnc created 2021-07-28T12:25:07-04:00

Configuration:

             Subjects: cnc
     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited


Cluster Information:

                 Name: cluster-hub
               Leader: srv-4222
              Replica: srv-4232, current, seen 0.00s ago
              Replica: srv-4282, current, seen 0.00s ago

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
```

Streams offer a variety of settings like storage backend, various limits etc..
I will only give arguments for values I want to be changed; shown here, stream replica count and subject to subscribe to. 
This way I can quickly go through the questionnaire.  

Let's also store a few messages:

```txt
> nats --context=hub pub cnc "hello world" --count 10
12:25:50 Published 11 bytes to "cnc"
12:25:50 Published 11 bytes to "cnc"
12:25:50 Published 11 bytes to "cnc"
12:25:50 Published 11 bytes to "cnc"
12:25:50 Published 11 bytes to "cnc"
12:25:50 Published 11 bytes to "cnc"
12:25:50 Published 11 bytes to "cnc"
12:25:50 Published 11 bytes to "cnc"
12:25:50 Published 11 bytes to "cnc"
12:25:50 Published 11 bytes to "cnc"
```

In each leaf node cluster I want a stream named `recv-cnc` that is a mirror of the `cnc` stream.
The `output` option causes `nats` to store the settings in a file named `recv-cnc`.

```txt
> nats --context=hub stream add recv-cnc --mirror cnc --replicas 3 --output recv-cnc
? Storage backend file
? Retention Policy Limits
? Discard Policy Old
? Stream Messages Limit -1
? Message size limit -1
? Maximum message age limit -1
? Maximum individual message size -1
? Adjust mirror start No
? Import mirror from a different JetStream domain Yes
? Foreign JetStream domain name hub
? Delivery prefix
```

Now I'm creating that stream in each domain.

```txt
> nats --context=hub stream add --config recv-cnc --js-domain spoke-1
Stream recv-cnc was created

Information for Stream recv-cnc created 2021-07-28T12:30:51-04:00

Configuration:

     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited
               Mirror: cnc, API Prefix: $JS.hub.API, Delivery Prefix:


Cluster Information:

                 Name: cluster-spoke-1
               Leader: srv-4252
              Replica: srv-4292, current, seen 0.00s ago
              Replica: srv-4242, current, seen 0.00s ago

Mirror Information:

          Stream Name: cnc
                  Lag: 0
            Last Seen: never
      Ext. API Prefix: $JS.hub.API

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
> nats --context=hub stream add --config recv-cnc --js-domain spoke-2
Stream recv-cnc was created

Information for Stream recv-cnc created 2021-07-28T12:30:58-04:00

Configuration:

     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited
               Mirror: cnc, API Prefix: $JS.hub.API, Delivery Prefix:


Cluster Information:

                 Name: cluster-spoke-2
               Leader: srv-4272
              Replica: srv-4202, current, seen 0.00s ago
              Replica: srv-4262, current, seen 0.00s ago

Mirror Information:

          Stream Name: cnc
                  Lag: 0
            Last Seen: never
      Ext. API Prefix: $JS.hub.API

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
```

It is important to point out that names only need to be unique within a JetStream domain.
On the left hand side you can see the streams created, and that our `recv-cnc` streams already copied all data from `cnc`.

```txt
> watch -n 1 "nats --context=hub s report ; nats --context=spoke-1 s report; nats --context=spoke-2 s report"

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ cnc    │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4222*, srv-4232, srv-4282 │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

Obtaining Stream stats

╭────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                           Stream Report                                            │
├──────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
╰──────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 1.69s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                           Stream Report                                            │
├──────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
╰──────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 0.53s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯
```

#### Stream Source Across Domains

Doing this the other way around, collecting data in the spokes and aggregating them in a stream in the hub is possible as well.

![`imgcat topology5-js-streams-source.png`](topology5-js-streams-source.png)

In each spoke JetStream domain I am going to create a stream named `test` that subscribes to `test`.

```txt
> nats --context=hub stream add test --subjects test --replicas 3 --output test
? Storage backend file
? Retention Policy Limits
? Discard Policy Old
? Stream Messages Limit -1
? Per Subject Messages Limit -1
? Message size limit -1
? Maximum message age limit -1
? Maximum individual message size -1
? Duplicate tracking time window 2m
```

Again, I stored the output, so I can now easily create the stream multiple times.

```txt
> nats --context=hub stream add --config test --js-domain spoke-1
Stream test was created

Information for Stream test created 2021-07-28T12:42:12-04:00

Configuration:

             Subjects: test
     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited


Cluster Information:

                 Name: cluster-spoke-1
               Leader: srv-4252
              Replica: srv-4242, current, seen 0.00s ago
              Replica: srv-4292, current, seen 0.00s ago

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
> nats --context=hub stream add --config test --js-domain spoke-2
Stream test was created

Information for Stream test created 2021-07-28T12:42:14-04:00

Configuration:

             Subjects: test
     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited


Cluster Information:

                 Name: cluster-spoke-2
               Leader: srv-4262
              Replica: srv-4272, current, seen 0.00s ago
              Replica: srv-4202, current, seen 0.00s ago

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
```

Please note that I created the streams in the respective domain while having been connected to the hub. 
If you want to interact with a JetStream that is not in the domain to which you are connected, you can provide the `js-domain` option.
This can also be set in the context.

Let's also send 10 messages.

```txt
> nats --context=hub pub test "hello world" --count 10
12:43:45 Published 11 bytes to "test"
12:43:45 Published 11 bytes to "test"
12:43:45 Published 11 bytes to "test"
12:43:45 Published 11 bytes to "test"
12:43:45 Published 11 bytes to "test"
12:43:45 Published 11 bytes to "test"
12:43:45 Published 11 bytes to "test"
12:43:45 Published 11 bytes to "test"
12:43:45 Published 11 bytes to "test"
12:43:45 Published 11 bytes to "test"
```

As you can see, the messages sent while connected to hub appear in both streams.
This is because they use the same subject. 
JetStream domains do not constrict message flow.

```txt
> watch -n 1 "nats --context=hub s report ; nats --context=spoke-1 s report; nats --context=spoke-2 s report"

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ cnc    │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4222*, srv-4232, srv-4282 │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

Obtaining Stream stats

╭────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                           Stream Report                                            │
├──────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
│ test     │ File    │ 0         │ 10       │ 450 B │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
╰──────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 1.35s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                           Stream Report                                            │
├──────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
│ test     │ File    │ 0         │ 10       │ 450 B │ 0    │ 0       │ srv-4202, srv-4262*, srv-4272 │
╰──────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 0.24s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯
```

If you are using a non-overlapping set of subject names in each domain, this won't happen to you.
Same if you are using a different set of accounts in each domain.
Later I will show a third way of doing this.

```txt
> nats --context=spoke-1 pub test "hello world" --count 10
12:46:27 Published 11 bytes to "test"
12:46:27 Published 11 bytes to "test"
12:46:27 Published 11 bytes to "test"
12:46:27 Published 11 bytes to "test"
12:46:27 Published 11 bytes to "test"
12:46:27 Published 11 bytes to "test"
12:46:27 Published 11 bytes to "test"
12:46:27 Published 11 bytes to "test"
12:46:27 Published 11 bytes to "test"
12:46:27 Published 11 bytes to "test"
```

Of course this also applies when you connect to a spoke as well.

```txt
> watch -n 1 "nats --context=hub s report ; nats --context=spoke-1 s report; nats --context=spoke-2 s report"

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ cnc    │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4222*, srv-4232, srv-4282 │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

Obtaining Stream stats

╭────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                           Stream Report                                            │
├──────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
│ test     │ File    │ 0         │ 20       │ 900 B │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
╰──────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 0.68s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                           Stream Report                                            │
├──────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
│ test     │ File    │ 0         │ 20       │ 900 B │ 0    │ 0       │ srv-4202, srv-4262*, srv-4272 │
╰──────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 1.60s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯
```

Now let's create a stream named `aggregate`, sourcing these streams.
Here I have to specify which domain to source from (`spoke-1`/`spoke-2`).

```txt
> nats --context=hub stream add aggregate --replicas 3 --source test --source test
? Storage backend file
? Retention Policy Limits
? Discard Policy Old
? Stream Messages Limit -1
? Message size limit -1
? Maximum message age limit -1
? Maximum individual message size -1
? Duplicate tracking time window 2m
? Adjust source "test" start No
? Import "test" from a different JetStream domain Yes
? test Source foreign JetStream domain name spoke-1
? test Source foreign JetStream domain delivery prefix
? Adjust source "test" start No
? Import "test" from a different JetStream domain Yes
? test Source foreign JetStream domain name spoke-2
? test Source foreign JetStream domain delivery prefix
Stream aggregate was created

Information for Stream aggregate created 2021-07-28T12:48:52-04:00

Configuration:

     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited
              Sources: test, API Prefix: $JS.spoke-1.API, Delivery Prefix:
                       test, API Prefix: $JS.spoke-2.API, Delivery Prefix:


Cluster Information:

                 Name: cluster-hub
               Leader: srv-4282
              Replica: srv-4232, current, seen 0.00s ago
              Replica: srv-4222, current, seen 0.00s ago

Source Information:

          Stream Name: test
                  Lag: 0
            Last Seen: 0.00s
      Ext. API Prefix: $JS.spoke-1.API

          Stream Name: test
                  Lag: 0
            Last Seen: 0.00s
      Ext. API Prefix: $JS.spoke-2.API

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
```

Here are the corresponding stream reports.

```txt
> watch -n 1 "nats --context=hub s report ; nats --context=spoke-1 s report; nats --context=spoke-2 s report"

Obtaining Stream stats

╭───────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                             Stream Report                                             │
├───────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream    │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├───────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ cnc       │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4222*, srv-4232, srv-4282 │
│ aggregate │ File    │ 0         │ 40       │ 3.8 KiB │ 0    │ 0       │ srv-4222, srv-4232, srv-4282* │
╰───────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭─────────────────────────────────────────────────────────────────────────────╮
│                             Replication Report                              │
├───────────┬────────┬─────────────────┬───────────────┬────────┬─────┬───────┤
│ Stream    │ Kind   │ API Prefix      │ Source Stream │ Active │ Lag │ Error │
├───────────┼────────┼─────────────────┼───────────────┼────────┼─────┼───────┤
│ aggregate │ Source │ $JS.spoke-1.API │ test          │ 0.19s  │ 0   │       │
│ aggregate │ Source │ $JS.spoke-2.API │ test          │ 0.18s  │ 0   │       │
╰───────────┴────────┴─────────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                           Stream Report                                            │
├──────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
│ test     │ File    │ 0         │ 20       │ 900 B │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
╰──────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 1.93s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                           Stream Report                                            │
├──────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
│ test     │ File    │ 0         │ 20       │ 900 B │ 0    │ 0       │ srv-4202, srv-4262*, srv-4272 │
╰──────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 0.89s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯
```

#### Source/Mirror Not Dependent on Topology

Let me also demonstrate that source and mirror stream relationships do not have to align with the underlying topology.

![`imgcat topology6-js-streams-backup.png`](topology6-js-streams-backup.png)

Now I'm going to create a backup of the stream test in `spoke-1`. 
The backup itself is located in `spoke-2`.

```
> nats --context=hub stream add backup-test-spoke-1 --replicas 3 --mirror test --js-domain spoke-2
? Storage backend file
? Retention Policy Limits
? Discard Policy Old
? Stream Messages Limit -1
? Message size limit -1
? Maximum message age limit -1
? Maximum individual message size -1
? Adjust mirror start No
? Import mirror from a different JetStream domain Yes
? Foreign JetStream domain name spoke-1
? Delivery prefix
Stream backup-test-spoke-1 was created

Information for Stream backup-test-spoke-1 created 2021-07-28T12:59:07-04:00

Configuration:

     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited
               Mirror: test, API Prefix: $JS.spoke-1.API, Delivery Prefix:


Cluster Information:

                 Name: cluster-spoke-2
               Leader: srv-4202
              Replica: srv-4262, current, seen 0.00s ago
              Replica: srv-4272, current, seen 0.00s ago

Mirror Information:

          Stream Name: test
                  Lag: 0
            Last Seen: never
      Ext. API Prefix: $JS.spoke-1.API

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
```

And send 10 messages. 

```txt
> nats --context=spoke-1 pub test "hello world" --count 10
13:00:48 Published 11 bytes to "test"
13:00:48 Published 11 bytes to "test"
13:00:48 Published 11 bytes to "test"
13:00:48 Published 11 bytes to "test"
13:00:48 Published 11 bytes to "test"
13:00:48 Published 11 bytes to "test"
13:00:48 Published 11 bytes to "test"
13:00:48 Published 11 bytes to "test"
13:00:48 Published 11 bytes to "test"
13:00:48 Published 11 bytes to "test"
```

These messages also appear in `backup-test-spoke-1`. 

```txt
> watch -n 1 "nats --context=hub s report ; nats --context=spoke-1 s report; nats --context=spoke-2 s report"

Obtaining Stream stats

╭───────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                             Stream Report                                             │
├───────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream    │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├───────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ cnc       │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4222*, srv-4232, srv-4282 │
│ aggregate │ File    │ 0         │ 60       │ 5.8 KiB │ 0    │ 0       │ srv-4222, srv-4232, srv-4282* │
╰───────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭─────────────────────────────────────────────────────────────────────────────╮
│                             Replication Report                              │
├───────────┬────────┬─────────────────┬───────────────┬────────┬─────┬───────┤
│ Stream    │ Kind   │ API Prefix      │ Source Stream │ Active │ Lag │ Error │
├───────────┼────────┼─────────────────┼───────────────┼────────┼─────┼───────┤
│ aggregate │ Source │ $JS.spoke-2.API │ test          │ 1.18s  │ 0   │       │
│ aggregate │ Source │ $JS.spoke-1.API │ test          │ 1.18s  │ 0   │       │
╰───────────┴────────┴─────────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                            Stream Report                                             │
├──────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
│ test     │ File    │ 0         │ 30       │ 1.3 KiB │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
╰──────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 1.76s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭─────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                  Stream Report                                                  │
├─────────────────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream              │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├─────────────────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc            │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
│ backup-test-spoke-1 │ File    │ 0         │ 30       │ 1.3 KiB │ 0    │ 0       │ srv-4202*, srv-4262, srv-4272 │
│ test                │ File    │ 0         │ 30       │ 1.3 KiB │ 0    │ 0       │ srv-4202, srv-4262*, srv-4272 │
╰─────────────────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭───────────────────────────────────────────────────────────────────────────────────────╮
│                                  Replication Report                                   │
├─────────────────────┬────────┬─────────────────┬───────────────┬────────┬─────┬───────┤
│ Stream              │ Kind   │ API Prefix      │ Source Stream │ Active │ Lag │ Error │
├─────────────────────┼────────┼─────────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc            │ Mirror │ $JS.hub.API     │ cnc           │ 0.67s  │ 0   │       │
│ backup-test-spoke-1 │ Mirror │ $JS.spoke-1.API │ test          │ 1.25s  │ 0   │       │
╰─────────────────────┴────────┴─────────────────┴───────────────┴────────┴─────┴───────╯
```

Let's quickly demonstrate that this works the way I explained by shutting down all servers in the cluster hub.
Here our watch command for the hub can't connect any longer.

```
> watch -n 1 "nats --context=hub s report ; nats --context=spoke-1 s report; nats --context=spoke-2 s report"

nats: error: setup failed: nats: no servers available for connection
Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                            Stream Report                                             │
├──────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
│ test     │ File    │ 0         │ 30       │ 1.3 KiB │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
╰──────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 9.48s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭─────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                  Stream Report                                                  │
├─────────────────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream              │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├─────────────────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc            │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
│ backup-test-spoke-1 │ File    │ 0         │ 30       │ 1.3 KiB │ 0    │ 0       │ srv-4202*, srv-4262, srv-4272 │
│ test                │ File    │ 0         │ 30       │ 1.3 KiB │ 0    │ 0       │ srv-4202, srv-4262*, srv-4272 │
╰─────────────────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭───────────────────────────────────────────────────────────────────────────────────────╮
│                                  Replication Report                                   │
├─────────────────────┬────────┬─────────────────┬───────────────┬────────┬─────┬───────┤
│ Stream              │ Kind   │ API Prefix      │ Source Stream │ Active │ Lag │ Error │
├─────────────────────┼────────┼─────────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc            │ Mirror │ $JS.hub.API     │ cnc           │ 10.42s │ 0   │       │
│ backup-test-spoke-1 │ Mirror │ $JS.spoke-1.API │ test          │ 11.00s │ 0   │       │
╰─────────────────────┴────────┴─────────────────┴───────────────┴────────┴─────┴───────╯
```

Let's send 10 messages.

```txt
nats --context=spoke-1 pub test "hello world" --count 10
13:08:32 Published 11 bytes to "test"
13:08:32 Published 11 bytes to "test"
13:08:32 Published 11 bytes to "test"
13:08:32 Published 11 bytes to "test"
13:08:32 Published 11 bytes to "test"
13:08:32 Published 11 bytes to "test"
13:08:32 Published 11 bytes to "test"
13:08:32 Published 11 bytes to "test"
13:08:32 Published 11 bytes to "test"
13:08:32 Published 11 bytes to "test"
```

And observe them being in the stream `test` but due to the missing connectivity via `hub` not yet in `backup-test-spoke-1`.

```txt
> watch -n 1 "nats --context=hub s report ; nats --context=spoke-1 s report; nats --context=spoke-2 s report"

nats: error: setup failed: nats: no servers available for connection
Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                            Stream Report                                             │
├──────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
│ test     │ File    │ 0         │ 40       │ 1.8 KiB │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
╰──────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 53.84s │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭─────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                  Stream Report                                                  │
├─────────────────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream              │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├─────────────────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc            │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
│ backup-test-spoke-1 │ File    │ 0         │ 30       │ 1.3 KiB │ 0    │ 0       │ srv-4202*, srv-4262, srv-4272 │
│ test                │ File    │ 0         │ 30       │ 1.3 KiB │ 0    │ 0       │ srv-4202, srv-4262*, srv-4272 │
╰─────────────────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭───────────────────────────────────────────────────────────────────────────────────────╮
│                                  Replication Report                                   │
├─────────────────────┬────────┬─────────────────┬───────────────┬────────┬─────┬───────┤
│ Stream              │ Kind   │ API Prefix      │ Source Stream │ Active │ Lag │ Error │
├─────────────────────┼────────┼─────────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc            │ Mirror │ $JS.hub.API     │ cnc           │ 54.78s │ 0   │       │
│ backup-test-spoke-1 │ Mirror │ $JS.spoke-1.API │ test          │ 55.37s │ 0   │       │
╰─────────────────────┴────────┴─────────────────┴───────────────┴────────┴─────┴───────╯
```

See there is a difference in message counts.

Until I start the hub again, at which point, the message count of `backup-test-spoke-1` is identical to `test` in `spoke-1` 

```txt
> watch -n 1 "nats --context=hub s report ; nats --context=spoke-1 s report; nats --context=spoke-2 s report"

Obtaining Stream stats

╭───────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                             Stream Report                                             │
├───────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream    │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├───────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ cnc       │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4222*, srv-4232, srv-4282 │
│ aggregate │ File    │ 0         │ 70       │ 6.8 KiB │ 0    │ 0       │ srv-4222, srv-4232*, srv-4282 │
╰───────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭─────────────────────────────────────────────────────────────────────────────╮
│                             Replication Report                              │
├───────────┬────────┬─────────────────┬───────────────┬────────┬─────┬───────┤
│ Stream    │ Kind   │ API Prefix      │ Source Stream │ Active │ Lag │ Error │
├───────────┼────────┼─────────────────┼───────────────┼────────┼─────┼───────┤
│ aggregate │ Source │ $JS.spoke-1.API │ test          │ 1.24s  │ 0   │       │
│ aggregate │ Source │ $JS.spoke-2.API │ test          │ 1.24s  │ 0   │       │
╰───────────┴────────┴─────────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                            Stream Report                                             │
├──────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
│ test     │ File    │ 0         │ 40       │ 1.8 KiB │ 0    │ 0       │ srv-4242, srv-4252*, srv-4292 │
╰──────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────╮
│                           Replication Report                           │
├──────────┬────────┬─────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix  │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc │ Mirror │ $JS.hub.API │ cnc           │ 0.58s  │ 0   │       │
╰──────────┴────────┴─────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭─────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                  Stream Report                                                  │
├─────────────────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream              │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├─────────────────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ recv-cnc            │ File    │ 0         │ 10       │ 440 B   │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
│ test                │ File    │ 0         │ 30       │ 1.3 KiB │ 0    │ 0       │ srv-4202*, srv-4262, srv-4272 │
│ backup-test-spoke-1 │ File    │ 0         │ 40       │ 1.8 KiB │ 0    │ 0       │ srv-4202*, srv-4262, srv-4272 │
╰─────────────────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭───────────────────────────────────────────────────────────────────────────────────────╮
│                                  Replication Report                                   │
├─────────────────────┬────────┬─────────────────┬───────────────┬────────┬─────┬───────┤
│ Stream              │ Kind   │ API Prefix      │ Source Stream │ Active │ Lag │ Error │
├─────────────────────┼────────┼─────────────────┼───────────────┼────────┼─────┼───────┤
│ recv-cnc            │ Mirror │ $JS.hub.API     │ cnc           │ 0.32s  │ 0   │       │
│ backup-test-spoke-1 │ Mirror │ $JS.spoke-1.API │ test          │ 0.51s  │ 0   │       │
╰─────────────────────┴────────┴─────────────────┴───────────────┴────────┴─────┴───────╯
```

#### Connect Streams Across Accounts

Until now we have exchanged streams across JetStream domains, but we stayed in the same account.
Now let me show you how to exchange stream data across accounts.

This will largely be an exercise in maintaining prefixes to avoid subject overlaps.

I am using `nsc` to create accounts, and users, and modify them. 
Because I am using the NATS account resolver, when done, I can simply `push` my changes into the network. 
The `nsc` directory is in the same directory where my servers got started.
Part of my server config just uses the creds files in the nsc keys directory.
I did this, so that for this demo I have everything in one place without having to copy files around.
However, the power of our JWT based approach is that you can have that `nsc` environment anywhere.
Provided you can connect to `push`, everything will work just the same. 

The commands I show next can be translated into accounts in regular config files as well.

Here is the resulting topology.

![`imgcat topology7-mirror-cross-account.png`](topology7-mirror-cross-account.png)

The account `TEST` is the account we have been using so far.
We want to mirror the stream `aggregate` that we just created into a stream named `crossacc`.
Mirroring that particular stream allows the importing stream in the other account to be independent of the actual number of spokes.

But first we need another JetStream enabled account and user:

```txt
> nsc add account -n IMPORTER
[ OK ] generated and stored account key "ABPGFDEBTHRPZIPYEDUMLTPUXPCSEG6DVG5IKW4PS55GHWQSYVMZBROI"
[ OK ] added account "IMPORTER"
> nsc edit account --name IMPORTER --js-disk-storage -1 --js-streams -1 --js-consumer -1
[ OK ] edited account "IMPORTER"
> nsc add user --account IMPORTER -n iuser
[ OK ] generated and stored user key "UCRX3AG5BCSJPHJ3ZI3PH4FLL2CRUMCI4LOMLNDVEJEOERQX55IBMFIY"
[ OK ] generated user creds file `~/test/jetstream-leaf-nodes-demo/keys/creds/OP/IMPORTER/iuser.creds`
[ OK ] added user "iuser" to account "IMPORTER"
```

Here I created the account, enabled JetStream and created a user.

To connect our accounts we need to export the consumer API needed when using mirror.
The advantage of copying the data from one account into another is that this avoids having to explicitly create a consumer for one account in another.
The other advantage is that you don't have to write and run a program that does the copying.

```txt
> nsc add export --account TEST --name Consumer-API --service --response-type Stream --subject '$JS.hub.API.CONSUMER.>'
[ OK ] added public service export "Consumer-API"
```

Here we are exporting the consumer API as public service with a stream as response (meaning more than one message as response).
This can also be done as private export which requires a token signed by the exporting account for the importing account. Therefore you have precise control on who can import.
You can also export the entire JetStream API by exporting `$JS.hub.API.>`.
If you do so, you are giving full control over JetStream to everyone importing.
I also export `$JS.hub.API` instead of `$JS.API`. 
This is so I can pin access to a particular JetStream domain and not just to the one I connect to.

On import we change `$JS.hub.API` to `JS.test@hub.API`. 
This is done to stay clear of the $JS prefix which may get additions as new features are added to JetStream.
We give it a different prefix and subsequently specify that prefix if we want to talk to that particular JetStream. 
Btw this import renaming feature is generally available. 
Different organizations working on different applications most likely have different naming schemes. 
So when they clash, just rename on import. 
As long as the same number/type of wildcards are present, you are good.
Reordering of wildcards would be possible to, but that's for another time.

```txt
> nsc add import --account IMPORTER --src-account TEST --name Remote-Consumer-API --service --remote-subject '$JS.hub.API.CONSUMER.>' --local-subject 'JS.test@hub.API.CONSUMER.>'
[ OK ] added service import "$JS.hub.API.CONSUMER.>"
```

We also need a subject on which to deliver our data.
It is important to note that the subject we will be using later in this tutorial is a lot longer than that.
Essentially, each mirror will need a unique subject. 

I am picking the common prefix on export and subsequently add parts.

```txt
> nsc add export --account TEST --name Data-Path --response-type Stream --subject 'deliver.>'
[ OK ] added public stream export "Data-Path"
```

On import the importing account's name is added.

```txt
> nsc add import --account IMPORTER --src-account TEST --name Remote-Data-Path --remote-subject 'deliver.importer.>'
[ OK ] added stream import "deliver.importer.>"
```

Upload the changes

```txt
> nsc push -A
[ OK ] push to nats-server "nats://localhost:4222,nats://localhost:4232,nats://localhost:4282" using system account "SYS":
       [ OK ] push IMPORTER to nats-server with nats account resolver:
              [ OK ] pushed "IMPORTER" to nats-server srv-4282: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4272: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4262: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4222: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4232: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4252: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4242: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4202: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4292: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push SYS to nats-server with nats account resolver:
              [ OK ] pushed "SYS" to nats-server srv-4282: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4272: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4262: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4232: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4222: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4202: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4252: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4292: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4242: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push TEST to nats-server with nats account resolver:
              [ OK ] pushed "TEST" to nats-server srv-4282: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4272: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4262: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4232: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4222: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4242: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4292: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4252: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4202: jwt updated
              [ OK ] pushed to a total of 9 nats-server
```

`nsc` also has a command to output account export/import relationships as plantuml file.

```txt
> nsc generate diagram component --detail --output-file account-component-diagram-cross.uml ; plantuml account-component-diagram-cross.uml 
```

We generate the diagram, process it with plantuml and show the generated png.
This usually gives a better overview of the relationships between accounts. 

![`imgcat account-component-diagram-cross.png`](account-component-diagram-cross.png)

When creating the mirror I have to import from a different account.
We also specify the prefix we set on import and add the stream name to our delivery subject. 

```txt
> nats --context=hub --creds keys/creds/OP/IMPORTER/iuser.creds s add crossacc --mirror aggregate --replicas 3
? Storage backend file
? Retention Policy Limits
? Discard Policy Old
? Stream Messages Limit -1
? Message size limit -1
? Maximum message age limit -1
? Maximum individual message size -1
? Adjust mirror start No
? Import mirror from a different JetStream domain No
? Import mirror from a different account Yes
? Foreign account API prefix JS.test@hub.API
? Foreign account delivery prefix deliver.importer.crossacc
Stream crossacc was created

Information for Stream crossacc created 2021-07-28T13:35:55-04:00

Configuration:

     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited
               Mirror: aggregate, API Prefix: JS.test@hub.API, Delivery Prefix: deliver.importer.crossacc


Cluster Information:

                 Name: cluster-hub
               Leader: srv-4222
              Replica: srv-4232, current, seen 0.00s ago
              Replica: srv-4282, current, seen 0.00s ago

Mirror Information:

          Stream Name: aggregate
                  Lag: 0
            Last Seen: never
      Ext. API Prefix: JS.test@hub.API
 Ext. Delivery Prefix: deliver.importer.crossacc

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
```

Here you see, the new mirror already has all messages from before.

```txt
> nats --context=hub s report --creds keys/creds/OP/IMPORTER/iuser.creds
Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                            Stream Report                                             │
├──────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream   │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├──────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ crossacc │ File    │ 0         │ 70       │ 6.8 KiB │ 0    │ 0       │ srv-4222*, srv-4232, srv-4282 │
╰──────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────────╮
│                             Replication Report                             │
├──────────┬────────┬─────────────────┬───────────────┬────────┬─────┬───────┤
│ Stream   │ Kind   │ API Prefix      │ Source Stream │ Active │ Lag │ Error │
├──────────┼────────┼─────────────────┼───────────────┼────────┼─────┼───────┤
│ crossacc │ Mirror │ JS.test@hub.API │ aggregate     │ 1.44s  │ 0   │       │
╰──────────┴────────┴─────────────────┴───────────────┴────────┴─────┴───────╯
```

The API prefix is what we changed the API on import to.
This is so we can differentiate between our JetStream and the JetStream in the other account (in possibly the same domain).
For the delivery prefix we need to add the stream name. 
Just consider, what if I wanted to mirror the same stream twice.
This is where the stream name helps to differentiate.

While we recommend exchanging stream data via source and mirror, I have to show you how to share direct access of a durable pull consumer as well.

![`imgcat topology8-consume-cross-account.png`](topology8-consume-cross-account.png)

Since exports in `TEST` exist already, I briefly clean them up to avoid warnings about overlapping subjects.

```txt
>  nsc delete export --account TEST --subject '$JS.hub.API.CONSUMER.>'
[ OK ] deleted service export "$JS.hub.API.CONSUMER.>"
> nsc delete import --account IMPORTER --subject '$JS.hub.API.CONSUMER.>'
[ OK ] deleted service import "$JS.hub.API.CONSUMER.>"
```

What needs to be exported as service responding with a stream is the consumer's `NEXT` subject 

```txt
> nsc add export --account TEST --name Next-API --service --response-type Stream --subject '$JS.hub.API.CONSUMER.MSG.NEXT.aggregate.DUR'
[ OK ] added public service export "Next-API"
```
The `NEXT` subject consists of the prefix (with domain), consumer message next, stream name and durable consumer name.

On import we rename `$JS.hub.API` to a different prefix, say `from.test.API`.

```txt
> nsc add import --account IMPORTER --name Remote-Next-API --src-account TEST --remote-subject '$JS.hub.API.CONSUMER.MSG.NEXT.aggregate.DUR' --local-subject 'from.test.API.CONSUMER.MSG.NEXT.aggregate.DUR' --service
[ OK ] added service import "$JS.hub.API.CONSUMER.MSG.NEXT.aggregate.DUR"
```

To acknowledge messages, the ack API needs to be exported/imported as well. We do so without name changes.

```txt
> nsc add export --account TEST --name Ack-API --service --response-type Stream --subject '$JS.ACK.aggregate.DUR.>'
[ OK ] added public service export "Ack-API"
> nsc add import --account IMPORTER --name Remote-Ack-API --src-account TEST --remote-subject '$JS.ACK.aggregate.DUR.>' --service
[ OK ] added service import "$JS.ACK.aggregate.DUR.>"
```

And upload the changes:

```txt
> nsc push -A
[ OK ] push to nats-server "nats://localhost:4222,nats://localhost:4232,nats://localhost:4282" using system account "SYS":
       [ OK ] push IMPORTER to nats-server with nats account resolver:
              [ OK ] pushed "IMPORTER" to nats-server srv-4222: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4202: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4232: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4282: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4242: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4252: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4292: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4272: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4262: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push SYS to nats-server with nats account resolver:
              [ OK ] pushed "SYS" to nats-server srv-4222: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4202: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4282: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4232: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4242: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4262: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4252: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4272: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4292: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push TEST to nats-server with nats account resolver:
              [ OK ] pushed "TEST" to nats-server srv-4222: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4202: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4232: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4282: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4292: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4242: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4252: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4272: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4262: jwt updated
              [ OK ] pushed to a total of 9 nats-server
```

Create the consumer `DUR` that we already referenced in our exports/imports.
Consumer add, stream name, durable consumer name, type pull consumer, deliver all messages in the stream:

```txt
> nats --context=hub c add aggregate DUR --pull --deliver all
? Replay policy instant
? Filter Stream by subject (blank for all)
? Maximum Allowed Deliveries -1
? Maximum Acknowledgements Pending 0
Information for Consumer aggregate > DUR created 2021-07-28T13:56:16-04:00

Configuration:

        Durable Name: DUR
           Pull Mode: true
         Deliver All: true
          Ack Policy: Explicit
            Ack Wait: 30s
       Replay Policy: Instant
     Max Ack Pending: 20,000
   Max Waiting Pulls: 512

Cluster Information:

                Name: cluster-hub
              Leader: srv-4282
             Replica: srv-4232, current, seen 0.00s ago
             Replica: srv-4222, current, seen 0.00s ago

State:

   Last Delivered Message: Consumer sequence: 0 Stream sequence: 0
     Acknowledgment floor: Consumer sequence: 0 Stream sequence: 0
         Outstanding Acks: 0 out of maximum 20000
     Redelivered Messages: 0
     Unprocessed Messages: 70
            Waiting Pulls: 0 of maximum 512

```

To consume, I am overwriting the credentials specified in the context.
This is the user we created just now, consumer, next stream, durable, and this is the important bit, `js-api-prefix`.
For that, we use the API prefix `from.test.API` that we set on import. 

And now we get our first message:

```txt
> nats --context=hub --creds keys/creds/OP/IMPORTER/iuser.creds consumer next aggregate DUR --js-api-prefix=from.test.API
[13:59:04] subj: test / tries: 1 / cons seq: 1 / str seq: 1 / pending: 69

Headers:

  Nats-Stream-Source: test:J3WG6St1 1

Data:


hello world

Acknowledged message

```

Let's briefly look at what would be necessary to do in a program.

```txt
> nl -b a  main.go
     1	package main
     2
     3	import (
     4		"fmt"
     5		"os"
     6		"os/signal"
     7		"syscall"
     8		"time"
     9
    10		"github.com/nats-io/nats.go"
    11	)
    12
    13	func main() {
    14		nc, err := nats.Connect(os.Args[1], nats.Name("JS sub test"), nats.UserCredentials(os.Args[2]))
    15		defer nc.Close()
    16		if err != nil {
    17			fmt.Printf("nats connect: %v\n", err)
    18			return
    19		}
    20		js, err := nc.JetStream(nats.APIPrefix("from.test.API"))
    21		if err != nil {
    22			fmt.Printf("JetStream: %v\n", err)
    23			if js == nil {
    24				return
    25			}
    26		}
    27		s, err := js.PullSubscribe("test", "DUR", nats.Bind("aggregate", "DUR"))
    28		if err != nil {
    29			fmt.Printf("PullSubscribe: %v\n", err)
    30			return
    31		}
    32
    33		shutdown := make(chan os.Signal, 1)
    34		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
    35
    36		fmt.Printf("starting\n")
    37		for {
    38			select {
    39			case <-shutdown:
    40				return
    41			default:
    42				if m, err := s.Fetch(1, nats.MaxWait(time.Second)); err != nil {
    43					fmt.Println(err)
    44				} else {
    45
    46					if meta, err := m[0].Metadata(); err == nil {
    47						fmt.Printf("%+v\n", meta)
    48					}
    49					fmt.Println(string(m[0].Data))
    50
    51					if err := m[0].Ack(); err != nil {
    52						fmt.Printf("ack error: %+v\n", err)
    53					}
    54				}
    55			}
    56		}
    57	}
```

Similar to the CLI, we specify `nats.APIPrefix` in line `20`.
Due to very specific export, this program has only limited JetStream API access.
Therefore `nats.Bind` in line `27` provides stream and durable name explicitly.

Let's run it, connecting to the hub and using our new user.

```txt
> ./main localhost:4222 keys/creds/OP/IMPORTER/iuser.creds
starting
&{Sequence:{Consumer:3 Stream:3} NumDelivered:1 NumPending:67 Timestamp:2021-07-28 12:48:52.466147 -0400 EDT Stream:aggregate Consumer:DUR}
hello world
&{Sequence:{Consumer:4 Stream:4} NumDelivered:1 NumPending:66 Timestamp:2021-07-28 12:48:52.46615 -0400 EDT Stream:aggregate Consumer:DUR}
hello world
&{Sequence:{Consumer:5 Stream:5} NumDelivered:1 NumPending:65 Timestamp:2021-07-28 12:48:52.466152 -0400 EDT Stream:aggregate Consumer:DUR}
hello world
...
&{Sequence:{Consumer:68 Stream:68} NumDelivered:1 NumPending:2 Timestamp:2021-07-28 13:11:10.789185 -0400 EDT Stream:aggregate Consumer:DUR}
hello world
&{Sequence:{Consumer:69 Stream:69} NumDelivered:1 NumPending:1 Timestamp:2021-07-28 13:11:10.789186 -0400 EDT Stream:aggregate Consumer:DUR}
hello world
&{Sequence:{Consumer:70 Stream:70} NumDelivered:1 NumPending:0 Timestamp:2021-07-28 13:11:10.789188 -0400 EDT Stream:aggregate Consumer:DUR}
hello world
nats: timeout
^Cnats: timeout
```

There you go, messages are directly received from a durable in another account.

#### Connect Streams Through Accounts 

Finally I want to show you an account setup that addresses the same subject issue mentioned earlier.
You will be able to use the same account on each leaf node and isolate it such that you don't have to worry about cross traffic.
This will clearly increase the complexity of your setup, so you would only do this if it would otherwise make your life simpler.

![`imgcat topology9-source-cross-domain-account.png`](topology9-source-cross-domain-account.png)

Essentially we will selectively connect subjects in our accounts through a dedicated exchange account.
We will need a name that identifies a leaf node or leaf node cluster uniquely. 
As I am demonstrating JetStream, I am using domain names for that, however you are not limited to JetStream and any of your subjects can be connected in a similar way.
This image shows the direction leaf node to hub, but the setup I am showing next only needs two extra imports to enable the other direction as well.
So, let's create these accounts:

I'm creating the exchange account and a user. (No JetStream on purpose):

```txt
> nsc add account -n EXCACC
[ OK ] generated and stored account key "ACSESHB7CAGJ6R5ITCOO5GXZKA7JZ6J2BRUIIH2LDSGZSH26C3YO6N64"
[ OK ] added account "EXCACC"
> nsc add user --account EXCACC --name exp
[ OK ] generated and stored user key "UDDE7EROVEJ37LZD2RU4Z56N5X6Z45BZ7GEOTIUTI7V4AVG5LP27TDQZ"
[ OK ] generated user creds file `~/test/jetstream-leaf-nodes-demo/keys/creds/OP/EXCACC/exp.creds`
[ OK ] added user "exp" to account "EXCACC"
```

Leaf account, with JetStream enabled and a user:

```txt
> nsc add account -n LEAFACC
[ OK ] generated and stored account key "ADDIPFFGFZNLMR4OSCWX2RHBHTZBBEEV7THVWKBNJ4M7L6TDJ4YAXHY6"
[ OK ] added account "LEAFACC"
> nsc edit account -n LEAFACC --js-disk-storage -1 --js-consumer -1 --js-streams -1
[ OK ] edited account "LEAFACC"
> nsc add user --account LEAFACC --name exp
[ OK ] generated and stored user key "UCPEVXUHMULNSEE4OZ7YS4EHLYT573CREUHJ57WDZNIMJS43WS5U4GU2"
[ OK ] generated user creds file `~/test/jetstream-leaf-nodes-demo/keys/creds/OP/LEAFACC/exp.creds`
[ OK ] added user "exp" to account "LEAFACC"
```

Hub account, with JetStream enabled and a user:

```txt
> nsc add account -n HUBACC
[ OK ] generated and stored account key "AC25IY7ID2R6VNGHFROBD75EFCS7LTE3G52YRA6BPYIJDW7RVHX6IEBY"
[ OK ] added account "HUBACC"
> nsc edit account -n HUBACC --js-disk-storage -1 --js-consumer -1 --js-streams -1
[ OK ] edited account "HUBACC"
> nsc add user --account HUBACC --name imp
[ OK ] generated and stored user key "UANIPWE67ANES3FLPJMD626TL62BYSL2U3L2YJ5PWNZ76BHPXBPUFSGR"
[ OK ] generated user creds file `~/test/jetstream-leaf-nodes-demo/keys/creds/OP/HUBACC/imp.creds`
[ OK ] added user "imp" to account "HUBACC"
```

And pushing all accounts:

```txt
> nsc push -A
[ OK ] push to nats-server "nats://localhost:4222,nats://localhost:4232,nats://localhost:4282" using system account "SYS":
       [ OK ] push EXCACC to nats-server with nats account resolver:
              [ OK ] pushed "EXCACC" to nats-server srv-4222: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4202: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4282: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4232: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4242: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4262: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4292: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4272: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4252: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push HUBACC to nats-server with nats account resolver:
              [ OK ] pushed "HUBACC" to nats-server srv-4222: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4282: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4232: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4202: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4262: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4292: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4242: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4272: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4252: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push IMPORTER to nats-server with nats account resolver:
              [ OK ] pushed "IMPORTER" to nats-server srv-4222: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4202: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4282: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4232: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4272: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4242: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4262: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4252: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4292: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push LEAFACC to nats-server with nats account resolver:
              [ OK ] pushed "LEAFACC" to nats-server srv-4222: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4232: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4202: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4282: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4292: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4252: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4262: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4242: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4272: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push SYS to nats-server with nats account resolver:
              [ OK ] pushed "SYS" to nats-server srv-4222: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4282: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4202: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4232: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4262: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4272: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4242: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4252: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4292: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push TEST to nats-server with nats account resolver:
              [ OK ] pushed "TEST" to nats-server srv-4222: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4202: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4232: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4282: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4242: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4292: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4252: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4262: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4272: jwt updated
              [ OK ] pushed to a total of 9 nats-server
```

Let's briefly change our watch command on the left to make use of our new users.
I'm connecting to the hub, hubaccount user, stream report.
The spoke, leaf account user,  stream report.
The second spoke is the same as first.

```txt
> watch -n 1 "nats --context=hub --creds keys/creds/OP/HUBACC/imp.creds s report ; \
 nats --context=spoke-1 --creds keys/creds/OP/LEAFACC/exp.creds s report ; \
 nats --context=spoke-2 --creds keys/creds/OP/LEAFACC/exp.creds s report"

Obtaining Stream stats

No Streams defined
Obtaining Stream stats

No Streams defined
Obtaining Stream stats

No Streams defined

```

We have nothing defined yet, which is why they are empty. 

Furthermore, ONLY put that exchange account `EXCACC` into the remotes.
Not listing the other accounts is what isolates them from the hub and each other.
I have looked up the account ID earlier using `nsc list keys`. 
Register account EXCACC as remote:

```txt
> nl -b a leafnode-remotes.cfg
     1	HUB-URLS=["nats-leaf://localhost:4224","nats-leaf://localhost:4234","nats-leaf://localhost:4284"]
     2	leafnodes {
     3	no_advertise: true
     4	    remotes = [
     5			{
     6				urls: $HUB-URLS
     7				account: ADECCNBUEBWZ727OMBFSN7OMK2FPYRM52TJS25TFQWYS76NPOJBN3KU4
     8	 			credentials: keys/creds/OP/SYS/sys.creds
     9			},
    10			{
    11				urls: $HUB-URLS
    12				account: AA5C56FAETBTUCYM7NC5BFBYFTKLOABIOIFPQDHO4RUEAPSN3FTY5R4G
    13				credentials: keys/creds/OP/TEST/leaf.creds
    14			},
    15			{
    16				urls: $HUB-URLS
    17				account: ACSESHB7CAGJ6R5ITCOO5GXZKA7JZ6J2BRUIIH2LDSGZSH26C3YO6N64
    18				credentials: keys/creds/OP/EXCACC/exp.creds
    19			},
    20		]
    21	}
```

A side effect of only having the exchange account in the remotes is that you can control access via exports and imports that can be pushed by `nsc`. 
You do not have to touch remotes in a config going forward. 
To demonstrate this I pushed accounts without any exports and imports already.
Right now I have to restart nats-server such that the change is picked up.

```txt
> pkill nats-server 
```

Now we are connecting the accounts.

Selectively, we export subjects we want to be able to send to and receive from in other accounts and domains.
For every domain and account I wish to connect, I'm exporting the entire domain specific JetStream API as well as a delivery subject.
You can also add dedicated subjects to communicate with regular NATS.

For every account, I export a delivery subject containing account name and domain as well the JetStream domain specific API.

```txt
> nsc add export --account HUBACC  --service --response-type Stream --subject '$JS.hub.API.>'
[ OK ] added public service export "$JS.hub.API.>"
> nsc add export --account LEAFACC --service --response-type Stream --subject '$JS.spoke-1.API.>'
[ OK ] added public service export "$JS.spoke-1.API.>"
> nsc add export --account LEAFACC --service --response-type Stream --subject '$JS.spoke-2.API.>'
[ OK ] added public service export "$JS.spoke-2.API.>"
> nsc add export --account HUBACC  --subject 'deliver.hubacc.hub.>'
[ OK ] added public stream export "deliver.hubacc.hub.>"
> nsc add export --account LEAFACC --subject 'deliver.leafacc.spoke-1.>'
[ OK ] added public stream export "deliver.leafacc.spoke-1.>"
> nsc add export --account LEAFACC --subject 'deliver.leafacc.spoke-2.>'
[ OK ] added public stream export "deliver.leafacc.spoke-2.>"
```

All of these exports are then imported into the exchange account.

Please note that every subject already contains a fixed portion, that functions as type, the exporting account name, as well as the domain.
In that order!

Where this is not the case, such as the JetStream API, I remap accordingly.
This allows us to properly identify streams and services in an account in a leaf node as identified by it's domain.

```txt
> nsc add import --account EXCACC --src-account HUBACC  --service --remote-subject '$JS.hub.API.>'     --local-subject '$JS.hubacc.hub.API.>'
[ OK ] added service import "$JS.hub.API.>"
> nsc add import --account EXCACC --src-account LEAFACC --service --remote-subject '$JS.spoke-1.API.>' --local-subject '$JS.leafacc.spoke-1.API.>'
[ OK ] added service import "$JS.spoke-1.API.>"
> nsc add import --account EXCACC --src-account LEAFACC --service --remote-subject '$JS.spoke-2.API.>' --local-subject '$JS.leafacc.spoke-2.API.>'
[ OK ] added service import "$JS.spoke-2.API.>"
> nsc add import --account EXCACC --src-account HUBACC  --remote-subject 'deliver.hubacc.hub.>'
[ OK ] added stream import "deliver.hubacc.hub.>"
> nsc add import --account EXCACC --src-account LEAFACC --remote-subject 'deliver.leafacc.spoke-1.>'
[ OK ] added stream import "deliver.leafacc.spoke-1.>"
> nsc add import --account EXCACC --src-account LEAFACC --remote-subject 'deliver.leafacc.spoke-2.>'
[ OK ] added stream import "deliver.leafacc.spoke-2.>"
```

Then we re-export all subjects together. This is where the type comes in handy with `type.>`.

```txt
> nsc add export --account EXCACC --service --response-type Stream --subject '$JS.>'
[ OK ] added public service export "$JS.>"
> nsc add export --account EXCACC --subject 'deliver.>'
[ OK ] added public stream export "deliver.>"
```

Finally, import into every account you want connected. 
I'm more specific than in the previous export as I need to avoid a self-import cycle with say `deliver.hubacc.>`.
If you want to be more specific and pin an import to a domain, just add it. 

```txt
> nsc add import --account HUBACC --src-account EXCACC --service --remote-subject '$JS.leafacc.>'
[ OK ] added service import "$JS.leafacc.>"
> nsc add import --account HUBACC --src-account EXCACC --remote-subject 'deliver.leafacc.>'
[ OK ] added stream import "deliver.leafacc.>"
```

Inspect the changes 

```txt
> nsc generate diagram component --detail --output-file account-component-diagram-through.uml ; plantuml account-component-diagram-through.uml 
```

![`imgcat account-component-diagram-through.png`](account-component-diagram-through.png)

Upload all changes:

```txt
> nsc push -A
[ OK ] push to nats-server "nats://localhost:4222,nats://localhost:4232,nats://localhost:4282" using system account "SYS":
       [ OK ] push EXCACC to nats-server with nats account resolver:
              [ OK ] pushed "EXCACC" to nats-server srv-4282: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4232: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4222: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4272: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4202: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4252: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4262: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4292: jwt updated
              [ OK ] pushed "EXCACC" to nats-server srv-4242: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push HUBACC to nats-server with nats account resolver:
              [ OK ] pushed "HUBACC" to nats-server srv-4282: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4232: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4222: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4202: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4262: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4272: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4252: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4292: jwt updated
              [ OK ] pushed "HUBACC" to nats-server srv-4242: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push IMPORTER to nats-server with nats account resolver:
              [ OK ] pushed "IMPORTER" to nats-server srv-4282: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4222: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4232: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4242: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4292: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4262: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4252: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4272: jwt updated
              [ OK ] pushed "IMPORTER" to nats-server srv-4202: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push LEAFACC to nats-server with nats account resolver:
              [ OK ] pushed "LEAFACC" to nats-server srv-4282: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4232: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4222: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4272: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4252: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4202: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4262: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4292: jwt updated
              [ OK ] pushed "LEAFACC" to nats-server srv-4242: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push SYS to nats-server with nats account resolver:
              [ OK ] pushed "SYS" to nats-server srv-4282: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4222: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4232: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4292: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4242: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4202: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4252: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4272: jwt updated
              [ OK ] pushed "SYS" to nats-server srv-4262: jwt updated
              [ OK ] pushed to a total of 9 nats-server
       [ OK ] push TEST to nats-server with nats account resolver:
              [ OK ] pushed "TEST" to nats-server srv-4282: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4222: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4232: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4292: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4242: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4272: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4202: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4252: jwt updated
              [ OK ] pushed "TEST" to nats-server srv-4262: jwt updated
              [ OK ] pushed to a total of 9 nats-server
```

Then we create a stream in each leaf domain using the stream configuration test from earlier.
Because it is possible, I do so by connecting to the `hub`, using hub account credentials.
A domain essentially means we use the prefix `$JS.<domain>.API`.

When connected to the hub, because I imported `$JS.leafacc.spoke-1.API` and I neglected to avoid `$JS`, I can use the domain name `leafacc.spoke-1` in the `nats` CLI.

```txt
> nats --context=hub --creds keys/creds/OP/HUBACC/imp.creds s add --config test --js-domain leafacc.spoke-1
Stream test was created

Information for Stream test created 2021-07-28T14:49:29-04:00

Configuration:

             Subjects: test
     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited


Cluster Information:

                 Name: cluster-spoke-1
               Leader: srv-4242
              Replica: srv-4252, current, seen 0.00s ago
              Replica: srv-4292, current, seen 0.00s ago

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
> nats --context=hub --creds keys/creds/OP/HUBACC/imp.creds s add --config test --js-domain leafacc.spoke-2
Stream test was created

Information for Stream test created 2021-07-28T14:49:37-04:00

Configuration:

             Subjects: test
     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited


Cluster Information:

                 Name: cluster-spoke-2
               Leader: srv-4272
              Replica: srv-4262, current, seen 0.00s ago
              Replica: srv-4202, current, seen 0.00s ago

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
```

There we go, streams are created.

```txt
> watch -n 1 "nats --context=hub --creds keys/creds/OP/HUBACC/imp.creds s report ; \
 nats --context=spoke-1 --creds keys/creds/OP/LEAFACC/exp.creds s report ; \
 nats --context=spoke-2 --creds keys/creds/OP/LEAFACC/exp.creds s report"

Obtaining Stream stats

No Streams defined
Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ test   │ File    │ 0         │ 0        │ 0 B   │ 0    │ 0       │ srv-4242*, srv-4252, srv-4292 │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ test   │ File    │ 0         │ 0        │ 0 B   │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯
```

Now I publish messages to each leaf cluster, using credentials for the leaf account.

```txt
> nats --context=spoke-1 --creds keys/creds/OP/LEAFACC/exp.creds pub test "" --count 10
14:51:39 Published 0 bytes to "test"
14:51:39 Published 0 bytes to "test"
14:51:39 Published 0 bytes to "test"
14:51:39 Published 0 bytes to "test"
14:51:39 Published 0 bytes to "test"
14:51:39 Published 0 bytes to "test"
14:51:39 Published 0 bytes to "test"
14:51:39 Published 0 bytes to "test"
14:51:39 Published 0 bytes to "test"
14:51:39 Published 0 bytes to "test"
> nats --context=spoke-2 --creds keys/creds/OP/LEAFACC/exp.creds pub test "" --count 10
14:51:45 Published 0 bytes to "test"
14:51:45 Published 0 bytes to "test"
14:51:45 Published 0 bytes to "test"
14:51:45 Published 0 bytes to "test"
14:51:45 Published 0 bytes to "test"
14:51:45 Published 0 bytes to "test"
14:51:45 Published 0 bytes to "test"
14:51:45 Published 0 bytes to "test"
14:51:45 Published 0 bytes to "test"
14:51:45 Published 0 bytes to "test"
```

Because of the isolation, the messages stay in the domain in which they originated.
Every stream contains 10 messages:

```txt
> watch -n 1 "nats --context=hub --creds keys/creds/OP/HUBACC/imp.creds s report ; \
 nats --context=spoke-1 --creds keys/creds/OP/LEAFACC/exp.creds s report ; \
 nats --context=spoke-2 --creds keys/creds/OP/LEAFACC/exp.creds s report"

Obtaining Stream stats

No Streams defined
Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ test   │ File    │ 0         │ 10       │ 340 B │ 0    │ 0       │ srv-4242*, srv-4252, srv-4292 │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ test   │ File    │ 0         │ 10       │ 340 B │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯
```

As a final test let's connect to the hub using the same credentials. 
These messages will not be received because of our isolation scheme. 

```txt
> nats --context=hub --creds keys/creds/OP/LEAFACC/exp.creds pub test "" --count 10
14:53:44 Published 0 bytes to "test"
14:53:44 Published 0 bytes to "test"
14:53:44 Published 0 bytes to "test"
14:53:44 Published 0 bytes to "test"
14:53:44 Published 0 bytes to "test"
14:53:44 Published 0 bytes to "test"
14:53:44 Published 0 bytes to "test"
14:53:44 Published 0 bytes to "test"
14:53:44 Published 0 bytes to "test"
14:53:44 Published 0 bytes to "test"
```

Message count is unaltered as spokes are isolated from the hub as well.

```txt
> watch -n 1 "nats --context=hub --creds keys/creds/OP/HUBACC/imp.creds s report ; \
 nats --context=spoke-1 --creds keys/creds/OP/LEAFACC/exp.creds s report ; \
 nats --context=spoke-2 --creds keys/creds/OP/LEAFACC/exp.creds s report"

Obtaining Stream stats

No Streams defined
Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ test   │ File    │ 0         │ 10       │ 340 B │ 0    │ 0       │ srv-4242*, srv-4252, srv-4292 │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ test   │ File    │ 0         │ 10       │ 340 B │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯
```

Now let's create the importing stream inside the hub and the hub account. 
I'm importing from a different account with the prefix `$JS.leafacc.spoke-1.API`.
The delivery prefix is `deliver.leafacc.spoke-1.hubacc.hub.aggregate`.
For `spoke-2` the corresponding prefix and deliver prefix are picked.

```txt
> nats --context=hub --creds keys/creds/OP/HUBACC/imp.creds s add aggregate --replicas 3 --source test --source test
? Storage backend file
? Retention Policy Limits
? Discard Policy Old
? Stream Messages Limit -1
? Message size limit -1
? Maximum message age limit -1
? Maximum individual message size -1
? Duplicate tracking time window 2m
? Adjust source "test" start No
? Import "test" from a different JetStream domain No
? Import "test" from a different account Yes
? test Source foreign account API prefix $JS.leafacc.spoke-1.API
? test Source foreign account delivery prefix deliver.leafacc.spoke-1.hubacc.hub.aggregate
? Adjust source "test" start No
? Import "test" from a different JetStream domain No
? Import "test" from a different account Yes
? test Source foreign account API prefix $JS.leafacc.spoke-2.API
? test Source foreign account delivery prefix deliver.leafacc.spoke-2.hubacc.hub.aggregate
Stream aggregate was created

Information for Stream aggregate created 2021-07-28T14:56:55-04:00

Configuration:

     Acknowledgements: true
            Retention: File - Limits
             Replicas: 3
       Discard Policy: Old
     Duplicate Window: 2m0s
     Maximum Messages: unlimited
        Maximum Bytes: unlimited
          Maximum Age: 0.00s
 Maximum Message Size: unlimited
    Maximum Consumers: unlimited
              Sources: test, API Prefix: $JS.leafacc.spoke-1.API, Delivery Prefix: deliver.leafacc.spoke-1.hubacc.hub.aggregate
                       test, API Prefix: $JS.leafacc.spoke-2.API, Delivery Prefix: deliver.leafacc.spoke-2.hubacc.hub.aggregate


Cluster Information:

                 Name: cluster-hub
               Leader: srv-4222
              Replica: srv-4282, current, seen 0.00s ago
              Replica: srv-4232, current, seen 0.00s ago

Source Information:

          Stream Name: test
                  Lag: 0
            Last Seen: 0.00s
      Ext. API Prefix: $JS.leafacc.spoke-1.API
 Ext. Delivery Prefix: deliver.leafacc.spoke-1.hubacc.hub.aggregate

          Stream Name: test
                  Lag: 0
            Last Seen: 0.00s
      Ext. API Prefix: $JS.leafacc.spoke-2.API
 Ext. Delivery Prefix: deliver.leafacc.spoke-2.hubacc.hub.aggregate

State:

             Messages: 0
                Bytes: 0 B
             FirstSeq: 0
              LastSeq: 0
     Active Consumers: 0
```

About the delivery prefix. You are essentially creating a stream that copies from one account in a domain to another account in another domain.
Thus I'd recommend that the delivery prefix consists of all these parts: type, from account, from domain, to account, to domain, importing stream name.

There you go, you can see it's working on the left hand side.

```txt
> watch -n 1 "nats --context=hub --creds keys/creds/OP/HUBACC/imp.creds s report ; \
 nats --context=spoke-1 --creds keys/creds/OP/LEAFACC/exp.creds s report ; \
 nats --context=spoke-2 --creds keys/creds/OP/LEAFACC/exp.creds s report"

Obtaining Stream stats

╭───────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                             Stream Report                                             │
├───────────┬─────────┬───────────┬──────────┬─────────┬──────┬─────────┬───────────────────────────────┤
│ Stream    │ Storage │ Consumers │ Messages │ Bytes   │ Lost │ Deleted │ Replicas                      │
├───────────┼─────────┼───────────┼──────────┼─────────┼──────┼─────────┼───────────────────────────────┤
│ aggregate │ File    │ 0         │ 20       │ 1.7 KiB │ 0    │ 0       │ srv-4222*, srv-4232, srv-4282 │
╰───────────┴─────────┴───────────┴──────────┴─────────┴──────┴─────────┴───────────────────────────────╯

╭─────────────────────────────────────────────────────────────────────────────────────╮
│                                 Replication Report                                  │
├───────────┬────────┬─────────────────────────┬───────────────┬────────┬─────┬───────┤
│ Stream    │ Kind   │ API Prefix              │ Source Stream │ Active │ Lag │ Error │
├───────────┼────────┼─────────────────────────┼───────────────┼────────┼─────┼───────┤
│ aggregate │ Source │ $JS.leafacc.spoke-1.API │ test          │ 0.24s  │ 0   │       │
│ aggregate │ Source │ $JS.leafacc.spoke-2.API │ test          │ 0.24s  │ 0   │       │
╰───────────┴────────┴─────────────────────────┴───────────────┴────────┴─────┴───────╯

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ test   │ File    │ 0         │ 10       │ 340 B │ 0    │ 0       │ srv-4242*, srv-4252, srv-4292 │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯

Obtaining Stream stats

╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                          Stream Report                                           │
├────────┬─────────┬───────────┬──────────┬───────┬──────┬─────────┬───────────────────────────────┤
│ Stream │ Storage │ Consumers │ Messages │ Bytes │ Lost │ Deleted │ Replicas                      │
├────────┼─────────┼───────────┼──────────┼───────┼──────┼─────────┼───────────────────────────────┤
│ test   │ File    │ 0         │ 10       │ 340 B │ 0    │ 0       │ srv-4202, srv-4262, srv-4272* │
╰────────┴─────────┴───────────┴──────────┴───────┴──────┴─────────┴───────────────────────────────╯
```

Please be aware that it does not matter that the hub is actually a hub.
I want you to take away that you can chain accounts and connect through them.
Just think of a scenario where you have multiple leaf nodes into an NGS account, collecting data locally.
One of your leaf nodes is specifed bigger and this is where you aggregate streams and do all of your analytics. 

One thing I hope you noticed is the lack of host names and the flexibility that gives you.
Of course, they are needed in the server configuration to create the NATS network.
They are also needed to connect to that network, but you don't need different ones for each of your applications. 
Instead you get to focus on your data and how it flows. 

We will continue to add improvements to the setups and workflows introduced in this tutorial. 
I certainly have noticed a few rough edges, so it is worthwhile checking back every once in a while.
If you have questions, please contact the NATS team on Slack [slack.nats.io](https://slack.nats.io).

#### Relevant Links

* get nats cli at: https://github.com/nats-io/natscli
* get nsc cli at: https://github.com/nats-io/nsc
* docs: https://docs.nats.io/
* docs for config: https://docs.nats.io/nats-server/configuration
* doc for config based accounts: https://docs.nats.io/nats-server/configuration/securing_nats/accounts
* jwt deep dive: https://docs.nats.io/developing-with-nats/tutorials/jwt
