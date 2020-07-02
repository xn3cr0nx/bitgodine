# Bitgodine

<img src="./assets/bitgodine_finder.png" width="150">

Chain analysis is the attempt to deanonymize pseudo confidential transactions in Bitcoin’s blockchain. This work is usually made by companies working with big institutions, and their techniques are largely unknown. Chain analysis should be open in order to help the technology to establish the best practices and mitigate these techniques.

[**bitgodine**](https://github.com/xn3cr0nx/bitgodine) is a new open source Bitcoin chain analysis platform, based on previous Politecnico's work, Bitiodine.
This work leverages previous work on chain analysis' heuristics to provide a new Bitcoin’s flows tracing approach, based on probabilistic reliability of heuristics on secure basis heuristics.
In this way we provide a convenient platform to trace Bitcoin's flows specifying their probability and entities tagging.

## Architecture

Bitgodine is composed by different services

### Parser

Parser is in charge of parsing .dat blockchain files, indexing the chain and storing it in a convenient way for following analysis.

### Clusterizer

Clusterizer generates persistent clusters of addresses, exploiting the common input heuristic and a disjoint set structure. Clusters can identify groups of addresses related to identities in order to expand semantic tagging to an entire group of addresses.

### Server

Server exposes a REST API proving both endpoints about chain exploring and analysis. [bitgodine_dashboard](https://github.com/xn3cr0nx/bitgodine_dashboard) is the most convenient way to use bitgodine explorer and tracing service.

### Spider

Spider is a web crawler scraping different websites to build a sematinc tags database in order to cross tags and clusters adding info on tracing flows.

## Requirements

As decribed, Bitgodine is composed by different movign pieces. We highly recommend running the platform using Docker in order to deploy the entire infrastructure easily.

### Docker

The code provides both a Dockerfile for each services and docker-compose files to bootstrap the platform locally and in production. Production docker-compose file is provided to use official images and external disk to have enough space available. Development docker-compose file will build services locally.

First, install [docker](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-using-the-convenience-script) and [docker-compose](https://docs.docker.com/compose/install/#install-compose).

In order to bootstrap the infrastructure with Docker you just need to type:

```bash
docker-compose up -d
```

To run the production version:

```bash
docker-compose up -d -f docker-compose.prod.yml
```

### Locally

If you want to run bitgodine on bare locally you will need to install some dependency to run the platform. First install [go](https://golang.org/doc/install).

The docker-compose file includes the following services you will need to install locally in order to setup the system:

- bitcoind: Bitcoin core full node. Use [Bitcoin Core installation guide](https://bitcoin.org/en/full-node) to install locally, Bitgodine will automatically look for data files in default bitcoind directory.
- Postgres: database to store tags and clusters relations
- Pgadmin: (optional) web UI to manage Postgres

### Build

You can use go to build each service. Bitgodine provides a convenient Makefile to manage these operations. The easiest way to build services is using:

```bash
make build
```

This will build *server*, *parser*, *clusterizer*, *spider* binaries in `build` folder.
In order to install services to be able to run them as part of the PATH use install cmd instead.

### Getting Started

Once you have at least a part of the blockchain synced through the node you can start the parser:

```bash
parser start
```

You will see the sync process going on and a log line every 100 synced blocks. You can stop the process in wherever moment with `Ctrl+C`, bitgodine will handle it gracefully. The process is persistent, hence you will be able to restart the process in another moment from where you stopped. You can add _--debug_ flag for a more verbose output.

Check the current synced block height with:

```bash
bitgodine block height
```

## Analysis

Once you have some part of the blockchain synced in bitgodine you can use the analysis command in the cli. Try:

```bash
bitgodine analysis --range 0-50
```

to analyze the transactions in the first 50 blocks. You will se a table with the result containing the applied heuristics and their effect.

You can specify a transaction to be analyzed by hash:

```bash
bitgodine analysis tx 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
```

or you can specify which heuristic you want to test:

```bash
bitgodine analysis reuse 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
```

Available heuristics commands:

- backward (Backward heuristic)
- behaviour (Client behaviour heuristic)
- forward (Forward)
- locktime (Locktime heuristic)
- optimal (Optimal change heuristic)
- peeling (Peeling chain heuristic)
- power (Power of ten heuristic)
- reuse (Reuse address heuristic)
- type (Address type heuristic)

### Tags

Bitgodine needs a set of tags in the database to be able to associate them with synced cluster. For this reason it can download ~100k tags using some spiders that can be concurrently launched on blockchain.com and bitcoinabuse.com (I would use a VPN to avoid to be blacklisted):

```bash
bitgodine tag sync
```

You can then check if a tag is present by name or address:

```bash
bitgodine tag get --name jgarzik
```

Tags contain the address, tag name, metadata that commonly is a proof link and whether the tag is verified or not

### Clusters

Parsing process stores a persisten disjoint set in dgraph out of it. You can check cluster size using size command. In every moment it is possible to export the cluster to a csv or to the db file using the export command:

```bash
bitgodine cluster export --csv
```

Exporting to Postgres database in necessaire to enable joined operation with stored address tags. Using tag command bitgodine will show up a table with tagged cluster by address or tag:

```bash
bitgodine cluster tag <address || tagname>
```

## How to contribute

You can find all the info about contributing to this project (every help is welcomed) in the CONTRIBUTING.md file.

## Roadmap

- [ ] Testing
- [x] Graceful shutdown
- [x] Persistent clusters
- [ ] Coinjoin entropy analysis
- [ ] Multisignature addresses support
- [ ] Realtime syncing
