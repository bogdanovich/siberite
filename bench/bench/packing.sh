#!/bin/bash

# packing test - check get/set throughput after having grown the queue to different sizes
# kestrel is on port 22133
# darner is on port 22134
# siberite is on port 22135

# kestrel reaches its best performance after a warmup period

echo -n "warming up kestrel..."

./bench -port 22133 -sets 100000 -gets 100000 > /dev/null

echo "done."

echo -ne "flush db_bench\r\n" | nc localhost 22133 >/dev/null

sync # don't leak across benchmarks

for i in 0 1024 16384 65536 262144 1048576 4194304 8388608
do
   ./bench -port 22133 -sets $i -gets 0 -item_size 1024 > /dev/null
   printf "kestrel %8i sets: " "$i"
   ./bench -port 22133 -sets 100000 -gets 100000 -item_size 1024 | grep -i "Requests per second" | awk -F" " '{print $4}'
done

# cleanup disk space after test
echo -ne "flush db_bench\r\n" | nc localhost 22133 >/dev/null

sync

for i in 0 1024 16384 65536 262144 1048576 4194304 8388608
do
   ./bench -port 22134 -sets $i -gets 0 -item_size 1024 > /dev/null
   printf "darner %8i sets: " "$i"
   ./bench -port 22134 -sets 100000 -gets 100000 -item_size 1024 | grep -i "Requests per second" | awk -F" " '{print $4}'
done

echo -ne "flush db_bench\r\n" | nc localhost 22134 >/dev/null


sync

for i in 0 1024 16384 65536 262144 1048576 4194304 8388608
do
   ./bench -port 22135 -sets $i -gets 0 -item_size 1024 > /dev/null
   printf "siberite %8i sets: " "$i"
   ./bench -port 22135 -sets 100000 -gets 100000 -item_size 1024 | grep -i "Requests per second" | awk -F" " '{print $4}'
done

echo -ne "flush db_bench\r\n" | nc localhost 22135 >/dev/null
