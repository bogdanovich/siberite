#!/bin/bash

# flood test - try flood get/set at different concurrencies to measure throughput
# kestrel is on port 22133
# darner is on port 22134
# siberite is on port 22135

# kestrel reaches its best performance after a warmup period
# before running this test, be sure to delete the all queues and all services

NUM_OP=1000000

echo -n "warming up kestrel..."

./bench -port 22133 -sets 100000 -gets 100000 >/dev/null

echo "done."

# fill the queue with just a few items, to ensure there's always an item available
echo -ne "flush db_bench\r\n" | nc localhost 22133 >/dev/null

for i in 1 2 5 10 50 100 200 300 400 600 800 1000 2000 4000 6000 8000
do
   sync # just in case dirty pages are lying around, don't leak across each run
   printf "kestrel  %5i conns: " "$i"
   ./bench -port 22133 -sets $[$NUM_OP/$i] -gets $[$NUM_OP/$i] -queues 10 -concurrency $i | grep -i "Requests per second" | awk -F" " '{print $4}'
done

for i in 1 2 5 10 50 100 200 300 400 600 800 1000 2000 4000 6000 8000
do
   sync
   printf "darner   %5i conns: " "$i"
   ./bench -port 22134 -sets $[$NUM_OP/$i] -gets $[$NUM_OP/$i] -queues 10 -concurrency $i | grep -i "Requests per second" | awk -F" " '{print $4}'
done

for i in 1 2 5 10 50 100 200 300 400 600 800 1000 2000 4000 6000 8000
do
   sync
   printf "siberite   %5i conns: " "$i"
   ./bench -port 22135 -sets $[$NUM_OP/$i] -gets $[$NUM_OP/$i] -queues 10 -concurrency $i | grep -i "Requests per second" | awk -F" " '{print $4}'
done
