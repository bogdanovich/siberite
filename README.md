# Siberite

[![Join the chat at https://gitter.im/bogdanovich/siberite](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/bogdanovich/siberite?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Siberite is a simple leveldb backed message queue server ([wavii/darner](https://github.com/wavii/darner) rewritten in Go).

## Build

Make sure your `GOROOT` and `GOPATH` is set correctly

```
git clone http://github.com/bogdanovich/siberite && cd siberite/main
go build siberite.go
mkdir ./data
./siberite -listen localhost:22133 -data ./data
2015/09/22 06:29:38 listening on 127.0.0.1:22134
2015/09/22 06:29:38 initializing...
2015/09/22 06:29:38 data directory:  ./data
```

## Usage

```
telnet localhost 22133
rying ::1...
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
STAT version siberite-0.2.1
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
# get work/abort
# flush work
# delete work
# flush_all
```


## Protocol

Siberite follows the same protocol as [Kestrel](http://github.com/robey/kestrel/blob/master/docs/guide.md#memcache), which is the memcache
protocol.

## Todo

  - close an exisitng item read and open a new one in one command GET work/close/open

## Not supported

  - waiting a given time limit for a new item to arrive /t=<milliseconds>
