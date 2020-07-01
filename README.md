# carlo

The bare minimum for the high performance, bidirectional fully-encrypted streaming-transmit and streaming-RPC in Go with zero memory allocations.

It is a hard fork of [monte](https://github.com/lithdew/monte), mainly for experimental learning by reconstruction, not for improvement.

## Further information
Thanks to Kenta Iwasaki for his excellent work.

## Benchmarks

```
% sysctl -a | grep machdep.cpu | grep 'brand_'
machdep.cpu.brand_string: Intel(R) Core(TM) i5-7267U CPU @ 3.10GHz

% go test -bench=. -benchtime=10s
goos: darwin
goarch: amd64
pkg: github.com/TheSmallBoat/carlo/streaming_transmit
BenchmarkSend-4                          1179807             10745 ns/op         130.29 MB/s         115 B/op          0 allocs/op
BenchmarkSendNoWait-4                    6526405              1998 ns/op         700.73 MB/s         504 B/op          0 allocs/op
BenchmarkRequest-4                        211022             52414 ns/op          26.71 MB/s         140 B/op          0 allocs/op
BenchmarkParallelSend-4                  1854140              6488 ns/op         215.79 MB/s         115 B/op          0 allocs/op
BenchmarkParallelSendNoWait-4            7892371              2516 ns/op         556.51 MB/s         602 B/op          0 allocs/op
BenchmarkParallelRequest-4                640964             16536 ns/op          84.66 MB/s         140 B/op          0 allocs/op
PASS
ok      github.com/TheSmallBoat/carlo/streaming_transmit        110.942s
```
