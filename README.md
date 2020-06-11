Non-blocking read-write map. The tradeoff exposed here is 2x the memory usage due to swapping between two maps.

The first initial version of "just getting it to work". Can't beat `sync.Map` and write performance is slow.

Run benchmark with
`$ docker run --rm --mount type=bind,source="$(pwd)",target=/evmap -w /evmap golang:alpine go test -count=1 -v -run=. -bench=. ./...`

```
=== RUN   TestMap
--- PASS: TestMap (0.00s)
goos: linux
goarch: amd64
pkg: github.com/mgnsk/evmap
BenchmarkEvMap
BenchmarkEvMap/read
BenchmarkEvMap/read/single_writer_single_reader
BenchmarkEvMap/read/single_writer_single_reader-4                8476042               134 ns/op
BenchmarkEvMap/read/multi_writer_single_reader
BenchmarkEvMap/read/multi_writer_single_reader-4                 8610511               138 ns/op
BenchmarkEvMap/read/single_writer_multi_reader
BenchmarkEvMap/read/single_writer_multi_reader-4                19326206                56.2 ns/op
BenchmarkEvMap/read/multi_writer_multi_reader
BenchmarkEvMap/read/multi_writer_multi_reader-4                 19056669                59.4 ns/op
BenchmarkEvMap/write
BenchmarkEvMap/write/single_writer_single_reader
BenchmarkEvMap/write/single_writer_single_reader-4               1332579               913 ns/op
BenchmarkEvMap/write/multi_writer_single_reader
BenchmarkEvMap/write/multi_writer_single_reader-4                 697406              1571 ns/op
BenchmarkEvMap/write/single_writer_multi_reader
BenchmarkEvMap/write/single_writer_multi_reader-4                    753           2616417 ns/op
BenchmarkEvMap/write/multi_writer_multi_reader
BenchmarkEvMap/write/multi_writer_multi_reader-4                     542           4373640 ns/op
BenchmarkMutexMap
BenchmarkMutexMap/read
BenchmarkMutexMap/read/single_writer_single_reader
BenchmarkMutexMap/read/single_writer_single_reader-4             1000000              1204 ns/op
BenchmarkMutexMap/read/multi_writer_single_reader
BenchmarkMutexMap/read/multi_writer_single_reader-4               830566              1850 ns/op
BenchmarkMutexMap/read/single_writer_multi_reader
BenchmarkMutexMap/read/single_writer_multi_reader-4              2619073               470 ns/op
BenchmarkMutexMap/read/multi_writer_multi_reader
BenchmarkMutexMap/read/multi_writer_multi_reader-4               2251362               540 ns/op
BenchmarkMutexMap/write
BenchmarkMutexMap/write/single_writer_single_reader
BenchmarkMutexMap/write/single_writer_single_reader-4            1000000              1170 ns/op
BenchmarkMutexMap/write/multi_writer_single_reader
BenchmarkMutexMap/write/multi_writer_single_reader-4              763395              1902 ns/op
BenchmarkMutexMap/write/single_writer_multi_reader
BenchmarkMutexMap/write/single_writer_multi_reader-4              751196              2091 ns/op
BenchmarkMutexMap/write/multi_writer_multi_reader
BenchmarkMutexMap/write/multi_writer_multi_reader-4               563013              2627 ns/op
BenchmarkSyncMap
BenchmarkSyncMap/read
BenchmarkSyncMap/read/single_writer_single_reader
BenchmarkSyncMap/read/single_writer_single_reader-4             14948563               103 ns/op
BenchmarkSyncMap/read/multi_writer_single_reader
BenchmarkSyncMap/read/multi_writer_single_reader-4               9950035               112 ns/op
BenchmarkSyncMap/read/single_writer_multi_reader
BenchmarkSyncMap/read/single_writer_multi_reader-4              35429257                37.5 ns/op
BenchmarkSyncMap/read/multi_writer_multi_reader
BenchmarkSyncMap/read/multi_writer_multi_reader-4               25953297                47.7 ns/op
BenchmarkSyncMap/write
BenchmarkSyncMap/write/single_writer_single_reader
BenchmarkSyncMap/write/single_writer_single_reader-4             5791599               207 ns/op
BenchmarkSyncMap/write/multi_writer_single_reader
BenchmarkSyncMap/write/multi_writer_single_reader-4             12453050                95.0 ns/op
BenchmarkSyncMap/write/single_writer_multi_reader
BenchmarkSyncMap/write/single_writer_multi_reader-4              4573100               258 ns/op
BenchmarkSyncMap/write/multi_writer_multi_reader
BenchmarkSyncMap/write/multi_writer_multi_reader-4               9819717               126 ns/op
PASS
ok      github.com/mgnsk/evmap  35.968s
```
