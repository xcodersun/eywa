Waterwheel ![travis](https://travis-ci.org/vivowares/waterwheel.svg?branch=master "build status")
==========

### *What is this?*

It's another logging package

### *Why do we need another logging package?*

This package is designed for heavy duty logging. It doesn't provide much config options and doesn't give you extra power of structured logging. The only goal here is to unblock the logging from the application runtime.

The standard `log` package is pretty good on performance, but it doesn't provide async logging, buffered logging, etc. There are also some comparisons among other logging packages in terms of performance. You can easily benchmark them. But you should find the result to be similar with [this post](https://www.reddit.com/r/golang/comments/3y4ag4/benchmarking_for_some_golang_logging_libraries/).

### *How does waterwheel then?*

Let's start from the root. Often times, logs are written into a file or a socket, etc. They all require synchronization between different writes. For example, when writing into a file, you need to make sure different goroutines are logging with locks. Failed to do locking will resulting unexpected data. It's usually OK when you only have a few logs per client and you don't have many concurrent clients. But when we started working on [Eywa](http://vivowares.github.io/eywa), we found that we want to support at least tens of thousands(or even close to million in the future release) of long-live connections per node, and each of them can generate tons of logs. When this happens, all the goroutines will be synchronized at the one single point - *Logging*. Not matter how much effort you put on improving other moving parts, logging will always block you if you don't make it asynchronous. That's why we created waterwheel and the only difference is that: Let the logging to have its own goroutine and, others can just fire and forget.

To show the difference, at the end of the documentation, we posted some benchmark results compared with standard `log` package.


### *How to use?*

To install the package:

```
  go get github.com/vivowares/waterwheel
```

Example of using:


```
  import "github.com/vivowares/waterwheel"

  var w io.WriteCloser                             // whatever io.WriteCloser
  bufSize := 512
  logger := waterwheel.NewAsyncLogger(
    waterwheel.NewBufferedWriteCloser(bufSize, w),
    waterwheel.SimpleFormatter, size, "debug",
  )
  logger.Info("test message")                      // INFO: [2009-11-10T23:00:00] test message

  logger.Close()                                   // don't forget to close the logger at the end.

```


### *Performance*

  **standard log**

```
  BenchmarkStandardLogSmallMessageToDiscard-8                                3000000         452 ns/op
  BenchmarkStandardLogLargeMessageToDiscard-8                                1000000        1417 ns/op

  BenchmarkStandardLogSmallMessageToFile-8                                   1000000        1751 ns/op
  BenchmarkStandardLogLargeMessageToFile-8                                    500000        4324 ns/op

```

  **waterwheel**

```
  BenchmarkAsyncLoggerSize1WithoutBufferedWriterSmallMessageToDiscard-8      1000000        1023 ns/op
  BenchmarkAsyncLoggerSize10WithoutBufferedWriterSmallMessageToDiscard-8     3000000         549 ns/op
  BenchmarkAsyncLoggerSize100WithoutBufferedWriterSmallMessageToDiscard-8    5000000         367 ns/op
  BenchmarkAsyncLoggerSize1000WithoutBufferedWriterSmallMessageToDiscard-8   5000000         339 ns/op
  BenchmarkAsyncLoggerSize10000WithoutBufferedWriterSmallMessageToDiscard-8  5000000         346 ns/op

  BenchmarkAsyncLoggerSize1WithoutBufferedWriterLargeMessageToDiscard-8      1000000        1047 ns/op
  BenchmarkAsyncLoggerSize10WithoutBufferedWriterLargeMessageToDiscard-8     3000000         577 ns/op
  BenchmarkAsyncLoggerSize100WithoutBufferedWriterLargeMessageToDiscard-8    3000000         449 ns/op
  BenchmarkAsyncLoggerSize1000WithoutBufferedWriterLargeMessageToDiscard-8   5000000         383 ns/op
  BenchmarkAsyncLoggerSize10000WithoutBufferedWriterLargeMessageToDiscard-8  5000000         401 ns/op

  BenchmarkAsyncLoggerSize1WithBufferedWriterSmallMessageToDiscard-8         1000000        1003 ns/op
  BenchmarkAsyncLoggerSize10WithBufferedWriterSmallMessageToDiscard-8        3000000         548 ns/op
  BenchmarkAsyncLoggerSize100WithBufferedWriterSmallMessageToDiscard-8       5000000         403 ns/op
  BenchmarkAsyncLoggerSize1000WithBufferedWriterSmallMessageToDiscard-8      5000000         358 ns/op
  BenchmarkAsyncLoggerSize10000WithBufferedWriterSmallMessageToDiscard-8     5000000         363 ns/op

  BenchmarkAsyncLoggerSize1WithBufferedWriterLargeMessageToDiscard-8         1000000        1107 ns/op
  BenchmarkAsyncLoggerSize10WithBufferedWriterLargeMessageToDiscard-8        2000000         649 ns/op
  BenchmarkAsyncLoggerSize100WithBufferedWriterLargeMessageToDiscard-8       3000000         508 ns/op
  BenchmarkAsyncLoggerSize1000WithBufferedWriterLargeMessageToDiscard-8      3000000         433 ns/op
  BenchmarkAsyncLoggerSize10000WithBufferedWriterLargeMessageToDiscard-8     3000000         442 ns/op

  BenchmarkAsyncLoggerSize1WithoutBufferedWriterSmallMessageToFile-8          500000        2367 ns/op
  BenchmarkAsyncLoggerSize10WithoutBufferedWriterSmallMessageToFile-8        1000000        1860 ns/op
  BenchmarkAsyncLoggerSize100WithoutBufferedWriterSmallMessageToFile-8       1000000        1508 ns/op
  BenchmarkAsyncLoggerSize1000WithoutBufferedWriterSmallMessageToFile-8      1000000        1509 ns/op
  BenchmarkAsyncLoggerSize10000WithoutBufferedWriterSmallMessageToFile-8     1000000        1526 ns/op

  BenchmarkAsyncLoggerSize1WithoutBufferedWriterLargeMessageToFile-8          300000        3454 ns/op
  BenchmarkAsyncLoggerSize10WithoutBufferedWriterLargeMessageToFile-8         500000        2825 ns/op
  BenchmarkAsyncLoggerSize100WithoutBufferedWriterLargeMessageToFile-8        500000        2443 ns/op
  BenchmarkAsyncLoggerSize1000WithoutBufferedWriterLargeMessageToFile-8       500000        2449 ns/op
  BenchmarkAsyncLoggerSize10000WithoutBufferedWriterLargeMessageToFile-8      500000        2458 ns/op

  BenchmarkAsyncLoggerSize1WithBufferedWriterSmallMessageToFile-8             500000        2546 ns/op
  BenchmarkAsyncLoggerSize10WithBufferedWriterSmallMessageToFile-8           2000000         755 ns/op
  BenchmarkAsyncLoggerSize100WithBufferedWriterSmallMessageToFile-8          3000000         514 ns/op
* BenchmarkAsyncLoggerSize1000WithBufferedWriterSmallMessageToFile-8         3000000         402 ns/op
  BenchmarkAsyncLoggerSize10000WithBufferedWriterSmallMessageToFile-8        3000000         404 ns/op

  BenchmarkAsyncLoggerSize1WithBufferedWriterLargeMessageToFile-8             500000        3564 ns/op
  BenchmarkAsyncLoggerSize10WithBufferedWriterLargeMessageToFile-8            500000        2177 ns/op
* BenchmarkAsyncLoggerSize100WithBufferedWriterLargeMessageToFile-8          1000000        1860 ns/op
  BenchmarkAsyncLoggerSize1000WithBufferedWriterLargeMessageToFile-8         1000000        1978 ns/op
  BenchmarkAsyncLoggerSize10000WithBufferedWriterLargeMessageToFile-8        1000000        2051 ns/op

  BenchmarkSyncLoggerWithoutBufferedWriterSmallMessageToDiscard-8            5000000         366 ns/op
  BenchmarkSyncLoggerWithoutBufferedWriterLargeMessageToDiscard-8            3000000         414 ns/op

  BenchmarkSyncLoggerSize1WithBufferedWriterSmallMessageToDiscard-8          3000000         407 ns/op
  BenchmarkSyncLoggerSize10WithBufferedWriterSmallMessageToDiscard-8         3000000         384 ns/op
  BenchmarkSyncLoggerSize100WithBufferedWriterSmallMessageToDiscard-8        5000000         396 ns/op
  BenchmarkSyncLoggerSize1000WithBufferedWriterSmallMessageToDiscard-8       5000000         376 ns/op
  BenchmarkSyncLoggerSize10000WithBufferedWriterSmallMessageToDiscard-8      5000000         378 ns/op

  BenchmarkSyncLoggerSize1WithBufferedWriterLargeMessageToDiscard-8          3000000         454 ns/op
  BenchmarkSyncLoggerSize10WithBufferedWriterLargeMessageToDiscard-8         3000000         462 ns/op
  BenchmarkSyncLoggerSize100WithBufferedWriterLargeMessageToDiscard-8        3000000         473 ns/op
  BenchmarkSyncLoggerSize1000WithBufferedWriterLargeMessageToDiscard-8       3000000         478 ns/op
  BenchmarkSyncLoggerSize10000WithBufferedWriterLargeMessageToDiscard-8      3000000         455 ns/op

  BenchmarkSyncLoggerWithoutBufferedWriterSmallMessageToFile-8               1000000        1626 ns/op
  BenchmarkSyncLoggerWithoutBufferedWriterLargeMessageToFile-8                500000        2934 ns/op

  BenchmarkSyncLoggerSize1WithBufferedWriterSmallMessageToFile-8             1000000        1815 ns/op
  BenchmarkSyncLoggerSize10WithBufferedWriterSmallMessageToFile-8            3000000         582 ns/op
  BenchmarkSyncLoggerSize100WithBufferedWriterSmallMessageToFile-8           3000000         512 ns/op
* BenchmarkSyncLoggerSize1000WithBufferedWriterSmallMessageToFile-8          3000000         455 ns/op
  BenchmarkSyncLoggerSize10000WithBufferedWriterSmallMessageToFile-8         3000000         446 ns/op

  BenchmarkSyncLoggerSize1WithBufferedWriterLargeMessageToFile-8              500000        2839 ns/op
  BenchmarkSyncLoggerSize10WithBufferedWriterLargeMessageToFile-8             500000        2409 ns/op
  BenchmarkSyncLoggerSize100WithBufferedWriterLargeMessageToFile-8            500000        2249 ns/op
* BenchmarkSyncLoggerSize1000WithBufferedWriterLargeMessageToFile-8           500000        2179 ns/op
  BenchmarkSyncLoggerSize10000WithBufferedWriterLargeMessageToFile-8         1000000        2188 ns/op
```

### *Conclusion:*

With `Async + Buffered` logging, writing to a regular file becomes even faster then standard logging to Discard!

