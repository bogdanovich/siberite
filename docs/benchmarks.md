Benchmark Details:
* MacBook Pro CPU: 2.2 GHz Intel Core i7, RAM: 16 GB 1600 MHz DDR3, Disk: SSD
* OS X Yosemite 10.10.5
* Kestrel 2.4.8, Java 1.6.0_65, -Xmx1024m
* Darner 0.2.5 [Innometrics/darner](https://github.com/Innometrics/darner) built with RocksDB
* Siberite 0.4.2

# Resident Memory

How much memory does the queue server use?  We are testing both steady-state memory resident, and also how aggressively
the server acquires and releases memory as queues expand and contract.
Kestrel memory settings: `-Xmx1024m`.

![Resident Memory Benchmark](images/benchmark_resident_memory.png)

```
$ ./mem_rss.sh
kestrel        0 requests: 168348 kB
kestrel     1024 requests: 198680 kB
kestrel     2048 requests: 217764 kB
kestrel     4096 requests: 246204 kB
kestrel     8192 requests: 240440 kB
kestrel    16384 requests: 255976 kB
kestrel    32768 requests: 295148 kB
kestrel    65536 requests: 321204 kB
kestrel   131072 requests: 459004 kB
kestrel   262024 requests: 775740 kB
kestrel   524048 requests: 833664 kB

darner         0 requests: 2832 kB
darner      1024 requests: 4632 kB
darner      2048 requests: 6868 kB
darner      4096 requests: 9140 kB
darner      8192 requests: 17296 kB
darner     16384 requests: 25040 kB
darner     32768 requests: 46352 kB
darner     65536 requests: 47584 kB
darner    131072 requests: 49060 kB
darner    262024 requests: 50764 kB
darner    524048 requests: 54112 kB

siberite         0 requests: 2420 kB
siberite      1024 requests: 10084 kB
siberite      2048 requests: 12324 kB
siberite      4096 requests: 20064 kB
siberite      8192 requests: 36932 kB
siberite     16384 requests: 45400 kB
siberite     32768 requests: 50612 kB
siberite     65536 requests: 58412 kB
siberite    131072 requests: 65208 kB
siberite    262024 requests: 81800 kB
siberite    524048 requests: 87360 kB
```

# Queue Flooding

How quickly can we flood items through 10 queues?  This tests the raw throughput of the server.

![Queue Flood Benchmark](images/benchmark_queue_flood.png)

```
$ ./flood.sh
warming up kestrel...done.
kestrel      1 conns: 16375 (requests/s mean)
kestrel      2 conns: 30039
kestrel      5 conns: 45235
kestrel     10 conns: 55656
kestrel     50 conns: 58945
kestrel    100 conns: 59103
kestrel    200 conns: 58564
kestrel    300 conns: 57807
kestrel    400 conns: 57621
kestrel    600 conns: 57065
kestrel    800 conns: 57273
kestrel   1000 conns: 56685
kestrel   2000 conns: 44625
kestrel   4000 conns: 37366
kestrel   6000 conns: 13404
kestrel   8000 conns: 18523

darner       1 conns: 20375
darner       2 conns: 38828
darner       5 conns: 51813
darner      10 conns: 54161
darner      50 conns: 54177
darner     100 conns: 53461
darner     200 conns: 52483
darner     300 conns: 51443
darner     400 conns: 51400
darner     600 conns: 51803
darner     800 conns: 48456
darner    1000 conns: 48253
darner    2000 conns: 42728
darner    4000 conns: 26062
darner    6000 conns: 23156
darner    8000 conns: 15040

siberite       1 conns: 16982
siberite       2 conns: 29770
siberite       5 conns: 48896
siberite      10 conns: 65952
siberite      50 conns: 72798
siberite     100 conns: 74385
siberite     200 conns: 69777
siberite     300 conns: 68507
siberite     400 conns: 67018
siberite     600 conns: 64851
siberite     800 conns: 63509
siberite    1000 conns: 61003
siberite    2000 conns: 48766
siberite    4000 conns: 25512
siberite    6000 conns: 21621
siberite    8000 conns: 15876
```

# Queue Packing (1024 byte message size)

This tests the queue server's behavior with a backlog of items.  The challenge for the queue server is to serve items
that no longer all fit in memory.  Absolute throughput isn't important here - item sizes are large to quickly saturate
free memory.  Instead it's important for the throughput to flatten out as the backlog grows.

![Queue Packing Benchmark](images/benchmark_queue_packing_1024.png)


```
$ ./packing.sh
warming up kestrel...done.
kestrel        0 sets: 15052
kestrel     1024 sets: 15525
kestrel    16384 sets: 15377
kestrel    65536 sets: 14683
kestrel   262144 sets: 14147
kestrel  1048576 sets: 14099
kestrel  4194304 sets: 14893
kestrel  8388608 sets: 14831

darner        0 sets: 19459
darner     1024 sets: 18821
darner    16384 sets: 16667
darner    65536 sets: 16206
darner   262144 sets: 16551
darner  1048576 sets: 15245
darner  4194304 sets: 14875
darner  8388608 sets: 14750

siberite        0 sets: 15466
siberite     1024 sets: 15583
siberite    16384 sets: 14077
siberite    65536 sets: 12898
siberite   262144 sets: 12180
siberite  1048576 sets: 11310
siberite  4194304 sets: 11287
siberite  8388608 sets: 11373
```

# Queue Packing and Unpacking (64 byte message size)

The challenge for the queue server is to serve items that no longer all fit
in memory. And to make sure that leveldb performance doesn't degrade because of
large number of delete queries.

![Queue Packing_Benchmark](images/benchmark_queue_packing_64.png)

![Queue Unpacking Benchmark](images/benchmark_queue_unpacking_64.png)


```
kestrel | items:          0 | speed:    16927 ops/s
kestrel | items:       1024 | speed:    16953 ops/s
kestrel | items:      17408 | speed:    16717 ops/s
kestrel | items:      82944 | speed:    16560 ops/s
kestrel | items:     345088 | speed:    16492 ops/s
kestrel | items:    1393664 | speed:    16731 ops/s
kestrel | items:    5587968 | speed:    15302 ops/s
kestrel | items:   13976576 | speed:    16144 ops/s
kestrel | items:   30753792 | speed:    15779 ops/s
kestrel | items:   64308224 | speed:    14888 ops/s
kestrel | items:  131417088 | speed:    16094 ops/s
kestrel | items:  265634816 | speed:    16846 ops/s
kestrel | items:  243498574 | speed:    16907 ops/s
kestrel | items:  221362340 | speed:    16874 ops/s
kestrel | items:  199226106 | speed:    16956 ops/s
kestrel | items:  177089872 | speed:    16826 ops/s
kestrel | items:  154953638 | speed:    16819 ops/s
kestrel | items:  132817404 | speed:    16924 ops/s
kestrel | items:  110681170 | speed:    16942 ops/s
kestrel | items:   88544936 | speed:    16940 ops/s
kestrel | items:   66408702 | speed:    16953 ops/s
kestrel | items:   44272468 | speed:    16159 ops/s
kestrel | items:   22136234 | speed:    16928 ops/s
kestrel | items:          0 | speed:    18098 ops/s

darner | items:          0 | speed:    20223 ops/s
darner | items:       1024 | speed:    19658 ops/s
darner | items:      17408 | speed:    18686 ops/s
darner | items:      82944 | speed:    17521 ops/s
darner | items:     345088 | speed:    17248 ops/s
darner | items:    1393664 | speed:    16978 ops/s
darner | items:    5587968 | speed:    16299 ops/s
darner | items:   13976576 | speed:    17190 ops/s
darner | items:   30753792 | speed:    15707 ops/s
darner | items:   64308224 | speed:    17279 ops/s
darner | items:  131417088 | speed:    16091 ops/s
darner | items:  265634816 | speed:    16080 ops/s
darner | items:  243498574 | speed:    17390 ops/s
darner | items:  221362340 | speed:    17705 ops/s
darner | items:  199226106 | speed:    17944 ops/s
darner | items:  177089872 | speed:    16877 ops/s
darner | items:  154953638 | speed:    17661 ops/s
darner | items:  132817404 | speed:    16966 ops/s
darner | items:  110681170 | speed:    17402 ops/s
darner | items:   88544936 | speed:    17008 ops/s
darner | items:   66408702 | speed:    16710 ops/s
darner | items:   44272468 | speed:    18028 ops/s
darner | items:   22136234 | speed:    17244 ops/s
darner | items:          0 | speed:    20864 ops/s

siberite | items:          0 | speed:    14768 ops/s
siberite | items:       1024 | speed:    15391 ops/s
siberite | items:      17408 | speed:    15076 ops/s
siberite | items:      82944 | speed:    15687 ops/s
siberite | items:     345088 | speed:    15776 ops/s
siberite | items:    1393664 | speed:    14984 ops/s
siberite | items:    5587968 | speed:    15207 ops/s
siberite | items:   13976576 | speed:    15121 ops/s
siberite | items:   30753792 | speed:    14596 ops/s
siberite | items:   64308224 | speed:    15309 ops/s
siberite | items:  131417088 | speed:    13376 ops/s
siberite | items:  265634816 | speed:    14628 ops/s
siberite | items:  243498574 | speed:    14704 ops/s
siberite | items:  221362340 | speed:    15243 ops/s
siberite | items:  199226106 | speed:    15157 ops/s
siberite | items:  177089872 | speed:    15086 ops/s
siberite | items:  154953638 | speed:    15086 ops/s
siberite | items:  132817404 | speed:    14033 ops/s
siberite | items:  110681170 | speed:    15070 ops/s
siberite | items:   88544936 | speed:    14226 ops/s
siberite | items:   66408702 | speed:    14343 ops/s
siberite | items:   44272468 | speed:    14517 ops/s
siberite | items:   22136234 | speed:    13788 ops/s
siberite | items:          0 | speed:    14602 ops/s
```
