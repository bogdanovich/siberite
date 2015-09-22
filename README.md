# Siberite

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

stats
STAT uptime 38
STAT time 1442903520
STAT version 0.1.1
STAT curr_connections 1
STAT total_connections 1
STAT cmd_get 0
STAT cmd_set 0
END

set my_queue_name 0 0 10
1234567890
STORED

set my_queue_name 0 0 2
12
STORED

get my_queue_name
VALUE my_queue_name 0 10
1234567890
END

get my_queue_name/open
VALUE my_queue_name 0 2
12
END

get my_queue_name/close
END

# other commands:
# get my_queue_name/peek
# get my_queue_name/open
# get my_queue_name/abort
# flush my_queue_name
# delete my_queue_name
# flush_all
```


## Protocol

Siberite follows the same protocol as [Kestrel](http://github.com/robey/kestrel/blob/master/docs/guide.md#memcache), which is the memcache
protocol.

## Todo

  - waiting a given time limit for a new item to arrive /t=<milliseconds>
  - waiting for an item and open it GET work/t=500/open
  - close an exisitng item read and open a new one in one command GET work/close/open

