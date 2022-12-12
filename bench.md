# Benchmark 

## Tree

### Existing

1. 10 routes
20109159               558.7 ns/op           846 B/op         14 allocs/op

2. 50 routes
11131761              1075 ns/op            2065 B/op         28 allocs/op

3. 100 routes
9724389              1231 ns/op            2261 B/op         33 allocs/op

4. 300 routes
6308816              1899 ns/op            3562 B/op         52 allocs/op

5. 500 routes
5763853              2082 ns/op            3835 B/op         56 allocs/op

### Not existing

1. 10 routes
34623898               333.8 ns/op           560 B/op         11 allocs/op

2. 50 routes
18032364               664.0 ns/op          1328 B/op         22 allocs/op

3. 100 routes
14049541               851.8 ns/op          1616 B/op         28 allocs/op

4. 300 routes
8112914              1479 ns/op            3296 B/op         43 allocs/op

5. 500 routes
7103973              1689 ns/op            3392 B/op         50 allocs/op
