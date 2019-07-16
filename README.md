<img src="./assets/bitgodine.png" width="150">


[**bitgodine**](https://github.com/xn3cr0nx/bitgodine) is a Go implementation of a Bitcoin forensic analysis tool that enable you to explore the Bitcoin blockchain and analyze the transactions with a set of heuristics.

Bitgodine provides an advanced cli to interact with all its functionalities. We foresee to provide soon a user interface to make the user experience easier to non developer users.

## Requirements

Bitgodine need external data layers to be able to parse and store the blockchain and associated metadata. The repository is provided with a docker-compose file to bootstrap the environment. The docker-compose file contains:

- Dgraph: graph database to store the blockchain and clusters (3 instances, web UI, server, and grpc service)
- Postgres: database to store tags and clusters relations
- Pgadmin: web UI to manage Postgres

It's recommended the use of docker. You can install docker [here](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-using-the-convenience-script) and docker-compose [here](https://docs.docker.com/compose/install/#install-compose).

Docker doesn't acctually spins up anything about blockchain, hence to be able to use the tool you have to be able to read the Bitcoin blockchain stored somewhere. If you don't know how to install a Bitcoin full node check the [Bitcoin Core installation guide](https://bitcoin.org/en/full-node). Bitgodine will automatically look for data files in default bitcoin directory.

## Installing

<!-- Install it with:

```
go get -u github.com/xn3cr0nx/bitgodine
``` -->
Currently you have to clone repository, but installation and launch of docker-compose will be automated. Clone the respository and install the binary:

```bash
make install
```

## Getting Started

Once you have at least a part of the blockchain synced you can start bitgodine:

```bash
bitgodine start
```

You will see the sync process going on and a log line every 100 synced blocks. You can stop the process in wherever moment with Ctrl+C, bitgodine will handle it gracefully. You will be able to restart the process in another moment from where you stopped. You can add *--debug* flag for a more verbose output.

Check the current synced block height with:

```bash
bitgodine block height
```

### Analysis

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

### Explorer

You can use bitgodine as a cli explorer to fast check some info about block and transaction. Check it out:

```bash
bitgodine block height 10 --verbose
```

### Recovery and startup

You are able in every moment to deleted synced data and restart the process using:

```bash
bitgodine rm
```

In case the persistent set corrupts you can restore it using

```bash
bitgodine cluster recovery
```

## How to contribute

You can find all the info about contributing to this project (every help is welcomed) in the CONTRIBUTING.md file.

## Roadmap

- [x] Graceful shutdown
- [x] Persistent clusters
- [ ] Bitgodine as a daemon
- [ ] Coinjoin entropy analysis
- [ ] Multisignature addresses support
- [ ] Realtime syncing
- [ ] Multicurrency parser
