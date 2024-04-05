## this is the HW for dcard

* performance comparison: 

1. first, only use default router
'wrk -t12 -c400 -d20s http://localhost:8080/api/v1/ad\?offset\=2\&limit\=3\&age\=35\&gender\=F\&country\=TW\&platform\=ios
Running 20s test @ http://localhost:8080/api/v1/ad?offset=2&limit=3&age=35&gender=F&country=TW&platform=ios
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   165.95ms  478.40ms   2.00s    90.78%
    Req/Sec    40.13     36.45   202.00     79.57%
  6950 requests in 20.09s, 1.54MB read
  Socket errors: connect 157, read 102, write 0, timeout 1454
Requests/sec:    345.99
Transfer/sec:     78.73KB'

2. second, use the middleware, router.Use(limit.MaxAllowed(10)) in main.go to constrain the number of concurrent requests handled at the same time: (In master branch, by default without any setting)

* wrk -t12 -c400 -d20s http://localhost:8080/api/v1/ad\?offset\=2\&limit\=3\&age\=35\&gender\=F\&country\=TW\&platform\=ios

> Running 20s test @ http://localhost:8080/api/v1/ad?offset=2&limit=3&age=35&gender=F&country=TW&platform=ios
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   353.85ms  678.40ms   2.00s    80.63%
    Req/Sec    48.95     46.74   380.00     82.43%
  8324 requests in 20.08s, 1.84MB read
  Socket errors: connect 157, read 102, write 0, timeout 766
  Non-2xx or 3xx responses: 49
Requests/sec:    414.53
Transfer/sec:     93.99KB'

3. use the background worker(redis):
- wrk -t10 -c1000 -d20s http://localhost:8080/api/v1/ad\?offset\=1\&limit\=3\&age\=35\&gender\=F\&country\=TW\&platform\=ios

> Running 20s test @ http://localhost:8080/api/v1/ad?offset=1&limit=3&age=35&gender=F&country=TW&platform=ios
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   202.83ms  107.89ms 550.44ms   66.42%
    Req/Sec   119.42     79.57   424.00     76.70%
  23639 requests in 20.11s, 4.90MB read
  Socket errors: connect 759, read 102, write 0, timeout 0
Requests/sec:   1175.75
Transfer/sec:    249.71KB'

- the other request: wrk -t10 -c1000 -d20s http://localhost:8081/api/v1/ad\?offset\=1\&limit\=3\&age\=35\&gender\=F\&country\=TW

> Running 20s test @ http://localhost:8081/api/v1/ad?offset=1&limit=3&age=35&gender=F&country=TW
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   285.45ms  204.20ms   1.34s    86.87%
    Req/Sec    86.91     66.85   393.00     66.51%
  16644 requests in 20.09s, 4.48MB read
  Socket errors: connect 759, read 120, write 0, timeout 0
Requests/sec:    828.36
Transfer/sec:    228.11KB