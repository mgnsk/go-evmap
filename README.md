Run benchmark with
`$ docker run --rm --mount type=bind,source="$(pwd)",target=/evmap -w /evmap golang:alpine go test -count=1 -v -run=. -bench=. ./...`

Concurrent benchmark with 4 writers and 4 readers:

* 100 unique keys:
```
=== RUN   TestMap
--- PASS: TestMap (0.00s)
goos: linux
goarch: amd64
pkg: github.com/mgnsk/evmap
BenchmarkEvMap
BenchmarkEvMap-4        12962691               114 ns/op
BenchmarkMutexMap
BenchmarkMutexMap-4      3735372               960 ns/op
BenchmarkSyncMap
BenchmarkSyncMap-4       7551626               245 ns/op
PASS
ok      github.com/mgnsk/evmap  9.407s

=== RUN   TestMap
--- PASS: TestMap (0.00s)
goos: linux
goarch: amd64
pkg: github.com/mgnsk/evmap
BenchmarkEvMap
BenchmarkEvMap-4        11968924               128 ns/op
BenchmarkMutexMap
BenchmarkMutexMap-4      3362301               947 ns/op
BenchmarkSyncMap
BenchmarkSyncMap-4       7560937               182 ns/op
PASS
ok      github.com/mgnsk/evmap  9.623s
```

* 1000 unique keys:
```
=== RUN   TestMap
--- PASS: TestMap (0.00s)
goos: linux
goarch: amd64
pkg: github.com/mgnsk/evmap
BenchmarkEvMap
BenchmarkEvMap-4        11867253                94.1 ns/op
BenchmarkMutexMap
BenchmarkMutexMap-4      3542584               814 ns/op
BenchmarkSyncMap
BenchmarkSyncMap-4       6641637               179 ns/op
PASS
ok      github.com/mgnsk/evmap  9.858s
```
