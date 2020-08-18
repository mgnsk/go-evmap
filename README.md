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
BenchmarkEvMap
BenchmarkEvMap/read
BenchmarkEvMap/read/single_writer_single_reader
BenchmarkEvMap/read/single_writer_single_reader-4               10677466               107 ns/op               1 B/op          0 allocs/op
BenchmarkEvMap/read/multi_writer_single_reader
BenchmarkEvMap/read/multi_writer_single_reader-4                 6955471               203 ns/op               0 B/op          0 allocs/op
BenchmarkEvMap/read/single_writer_multi_reader
BenchmarkEvMap/read/single_writer_multi_reader-4                28758618                44.8 ns/op             0 B/op          0 allocs/op
BenchmarkEvMap/read/multi_writer_multi_reader
BenchmarkEvMap/read/multi_writer_multi_reader-4                 13886259                96.0 ns/op             0 B/op          0 allocs/op
BenchmarkEvMap/write
BenchmarkEvMap/write/single_writer_single_reader
BenchmarkEvMap/write/single_writer_single_reader-4               1483501               826 ns/op               8 B/op          1 allocs/op
BenchmarkEvMap/write/multi_writer_single_reader
BenchmarkEvMap/write/multi_writer_single_reader-4                 564196              2782 ns/op               8 B/op          1 allocs/op
BenchmarkEvMap/write/single_writer_multi_reader
BenchmarkEvMap/write/single_writer_multi_reader-4                   4048           1503147 ns/op               9 B/op          1 allocs/op
BenchmarkEvMap/write/multi_writer_multi_reader
BenchmarkEvMap/write/multi_writer_multi_reader-4                     100          16217407 ns/op             207 B/op          1 allocs/op
BenchmarkMutexMap
BenchmarkMutexMap/read
BenchmarkMutexMap/read/single_writer_single_reader
BenchmarkMutexMap/read/single_writer_single_reader-4             1208575              1017 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMap/read/multi_writer_single_reader
BenchmarkMutexMap/read/multi_writer_single_reader-4               796638              1639 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMap/read/single_writer_multi_reader
BenchmarkMutexMap/read/single_writer_multi_reader-4              2735184               439 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMap/read/multi_writer_multi_reader
BenchmarkMutexMap/read/multi_writer_multi_reader-4               2428054               503 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMap/write
BenchmarkMutexMap/write/single_writer_single_reader
BenchmarkMutexMap/write/single_writer_single_reader-4            1213567               976 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMap/write/multi_writer_single_reader
BenchmarkMutexMap/write/multi_writer_single_reader-4              831890              1664 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMap/write/single_writer_multi_reader
BenchmarkMutexMap/write/single_writer_multi_reader-4              732358              1988 ns/op               0 B/op          0 allocs/op
BenchmarkMutexMap/write/multi_writer_multi_reader
BenchmarkMutexMap/write/multi_writer_multi_reader-4               538677              2481 ns/op               0 B/op          0 allocs/op
BenchmarkSyncMap
BenchmarkSyncMap/read
BenchmarkSyncMap/read/single_writer_single_reader
BenchmarkSyncMap/read/single_writer_single_reader-4             14777649                80.6 ns/op             5 B/op          0 allocs/op
BenchmarkSyncMap/read/multi_writer_single_reader
BenchmarkSyncMap/read/multi_writer_single_reader-4              10782256               110 ns/op              14 B/op          0 allocs/op
BenchmarkSyncMap/read/single_writer_multi_reader
BenchmarkSyncMap/read/single_writer_multi_reader-4              31052943                34.7 ns/op             1 B/op          0 allocs/op
BenchmarkSyncMap/read/multi_writer_multi_reader
BenchmarkSyncMap/read/multi_writer_multi_reader-4               24834422                48.3 ns/op             3 B/op          0 allocs/op
BenchmarkSyncMap/write
BenchmarkSyncMap/write/single_writer_single_reader
BenchmarkSyncMap/write/single_writer_single_reader-4             6737726               182 ns/op              16 B/op          1 allocs/op
BenchmarkSyncMap/write/multi_writer_single_reader
BenchmarkSyncMap/write/multi_writer_single_reader-4             14015390                81.8 ns/op            16 B/op          1 allocs/op
BenchmarkSyncMap/write/single_writer_multi_reader
BenchmarkSyncMap/write/single_writer_multi_reader-4              5226260               227 ns/op              16 B/op          1 allocs/op
BenchmarkSyncMap/write/multi_writer_multi_reader
BenchmarkSyncMap/write/multi_writer_multi_reader-4              10486099               110 ns/op              16 B/op          1 allocs/op
PASS
ok      github.com/mgnsk/evmap  43.262s
```
