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
BenchmarkEvMapReadSync/read/single_writer_single_reader-4                8752372               127 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapReadSync/read/multi_writer_single_reader
BenchmarkEvMapReadSync/read/multi_writer_single_reader-4                 5005664               202 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapReadSync/read/single_writer_multi_reader
BenchmarkEvMapReadSync/read/single_writer_multi_reader-4                21856057                46.2 ns/op             0 B/op          0 allocs/op
BenchmarkEvMapReadSync/read/multi_writer_multi_reader
BenchmarkEvMapReadSync/read/multi_writer_multi_reader-4                 12506944               113 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapRead100Millisecond
BenchmarkEvMapRead100Millisecond/read
BenchmarkEvMapRead100Millisecond/read/single_writer_single_reader
BenchmarkEvMapRead100Millisecond/read/single_writer_single_reader-4             13871166               104 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapRead100Millisecond/read/multi_writer_single_reader
BenchmarkEvMapRead100Millisecond/read/multi_writer_single_reader-4               6026190               185 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapRead100Millisecond/read/single_writer_multi_reader
BenchmarkEvMapRead100Millisecond/read/single_writer_multi_reader-4              23838170                46.9 ns/op             0 B/op          0 allocs/op
BenchmarkEvMapRead100Millisecond/read/multi_writer_multi_reader
BenchmarkEvMapRead100Millisecond/read/multi_writer_multi_reader-4               18426666               110 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapRead
BenchmarkMutexMapRead/read
BenchmarkMutexMapRead/read/single_writer_single_reader
BenchmarkMutexMapRead/read/single_writer_single_reader-4                         1000000              2403 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapRead/read/multi_writer_single_reader
BenchmarkMutexMapRead/read/multi_writer_single_reader-4                           636568              4955 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapRead/read/single_writer_multi_reader
BenchmarkMutexMapRead/read/single_writer_multi_reader-4                          1388348              1140 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapRead/read/multi_writer_multi_reader
BenchmarkMutexMapRead/read/multi_writer_multi_reader-4                           1360434               999 ns/op               0 B/op          0 allocs/op
BenchmarkSyncMapRead
BenchmarkSyncMapRead/read
BenchmarkSyncMapRead/read/single_writer_single_reader
BenchmarkSyncMapRead/read/single_writer_single_reader-4                         14548926                83.0 ns/op             5 B/op          0 allocs/op
BenchmarkSyncMapRead/read/multi_writer_single_reader
BenchmarkSyncMapRead/read/multi_writer_single_reader-4                          10583112               113 ns/op              14 B/op          0 allocs/op
BenchmarkSyncMapRead/read/single_writer_multi_reader
BenchmarkSyncMapRead/read/single_writer_multi_reader-4                          30319982                38.1 ns/op             1 B/op          0 allocs/op
BenchmarkSyncMapRead/read/multi_writer_multi_reader
BenchmarkSyncMapRead/read/multi_writer_multi_reader-4                           27150357                56.0 ns/op             5 B/op          0 allocs/op
BenchmarkEvMapWriteSync
BenchmarkEvMapWriteSync/write
BenchmarkEvMapWriteSync/write/single_writer_single_reader
BenchmarkEvMapWriteSync/write/single_writer_single_reader-4                      1824427               649 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWriteSync/write/multi_writer_single_reader
BenchmarkEvMapWriteSync/write/multi_writer_single_reader-4                        763472              1620 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWriteSync/write/single_writer_multi_reader
BenchmarkEvMapWriteSync/write/single_writer_multi_reader-4                          3822           1930265 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWriteSync/write/multi_writer_multi_reader
BenchmarkEvMapWriteSync/write/multi_writer_multi_reader-4                            100          23744591 ns/op             196 B/op          0 allocs/op
BenchmarkEvMapWrite100Millisecond
BenchmarkEvMapWrite100Millisecond/write
BenchmarkEvMapWrite100Millisecond/write/single_writer_single_reader
BenchmarkEvMapWrite100Millisecond/write/single_writer_single_reader-4            4166439               262 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWrite100Millisecond/write/multi_writer_single_reader
BenchmarkEvMapWrite100Millisecond/write/multi_writer_single_reader-4             1546425               836 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWrite100Millisecond/write/single_writer_multi_reader
BenchmarkEvMapWrite100Millisecond/write/single_writer_multi_reader-4             2876132               417 ns/op               0 B/op          0 allocs/op
BenchmarkEvMapWrite100Millisecond/write/multi_writer_multi_reader
BenchmarkEvMapWrite100Millisecond/write/multi_writer_multi_reader-4              1238622              1054 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapWrite
BenchmarkMutexMapWrite/write
BenchmarkMutexMapWrite/write/single_writer_single_reader
BenchmarkMutexMapWrite/write/single_writer_single_reader-4                       1387116              3001 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapWrite/write/multi_writer_single_reader
BenchmarkMutexMapWrite/write/multi_writer_single_reader-4                         877587              5249 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapWrite/write/single_writer_multi_reader
BenchmarkMutexMapWrite/write/single_writer_multi_reader-4                         466520              4555 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMapWrite/write/multi_writer_multi_reader
BenchmarkMutexMapWrite/write/multi_writer_multi_reader-4                          275862              6844 ns/op               0 B/op          0 allocs/op
BenchmarkSyncMapWrite
BenchmarkSyncMapWrite/write
BenchmarkSyncMapWrite/write/single_writer_single_reader
BenchmarkSyncMapWrite/write/single_writer_single_reader-4                        6667801               181 ns/op              16 B/op          1 allocs/op
BenchmarkSyncMapWrite/write/multi_writer_single_reader
BenchmarkSyncMapWrite/write/multi_writer_single_reader-4                        12297567                82.8 ns/op            16 B/op          1 allocs/op
BenchmarkSyncMapWrite/write/single_writer_multi_reader
BenchmarkSyncMapWrite/write/single_writer_multi_reader-4                         4971781               230 ns/op              16 B/op          1 allocs/op
BenchmarkSyncMapWrite/write/multi_writer_multi_reader
BenchmarkSyncMapWrite/write/multi_writer_multi_reader-4                         10579861               109 ns/op              16 B/op          1 allocs/op
PASS
ok      github.com/mgnsk/evmap  67.431s
```
