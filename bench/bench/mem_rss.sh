#!/bin/bash

# memory test - how much resident memory is used after growing and then shrinking a queue
# kestrel is on port 22133
# darner is on port 22134
# siberite is on port 22135
# before running this test, be sure to delete the db_bench queue and restart both services

for i in 0 1024 2048 4096 8192 16384 32768 65536 131072 262024 524048
do
   printf "kestrel  %7i requests: " "$i"
   ./bench -port 22133 -sets $i -gets 0 -item_size 1024 >/dev/null
   ./bench -port 22133 -sets 0 -gets $i -item_size 1024 >/dev/null
   printf "%s kB\n" $(ps -o rss= -p `pgrep -f kestrel`)
done

for i in 0 1024 2048 4096 8192 16384 32768 65536 131072 262024 524048
do
   printf "darner   %7i requests: " "$i"
   ./bench -port 22134 -sets $i -gets 0 -item_size 1024 >/dev/null
   ./bench -port 22134 -sets 0 -gets $i -item_size 1024 >/dev/null
   printf "%s kB\n" $(ps -o rss= -p `pgrep darner`)
done

for i in 0 1024 2048 4096 8192 16384 32768 65536 131072 262024 524048
do
   printf "siberite   %7i requests: " "$i"
   ./bench -port 22135 -sets $i -gets 0 -item_size 1024 >/dev/null
   ./bench -port 22135 -sets 0 -gets $i -item_size 1024 >/dev/null
   printf "%s kB\n" $(ps -o rss= -p `pgrep siberite`)
done
