# Siberite
[![License][License-Image]][License-Url] [![Build][Build-Status-Image]][Build-Status-Url] [![Release][Release-Image]][Release-Url]

Siberite is a simple LevelDB-backed message queue server<br>
([twitter/kestrel](https://github.com/twitter/kestrel), [wavii/darner](https://github.com/wavii/darner) rewritten in Go).

Siberite is a simple message queue server.  Unlike in-memory servers like [Redis](http://redis.io/), Siberite is
designed to handle queues much larger than what can fit in RAM. And unlike enterprise-level servers such as
[RabbitMQ](http://www.rabbitmq.com/), Siberite stores all messages **out of process**,
using [goleveldb](https://github.com/syndtr/goleveldb) for persistent storage.

The result is a durable queue server that uses minimal in-resident memory, regardless of the queue size.

Siberite is based on Robey Pointer's [Kestrel](https://github.com/robey/kestrel), a simple, distributed message queue.
Like Kestrel, Siberite follows the "No talking! Shhh!" approach to distributed queues:
A single Siberite server maintains a set of queues identified by name. Each queue operates as a strictly-ordered FIFO,
while querying from multiple Siberite servers results in a loosely-ordered distributed queue.
Siberite also supports Kestrel's two-phase reliable fetch. If a client disconnects before confirming
that it has processed a message, the message will be handed off to the next client.

Compared to Kestrel and Darner, Siberite is easier to build, maintain, and distribute.
It consumes significantly less memory than Kestrel and offers the ability
to consume a queue multiple times using durable cursors.

## Features

1. **Durable cursors for multiple reads**

  - Siberite clients can consume a single source queue multiple times using the `get <queue>.<cursor_name>` syntax.
  - Normally, with the `get <queue>` syntax, the returned message is expired and deleted from the queue.
  - Using the cursor syntax `get <queue>.<cursor_name>`, a durable cursor is initialized. It advances with every read without deleting messages from the source queue. There is no limit on the number of cursors per queue.
  - If you resume reading from the queue using the standard syntax, Siberite will continue deleting already-served messages from the queue head. Any existing cursor that points to an expired message will restart reading from the current queue head on the next read.
  - Durable cursors also support two-phase reliable reads. Failed reads for each cursor are stored in the cursor’s persistent queue and served to other cursor readers.

2. **Fanout queues**

  - Siberite allows inserting a message into multiple queues simultaneously using the following syntax: `set <queue>+<another_queue>+<third_queue> ...`


## Benchmarks

[Siberite performance benchmarks](docs/benchmarks.md)


## Build

Make sure your `GOPATH` is correct

```
go get github.com/bogdanovich/siberite
cd $GOPATH/src/github.com/bogdanovich/siberite
go get ./...
go build siberite.go
mkdir ./data
./siberite -listen localhost:22133 -data ./data
2015/09/22 06:29:38 listening on 127.0.0.1:22133
2015/09/22 06:29:38 initializing...
2015/09/22 06:29:38 data directory:  ./data
```

or download [darwin-x86_64 or linux-x86_64 builds](https://github.com/bogdanovich/siberite/releases)

## Protocol

Siberite follows the same protocol as [Kestrel](http://github.com/robey/kestrel/blob/master/docs/guide.md#memcache),
which is the memcache TCP text protocol.

[List of compatible clients](docs/clients.md)

## Telnet demo

```
telnet localhost 22133
Connected to localhost.
Escape character is '^]'.

set work 0 0 10
1234567890
STORED

set work 0 0 2
12
STORED

get work
VALUE work 0 10
1234567890
END

get work/open
VALUE work 0 2
12
END

get work/close
END

stats
STAT uptime 47
STAT time 1443308758
STAT version siberite-0.4.1
STAT curr_connections 1
STAT total_connections 1
STAT cmd_get 2
STAT cmd_set 2
STAT queue_work_items 0
STAT queue_work_open_transactions 0
END

# other commands:
# get work/peek
# get work/open
# get work/close/open
# get work/abort
# get work.cursor_name
# get work.cursor_name/open
# get work.my_cursor/close/open
# set work+fanout_queue
# flush work
# delete work
# flush_all
# quit
```


## Not supported

  - Waiting a given time limit for a new item to arrive /t=<milliseconds> (allowed by protocol but does nothing)

[License-Url]: http://opensource.org/licenses/Apache-2.0
[License-Image]: https://img.shields.io/hexpm/l/plug.svg
[Build-Status-Url]: https://travis-ci.org/bogdanovich/siberite
[Build-Status-Image]: https://travis-ci.org/bogdanovich/siberite.svg?branch=master
[Release-Url]: https://github.com/bogdanovich/siberite/releases/latest
[Release-image]: https://img.shields.io/badge/release-v0.6-blue.svg
