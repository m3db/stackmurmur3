# stackmurmur3

#### Note: This is a fork of [github.com/spaolacci/murmur3](http://github.com/spaolacci/murmur3) (for v2, [twmb/murmur3](http://github.com/twmb/murmur3)), that provides digests that are allocated on the stack and can be incrementally written to. This is useful for places where you perform concurrent hashing and there's no good place to cache a hash without needing to acquire it expensively (under lock, etc).

[![Build Status](https://travis-ci.org/m3db/stackmurmur3.svg?branch=master)](https://travis-ci.org/m3db/stackmurmur3)

Native Go implementation of Austin Appleby's third MurmurHash revision (aka MurmurHash3).

Reference algorithm has been slightly hacked as to support the streaming mode required by Go's standard [Hash interface](http://golang.org/pkg/hash/#Hash).

## Benchmarks

### V2 benchmarks
The [twmb/murmur3](http://github.com/twmb/murmur3)-based version is significantly faster than V1 for small payloads.
It is, however, slightly slower for large payloads, as it does not use `unsafe` pointer conversions that are not always safe, and is portable across architectures.

Old is `twmb/murmur3` incremental hash, new is zero-alloc `v2/stack` digest:
<pre>
name                     old speed      new speed       delta
Partial128Sizes/32-12     494MB/s ± 2%   1478MB/s ± 1%  +199.52%  (p=0.008 n=5+5)
Partial128Sizes/64-12     932MB/s ± 0%   2533MB/s ± 1%  +171.93%  (p=0.008 n=5+5)
Partial128Sizes/128-12   1.65GB/s ± 0%   3.85GB/s ± 2%  +134.06%  (p=0.008 n=5+5)
Partial128Sizes/256-12   2.68GB/s ± 1%   5.18GB/s ± 0%   +93.44%  (p=0.008 n=5+5)
Partial128Sizes/512-12   3.90GB/s ± 1%   6.10GB/s ± 2%   +56.23%  (p=0.008 n=5+5)
Partial128Sizes/1024-12  5.08GB/s ± 0%   6.48GB/s ± 2%   +27.57%  (p=0.008 n=5+5)
Partial128Sizes/2048-12  5.86GB/s ± 1%   6.94GB/s ± 1%   +18.37%  (p=0.008 n=5+5)
Partial128Sizes/4096-12  6.52GB/s ± 0%   7.14GB/s ± 1%    +9.61%  (p=0.008 n=5+5)
Partial128Sizes/8192-12  6.91GB/s ± 1%   7.24GB/s ± 2%    +4.77%  (p=0.008 n=5+5)
Partial32Branches/1-12   17.7MB/s ± 1%   51.9MB/s ± 1%  +192.50%  (p=0.008 n=5+5)
Partial32Branches/2-12   34.9MB/s ± 2%  100.4MB/s ± 1%  +187.98%  (p=0.008 n=5+5)
Partial32Branches/3-12   52.2MB/s ± 1%  149.3MB/s ± 1%  +185.86%  (p=0.008 n=5+5)
Partial32Branches/4-12   70.4MB/s ± 1%  270.0MB/s ± 0%  +283.65%  (p=0.008 n=5+5)
Partial32Sizes/32-12      482MB/s ± 0%   1460MB/s ± 2%  +203.14%  (p=0.008 n=5+5)
Partial32Sizes/64-12      837MB/s ± 1%   2096MB/s ± 2%  +150.47%  (p=0.008 n=5+5)
Partial32Sizes/128-12    1.33GB/s ± 1%   2.58GB/s ± 0%   +94.60%  (p=0.008 n=5+5)
Partial32Sizes/256-12    1.79GB/s ± 0%   2.86GB/s ± 3%   +60.17%  (p=0.016 n=4+5)
Partial32Sizes/512-12    2.20GB/s ± 1%   2.94GB/s ± 1%   +33.74%  (p=0.008 n=5+5)
Partial32Sizes/1024-12   2.58GB/s ± 2%   3.09GB/s ± 1%   +19.65%  (p=0.008 n=5+5)
Partial32Sizes/2048-12   2.88GB/s ± 1%   3.15GB/s ± 1%    +9.66%  (p=0.008 n=5+5)
Partial32Sizes/4096-12   3.02GB/s ± 0%   3.20GB/s ± 1%    +6.02%  (p=0.016 n=4+5)
Partial32Sizes/8192-12   3.10GB/s ± 1%   3.23GB/s ± 1%    +3.97%  (p=0.008 n=5+5)

name                     old alloc/op   new alloc/op    delta
Partial128Sizes/32-12       96.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial128Sizes/64-12       96.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial128Sizes/128-12      96.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial128Sizes/256-12      96.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial128Sizes/512-12      96.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial128Sizes/1024-12     96.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial128Sizes/2048-12     96.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial128Sizes/4096-12     96.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial128Sizes/8192-12     96.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Branches/0-12      80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Branches/1-12      80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Branches/2-12      80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Branches/3-12      80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Branches/4-12      80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Sizes/32-12        80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Sizes/64-12        80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Sizes/128-12       80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Sizes/256-12       80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Sizes/512-12       80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Sizes/1024-12      80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Sizes/2048-12      80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Sizes/4096-12      80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
Partial32Sizes/8192-12      80.0B ± 0%       0.0B       -100.00%  (p=0.008 n=5+5)
</pre>

Old values are for `m3db/stackmurmur3`, new - `m3db/stackmurmur3/v2/stackmurmur3`.
Note how big payloads are actually slower, since it's not using unsafe code in hot path.
<pre>
name                    old speed      new speed      delta
Incremental128_1-12      219MB/s ± 3%   463MB/s ± 2%  +111.63%  (p=0.000 n=30+29)
Incremental128_2-12      198MB/s ± 2%   454MB/s ± 3%  +129.76%  (p=0.000 n=29+29)
Incremental128_4-12      228MB/s ± 2%   503MB/s ± 2%  +120.85%  (p=0.000 n=29+29)
Incremental128_8-12      280MB/s ± 2%   650MB/s ± 1%  +132.27%  (p=0.000 n=30+27)
Incremental128_16-12     531MB/s ± 2%   958MB/s ± 2%   +80.46%  (p=0.000 n=30+28)
Incremental128_32-12     805MB/s ± 2%  1371MB/s ± 2%   +70.39%  (p=0.000 n=30+30)
Incremental128_64-12    1.27GB/s ± 2%  2.07GB/s ± 2%   +62.95%  (p=0.000 n=30+28)
Incremental128_128-12   2.05GB/s ± 2%  3.03GB/s ± 3%   +47.84%  (p=0.000 n=29+30)
Incremental128_256-12   3.14GB/s ± 3%  4.22GB/s ± 2%   +34.58%  (p=0.000 n=30+28)
Incremental128_512-12   4.43GB/s ± 2%  5.31GB/s ± 3%   +19.95%  (p=0.000 n=29+30)
Incremental128_1024-12  5.62GB/s ± 3%  6.15GB/s ± 3%    +9.40%  (p=0.000 n=30+30)
Incremental128_2048-12  6.45GB/s ± 3%  6.55GB/s ± 1%    +1.60%  (p=0.000 n=30+30)
Incremental128_4096-12  7.02GB/s ± 2%  6.89GB/s ± 3%    -1.85%  (p=0.000 n=30+29)
Incremental128_8192-12  7.33GB/s ± 2%  7.04GB/s ± 3%    -3.96%  (p=0.000 n=29+30)
</pre>

### V1 benchmarks
This shows that its really only useful to use this stack version when allocating per operation is too expensive (i.e. GC already the limiting factor).

<pre>

Benchmark_Incremental_Origin_128_1-4       20000000      80.4 ns/op      12.43 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_2-4       20000000      89.5 ns/op      22.35 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_4-4       20000000     106 ns/op        37.72 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_8-4       10000000     108 ns/op        73.65 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_16-4      20000000     110 ns/op       144.97 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_32-4      20000000     102 ns/op      1253.59 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_64-4      20000000    88.3 ns/op       725.15 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_128-4     20000000     104 ns/op      1230.40 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_256-4     10000000     141 ns/op      1806.16 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_512-4     10000000     177 ns/op      2882.86 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_1024-4     5000000     317 ns/op      3226.23 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_2048-4     3000000     495 ns/op      4133.69 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_4096-4     2000000     876 ns/op      4673.40 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Origin_128_8192-4     1000000    1719 ns/op      4763.41 MB/s    16 B/op    1 allocs/op
Benchmark_Incremental_Forked_128_1-4       10000000     135 ns/op         7.36 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_2-4       10000000     160 ns/op        12.43 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_4-4       10000000     158 ns/op        25.19 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_8-4       10000000     152 ns/op        52.45 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_16-4      20000000     109 ns/op       145.64 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_32-4      10000000     131 ns/op       970.13 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_64-4      10000000     123 ns/op       517.18 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_128-4     10000000     151 ns/op       844.44 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_256-4     10000000     155 ns/op      1646.55 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_512-4     10000000     205 ns/op      2491.70 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_1024-4     5000000     305 ns/op      3346.95 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_2048-4     3000000     531 ns/op      3853.73 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_4096-4     2000000     908 ns/op      4506.57 MB/s     0 B/op    0 allocs/op
Benchmark_Incremental_Forked_128_8192-4     1000000    1711 ns/op      4785.81 MB/s     0 B/op    0 allocs/op

</pre>
