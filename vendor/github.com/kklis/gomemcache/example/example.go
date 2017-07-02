package main

import (
	"fmt"
	"github.com/kklis/gomemcache"
)

func main() {
	// If you want to use this with UNIX domain socket, you can use like a following source code.
	// On a UNIX domain socket, port is 0.
	// mc, err := gomemcache.Connect("/path/to/memcached.sock", 0)
	memc, err := gomemcache.Connect("127.0.0.1", 11211)
	if err != nil {
		panic(err)
	}
	err = memc.Set("foo", []uint8("bar"), 0, 0)
	if err != nil {
		panic(err)
	}
	val, fl, _ := memc.Get("foo")
	fmt.Printf("%s %d\n", val, fl)
}
