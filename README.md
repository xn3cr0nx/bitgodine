<img src="./assets/bitgodine.png" width="150">


**Bitgodine** is a Go implementation of a Bitcoin forensic analysis tool that enable you to explore the Bitcoin blockchain and analyze the transactions with a set of heuristics.

Bitgodine provides an advanced cli to interact with all its functionalities. We foresee to provide soon a user interface to make the user experience easier to non developer users.

## How to use it

Install it with:

```
go get github.com/xn3cr0nx/bitgodine
```

To be able to use the tool you have to be able to read the Bitcoin blockchain stored somewhere. If you don't know how to install a Bitcoin full node check the [Bitcoin Core installation guide](https://bitcoin.org/en/full-node).

Once you have at least a part of the blockchain synced you can sync bitgodine:

```
bitgodine sync
```

You will see the sync process going on and a log line every 100 synced blocks. You can stop the process in wherever moment with Ctrl+C, bitgodine will handle it gracefully. You will be able to restart the process in another moment from where you stopped.

Once you have some part of the blockchain synced in bitgodine you can use the analysis command in the cli. Try:

```
bitgodine analysis --range 0-50
```

to analyze the transactions in the first 50 blocks. You will se a table with the result containing the applied heuristics and their effect.

You can use bitgodine as a cli explorer to fast check some info about block and transaction. Check it out:

```
bitgodine block height 10 --verbose
```

## How to contribute

You can find all the info about contributing to this project (every help is welcomed) in the CONTRIBUTING.md file.

## Roadmap

- [x] Graceful shutdown
- [ ] Realtime syncing
- [ ] Multisignature addresses support
- [ ] Segwit addresses support
