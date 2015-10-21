#!/bin/bash

# packing test - check get/set throughput and disk space usage after having
#                grown the queue to different sizes

display_usage() {
  echo -e "\nUsage:\n$0 -name siberite -port 22133 -dir ./data  -item_size 64\n"
  echo "    -name siberite               software name for a nice output"
  echo "    -port 22133                  port to use for the test"
  echo "    -item_size  64               size of each queue item"
  echo "    -h                           display help"
}

options=$@
arguments=($options)
index=0
for argument in $options
  do
    index=`expr $index + 1`
    case $argument in
      -h) display_usage; exit;;
      -name) NAME=${arguments[index]} ;;
      -port) PORT=${arguments[index]} ;;
      -item_size) ITEM_SIZE=${arguments[index]} ;;
    esac
  done

if [ -z "$NAME" ]; then NAME="siberite"; fi
if [ -z "$PORT" ]; then PORT="22133"; fi
if [ -z "$ITEM_SIZE" ]; then  ITEM_SIZE="64"; fi

if [ $NAME = "kestrel" ]; then
  # kestrel reaches its best performance after a warmup period
  echo -n "warming up..."
  ./bench -port $PORT -sets 100000 -gets 100000 -item_size $ITEM_SIZE > /dev/null
  echo "done."
fi

echo -ne "flush db_bench\r\n" | nc localhost $PORT >/dev/null

sync

for i in 0 1024 16384 65536 262144 1048576 4194304 8388608 16777216 33554432 67108864 134217728
do
   ./bench -port $PORT -sets $i -item_size $ITEM_SIZE > /dev/null
   SIZE="$(echo -ne "stats\r\n" | nc localhost $PORT | grep queue_db_bench_items | awk -F" " '{print $3}' | tr -d ' \r')"
   SPEED="$(./bench -port $PORT -sets 200000 -gets 200000 -item_size $ITEM_SIZE | grep -i "Requests per second" | awk -F" " '{print $4}')"
   printf "%10s | items: %10s | speed: %8s ops/s\r\n" "$NAME" "$SIZE" "$SPEED"
done

# consuming the queue

for i in 22136242 22136234 22136234 22136234 22136234 22136234 22136234 22136234 22136234 22136234 22136234 22136234
do
   ./bench -port $PORT -gets $i -item_size $ITEM_SIZE > /dev/null
   SIZE="$(echo -ne "stats\r\n" | nc localhost $PORT | grep queue_db_bench_items | awk -F" " '{print $3}' | tr -d ' \r')"
   SPEED="$(./bench -port $PORT -sets 200000 -gets 200000 -item_size $ITEM_SIZE | grep -i "Requests per second" | awk -F" " '{print $4}')"
   printf "%10s | items: %10s | speed: %8s ops/s\r\n" "$NAME" "$SIZE" "$SPEED"
done


echo -ne "flush db_bench\r\n" | nc localhost $PORT >/dev/null

sync
