## Description
This is a memcache client package for the Go programming language.
The following commands are implemented:
* get (single key)
* set, add, replace, append, prepend
* delete
* incr, decr

## Installation
```
go get github.com/kklis/gomemcache
```
Depending on your environment configuration, you may need root (Linux) or administrator (Windows) access rights to run the above command.

## Testing
* Install gomemcache package (as described above).
* Start memcached at 127.0.0.1:11211 before running the test.
* On Unix start memcache socket listener: `memcached -s /tmp/memcached.sock -a 0755`
* Run command: `go test github.com/kklis/gomemcache`

**Warning**: Test suite includes a test that flushes all memcache content.

**Note**: On systems that don't support Unix sockets (like Microsoft Windows) TestDial_UNIX will fail.

## Example usage
* Go to $GOPATH/src/github.com/kklis/gomemcache/example/
* Compile example with: `go build example.go`
* Run the binary
