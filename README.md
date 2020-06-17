# carlo

It is a hard fork of Monte, mainly for experimental learning by reconstruction, not for improvement.

## Further information
For the description and license of [monte](https://github.com/lithdew/monte), please check the [docs directory](https://github.com/TheSmallBoat/carlo/tree/master/docs).

Thanks to Kenta Iwasaki for his excellent work.

## Benchmarks

```
% sysctl -a | grep machdep.cpu | grep 'brand_'
machdep.cpu.brand_string: Intel(R) Core(TM) i5-7267U CPU @ 3.10GHz

% go test -bench=. -benchtime=10s
goos: darwin
goarch: amd64
pkg: github.com/TheSmallBoat/carlo/carlolib/carlolib
BenchmarkSend-4                 	  991206	     10430 ns/op	 134.23 MB/s	     115 B/op	       0 allocs/op
BenchmarkSendNoWait-4           	 7493818	      1569 ns/op	 892.30 MB/s	     512 B/op	       0 allocs/op
BenchmarkRequest-4              	  246334	     49136 ns/op	  28.49 MB/s	     140 B/op	       0 allocs/op
BenchmarkParallelSend-4         	 1632208	      8689 ns/op	 161.12 MB/s	     115 B/op	       0 allocs/op
BenchmarkParallelSendNoWait-4   	 9503029	      2364 ns/op	 592.31 MB/s	     652 B/op	       0 allocs/op
BenchmarkParallelRequest-4      	  668157	     16349 ns/op	  85.63 MB/s	     140 B/op	       0 allocs/op
```
