# Siberite
[![License][License-Image]][License-Url] [![Build][Build-Status-Image]][Build-Status-Url] [![Release][Release-Image]][Release-Url]
[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/bogdanovich/siberite)

Siberite is a simple leveldb backed message queue server<br>
([twitter/kestrel](https://github.com/twitter/kestrel), [wavii/darner](https://github.com/wavii/darner) rewritten in Go).

Siberite is a very simple message queue server.  Unlike in-memory servers such as [redis](http://redis.io/), Siberite is
designed to handle queues much larger than what can be held in RAM.  And unlike enterprise queue servers such as
[RabbitMQ](http://www.rabbitmq.com/), Siberite keeps all messages **out of process**,
using [goleveldb](https://github.com/syndtr/goleveldb) as a persistent storage.

The result is a durable queue server that uses a small amount of in-resident memory regardless of queue size.

Siberite is based on Robey Pointer's [Kestrel](https://github.com/robey/kestrel) - simple, distributed message queue.
Like Kestrel, Siberite follows the "No talking! Shhh!" approach to distributed queues:
A single Siberite server has a set of queues identified by name.  Each queue is a strictly-ordered FIFO,
and querying from a fleet of Siberite servers provides a loosely-ordered queue.
Siberite also supports Kestrel's two-phase reliable fetch: if a client disconnects before confirming it handled
a message, the message will be handed to the next client.

Compared to Kestrel and Darner, Siberite is easier to build, maintain and distribute.
It uses an order of magnitude less memory compared to Kestrel, and has an ability
to consume queue multiple times (using durable cursors feature).

## Features

1. Siberite clients can consume single source queue multiple times using `get <queue>.<cursor_name>` syntax.

  - Usually, with `get <queue>` syntax, returned message gets expired and deleted from queue.
  - With cursor syntax `get <queue>.<cursor_name>`, a durable
    cursor gets initialized. It shifts forward with every read without deleting
    any messages in the source queue. Number of cursors per queue is not limited.
  - If you continue reads from the source queue with usual syntax again, siberite will continue
    deleting already serverd messages from the head of the queue. Any existing cursor that is
    internally points to an already expired message will start serving messages
    from the current queue head on the next read.
  - Durable cursors are also support two-phase reliable reads. All failed reliable
    reads for each cursor get stored in cursor's own small persistent queue and get
    served to other cursor readers.

2. Fanout queues

  - Siberite allows you to insert new message into multiple queues at once
    by using the following syntax `set <queue>+<another_queue>+<third_queue> ...`



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
