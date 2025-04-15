# EventBus

[![Go Report Card](https://goreportcard.com/badge/github.com/surasithaof/eventbus)](https://goreportcard.com/report/github.com/surasithaof/eventbus)
[![Go Reference](https://pkg.go.dev/badge/github.com/surasithaof/eventbus.svg)](https://pkg.go.dev/github.com/surasithaof/eventbus)
[![license](https://img.shields.io/github/license/surasithaof/eventbus)](https://github.com/surasithaof/eventbus/blob/master/LICENSE.md)

A simple Go package for implementing an internal async event bus. This is for learning Goroutine to handle asynchronous event publishing and subscription, making it efficient and easy to integrate with Go apps.

## Benchmark

```sh
go test -benchtime 10000000x -benchmem -run=^$ -bench=.
```

```
goos: darwin
goarch: arm64
pkg: github.com/surasithaof/eventbus
cpu: Apple M1
BenchmarkSubscribe-8    10000000                53.07 ns/op           42 B/op          0 allocs/op
BenchmarkPublish-8      10000000               346.5 ns/op            96 B/op          2 allocs/op
PASS
ok      github.com/surasithaof/eventbus 4.221s
```
