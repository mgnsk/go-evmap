Run benchmark with
`$ docker run --rm --mount type=bind,source="$(pwd)",target=/evmap -w /evmap golang:alpine go test -count=1 -v -run=. -bench=. ./...`
