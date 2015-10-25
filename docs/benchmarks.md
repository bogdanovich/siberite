Benchmark Details:
* MacBook Pro CPU: 2.2 GHz Intel Core i7, RAM: 16 GB 1600 MHz DDR3, Disk: SSD
* OS X Yosemite 10.10.5
* Kestrel 2.4.8, Java 1.6.0_65, -Xmx1024m
* Darner 0.2.5 [Innometrics/darner](https://github.com/Innometrics/darner) built with RocksDB
* Siberite 0.4

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

siberite         0 requests: 3352 kB
siberite      1024 requests: 10508 kB
siberite      2048 requests: 12672 kB
siberite      4096 requests: 20472 kB
siberite      8192 requests: 23504 kB
siberite     16384 requests: 25908 kB
siberite     32768 requests: 29360 kB
siberite     65536 requests: 37652 kB
siberite    131072 requests: 48356 kB
siberite    262024 requests: 74724 kB
siberite    524048 requests: 74788 kB
```

# Queue Flooding

How quickly can we flood items through 10 queues?  This tests the raw throughput of the server.

![Queue Flood Benchmark](images/benchmark_queue_flood.png)

```
$ ./flood.sh
warming up kestrel...done.
kestrel      1 conns: 16807.879854 (ops/s mean)
kestrel      2 conns: 31177.408391
kestrel      5 conns: 45585.286014
kestrel     10 conns: 58226.020907
kestrel     50 conns: 61077.076049
kestrel    100 conns: 62672.071079
kestrel    200 conns: 61559.587967
kestrel    300 conns: 61625.581672
kestrel    400 conns: 61602.083759
kestrel    600 conns: 60304.539948
kestrel    800 conns: 60008.272035
kestrel   1000 conns: 59328.929184
kestrel   2000 conns: 36193.558029
kestrel   4000 conns: 33808.110691
kestrel   6000 conns: 15335.226914
kestrel   8000 conns: 15026.652097

darner       1 conns: 20485.164610
darner       2 conns: 39084.504255
darner       5 conns: 53805.902578
darner      10 conns: 56233.919391
darner      50 conns: 58144.865233
darner     100 conns: 54192.627376
darner     200 conns: 52301.009202
darner     300 conns: 53942.452802
darner     400 conns: 53321.372943
darner     600 conns: 52564.547239
darner     800 conns: 49931.330399
darner    1000 conns: 48939.109464
darner    2000 conns: 43240.382904
darner    4000 conns: 24467.149941
darner    6000 conns: 22490.386543
darner    8000 conns: 14474.205194

siberite       1 conns: 16685.785422
siberite       2 conns: 29680.072714
siberite       5 conns: 48004.527781
siberite      10 conns: 66177.472140
siberite      50 conns: 73205.980482
siberite     100 conns: 74767.933948
siberite     200 conns: 70195.272648
siberite     300 conns: 68949.940311
siberite     400 conns: 68624.760247
siberite     600 conns: 66170.498654
siberite     800 conns: 62391.969673
siberite    1000 conns: 60788.384543
siberite    2000 conns: 50584.645539
siberite    4000 conns: 27876.874830
siberite    6000 conns: 22633.264009
siberite    8000 conns: 19387.906547
```

# Queue Packing (1024 byte message size)

This tests the queue server's behavior with a backlog of items.  The challenge for the queue server is to serve items
that no longer all fit in memory.  Absolute throughput isn't important here - item sizes are large to quickly saturate
free memory.  Instead it's important for the throughput to flatten out as the backlog grows.

![Queue Packing Benchmark](images/benchmark_queue_packing_1024.png)


```
$ ./packing.sh
warming up kestrel...done.
kestrel        0 sets: 15052.481901
kestrel     1024 sets: 15525.517448
kestrel    16384 sets: 15377.189029
kestrel    65536 sets: 14683.953159
kestrel   262144 sets: 14147.473998
kestrel  1048576 sets: 14099.458784
kestrel  4194304 sets: 14893.911809
kestrel  8388608 sets: 14831.780153

darner        0 sets: 19459.351790
darner     1024 sets: 18821.834508
darner    16384 sets: 16667.949078
darner    65536 sets: 16206.286265
darner   262144 sets: 16551.859558
darner  1048576 sets: 15245.079659
darner  4194304 sets: 14875.396451
darner  8388608 sets: 14750.351526

siberite        0 sets: 16009.303237
siberite     1024 sets: 15615.363126
siberite    16384 sets: 14026.486300
siberite    65536 sets: 12975.689809
siberite   262144 sets: 11783.504995
siberite  1048576 sets: 10107.638889
siberite  4194304 sets: 10036.420823
siberite  8388608 sets: 9868.384511
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

siberite | items:          0 | speed:    16946 ops/s
siberite | items:       1024 | speed:    16875 ops/s
siberite | items:      17408 | speed:    16393 ops/s
siberite | items:      82944 | speed:    15279 ops/s
siberite | items:     345088 | speed:    15083 ops/s
siberite | items:    1393664 | speed:    14756 ops/s
siberite | items:    5587968 | speed:    14020 ops/s
siberite | items:   13976576 | speed:    14729 ops/s
siberite | items:   30753792 | speed:    14204 ops/s
siberite | items:   64308224 | speed:    14552 ops/s
siberite | items:  131417088 | speed:    12908 ops/s
siberite | items:  265634816 | speed:    13490 ops/s
siberite | items:  243498574 | speed:    14021 ops/s
siberite | items:  221362340 | speed:    14382 ops/s
siberite | items:  199226106 | speed:    14438 ops/s
siberite | items:  177089872 | speed:    13233 ops/s
siberite | items:  154953638 | speed:    14254 ops/s
siberite | items:  132817404 | speed:    14904 ops/s
siberite | items:  110681170 | speed:    14793 ops/s
siberite | items:   88544936 | speed:    14720 ops/s
siberite | items:   66408702 | speed:    14968 ops/s
siberite | items:   44272468 | speed:    14872 ops/s
siberite | items:   22136234 | speed:    14757 ops/s
siberite | items:          0 | speed:    16968 ops/s
```
