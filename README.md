POC non-blocking read-write map. The tradeoff exposed here is 2x the memory usage due to swapping between two maps.

Inspired by: https://www.youtube.com/watch?v=s19G6n0UjsM

The first initial version of "just getting it to work". Can't beat `sync.Map` and write performance is slow.

Run benchmark with
`$ docker run --rm --mount type=bind,source="$(pwd)",target=/evmap -w /evmap golang:alpine go test -count=1 -v -run=. -bench=. ./...`

Passes the race detector when run as `go test -race -count=1 -v -run=. -bench=. ./...` albeit slowly.

```
=== RUN   TestMap
--- PASS: TestMap (0.00s)
goos: linux
goarch: amd64
pkg: github.com/mgnsk/evmap
BenchmarkEvMapReadSync
BenchmarkEvMapReadSync/read
BenchmarkEvMapReadSync/read/single_writer_single_reader
BenchmarkEvMapReadSync/read/single_writer_single_reader-4               10271283               126 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapReadSync/read/multi_writer_single_reader
BenchmarkEvMapReadSync/read/multi_writer_single_reader-4                 6374324               187 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapReadSync/read/single_writer_multi_reader
BenchmarkEvMapReadSync/read/single_writer_multi_reader-4                20645684                53.4 ns/op             0 B/op          0 allocs/op
BenchmarkEvMapReadSync/read/multi_writer_multi_reader
BenchmarkEvMapReadSync/read/multi_writer_multi_reader-4                 14648187                94.7 ns/op             0 B/op          0 allocs/op
BenchmarkEvMapRead100Millisecond
BenchmarkEvMapRead100Millisecond/read
BenchmarkEvMapRead100Millisecond/read/single_writer_single_reader
BenchmarkEvMapRead100Millisecond/read/single_writer_single_reader-4             14071902               112 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapRead100Millisecond/read/multi_writer_single_reader
BenchmarkEvMapRead100Millisecond/read/multi_writer_single_reader-4               7361605               172 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapRead100Millisecond/read/single_writer_multi_reader
BenchmarkEvMapRead100Millisecond/read/single_writer_multi_reader-4              22773284                54.2 ns/op             0 B/op          0 allocs/op
BenchmarkEvMapRead100Millisecond/read/multi_writer_multi_reader
BenchmarkEvMapRead100Millisecond/read/multi_writer_multi_reader-4               13505457                77.3 ns/op             0 B/op          0 allocs/op
BenchmarkMutexMapRead
BenchmarkMutexMapRead/read
BenchmarkMutexMapRead/read/single_writer_single_reader
BenchmarkMutexMapRead/read/single_writer_single_reader-4                         1206459               977 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapRead/read/multi_writer_single_reader
BenchmarkMutexMapRead/read/multi_writer_single_reader-4                           955950              2018 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapRead/read/single_writer_multi_reader
BenchmarkMutexMapRead/read/single_writer_multi_reader-4                          2999178               361 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapRead/read/multi_writer_multi_reader
BenchmarkMutexMapRead/read/multi_writer_multi_reader-4                           2717976               424 ns/op               0 B/op          0 allocs/op
BenchmarkSyncMapRead
BenchmarkSyncMapRead/read
BenchmarkSyncMapRead/read/single_writer_single_reader
BenchmarkSyncMapRead/read/single_writer_single_reader-4                         14560374                81.0 ns/op             5 B/op          0 allocs/op
BenchmarkSyncMapRead/read/multi_writer_single_reader
BenchmarkSyncMapRead/read/multi_writer_single_reader-4                          10915544               112 ns/op              14 B/op          0 allocs/op
BenchmarkSyncMapRead/read/single_writer_multi_reader
BenchmarkSyncMapRead/read/single_writer_multi_reader-4                          29487644                35.8 ns/op             1 B/op          0 allocs/op
BenchmarkSyncMapRead/read/multi_writer_multi_reader
BenchmarkSyncMapRead/read/multi_writer_multi_reader-4                           26232830                50.9 ns/op             4 B/op          0 allocs/op
BenchmarkEvMapWriteSync
BenchmarkEvMapWriteSync/write
BenchmarkEvMapWriteSync/write/single_writer_single_reader
BenchmarkEvMapWriteSync/write/single_writer_single_reader-4                      1526110               801 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWriteSync/write/multi_writer_single_reader
BenchmarkEvMapWriteSync/write/multi_writer_single_reader-4                        731452              1927 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWriteSync/write/single_writer_multi_reader
BenchmarkEvMapWriteSync/write/single_writer_multi_reader-4                          5060            772076 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWriteSync/write/multi_writer_multi_reader
BenchmarkEvMapWriteSync/write/multi_writer_multi_reader-4                            100          21816629 ns/op             191 B/op          0 allocs/op
BenchmarkEvMapWrite100Millisecond
BenchmarkEvMapWrite100Millisecond/write
BenchmarkEvMapWrite100Millisecond/write/single_writer_single_reader
BenchmarkEvMapWrite100Millisecond/write/single_writer_single_reader-4            4138698               276 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWrite100Millisecond/write/multi_writer_single_reader
BenchmarkEvMapWrite100Millisecond/write/multi_writer_single_reader-4             1765818               711 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWrite100Millisecond/write/single_writer_multi_reader
BenchmarkEvMapWrite100Millisecond/write/single_writer_multi_reader-4             2840790               424 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWrite100Millisecond/write/multi_writer_multi_reader
BenchmarkEvMapWrite100Millisecond/write/multi_writer_multi_reader-4              1296871               836 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapWrite
BenchmarkMutexMapWrite/write
BenchmarkMutexMapWrite/write/single_writer_single_reader
BenchmarkMutexMapWrite/write/single_writer_single_reader-4                       1214139              1155 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapWrite/write/multi_writer_single_reader
BenchmarkMutexMapWrite/write/multi_writer_single_reader-4                         596884              2366 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapWrite/write/single_writer_multi_reader
BenchmarkMutexMapWrite/write/single_writer_multi_reader-4                         765938              1977 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapWrite/write/multi_writer_multi_reader
BenchmarkMutexMapWrite/write/multi_writer_multi_reader-4                          588924              2254 ns/op               0 B/op          0 allocs/op
BenchmarkSyncMapWrite
BenchmarkSyncMapWrite/write
BenchmarkSyncMapWrite/write/single_writer_single_reader
BenchmarkSyncMapWrite/write/single_writer_single_reader-4                        6648818               180 ns/op              16 B/op          1 allocs/op
BenchmarkSyncMapWrite/write/multi_writer_single_reader
BenchmarkSyncMapWrite/write/multi_writer_single_reader-4                        13270246                83.6 ns/op            16 B/op          1 allocs/op
BenchmarkSyncMapWrite/write/single_writer_multi_reader
BenchmarkSyncMapWrite/write/single_writer_multi_reader-4                         5214732               274 ns/op              16 B/op          1 allocs/op
BenchmarkSyncMapWrite/write/multi_writer_multi_reader
BenchmarkSyncMapWrite/write/multi_writer_multi_reader-4                         11835210               107 ns/op              16 B/op          1 allocs/op
PASS
ok      github.com/mgnsk/evmap  55.018s
```
