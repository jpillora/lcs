# lcs

Another implementation of [](https://en.wikipedia.org/wiki/Longest_common_subsequence_problem). This implementation is an optimized version with `Operations` exposed.

:warning: Many TODOs still:

* This is a direct port of [this C# version](http://www.mathertel.de/Diff/ViewSrc.aspx) so needs to be cleaned up to be idiomatic Go.
* Need to define a `Sequence interface` (Similar to the `sort.Interface` interface)
* Need to define a `Diff(s1, s1 Sequence) []Operation`
* Need to define a `StringSequence` implementation for strings
* Need to define a `Patch(seq Sequence, ops []Operation) Sequence` function
* Need to define a `Hasher` interface for optional equivalence caching

---

Try it out with

``` sh
$ go test
2016/02/28 21:14:54 XMJYAUZZZ -> [1 2 3 4 5 6 7 7 7]
2016/02/28 21:14:54 MZJAWXU -> [2 7 3 5 8 1 6]
0: {StartA:0 StartB:0 DeleteA:1 InsertB:0} = del at [0] MJYAUZZZ
1: {StartA:2 StartB:1 DeleteA:0 InsertB:1} = ins at [2] MZJYAUZZZ
2: {StartA:3 StartB:3 DeleteA:1 InsertB:0} = del at [3] MZJAUZZZ
3: {StartA:5 StartB:4 DeleteA:0 InsertB:2} = ins at [5] MZJAWXUZZZ
4: {StartA:6 StartB:7 DeleteA:3 InsertB:0} = del at [6] MZJAWXU
PASS
```
