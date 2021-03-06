# [Go: Inlining Strategy & Limitation](https://medium.com/a-journey-with-go/go-inlining-strategy-limitation-6b6d7fc3b1be)

![image](https://user-images.githubusercontent.com/1940588/74205115-ecdbbc00-4cb0-11ea-8e16-d298f544b07c.png)

Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.


The [inlining](https://github.com/golang/go/wiki/CompilerOptimizations#function-inlining) process replaces a function call by the body of this function. Although this optimization increases the binary size, it improves the performance of the programs. However, Go does not inline all the functions and follows some rules.

## Rules

`go run -gcflags="-m" .`

```bash
$ go run -gcflags="-m" .
# InliningStrategyLimitation
./op.go:3:6: can inline add
./op.go:7:6: can inline sub
./main.go:16:11: inlining call to sub
./main.go:14:11: inlining call to add
./main.go:7:12: inlining call to fmt.Printf
./main.go:10:10: sum s does not escape
./main.go:6:16: main []float32 literal does not escape
./main.go:7:40: sum(n) escapes to heap
./main.go:7:12: main []interface {} literal does not escape
./main.go:7:12: io.Writer(os.Stdout) escapes to heap
<autogenerated>:1: (*File).close .this does not escape
The total is -167.40
```

`go run -gcflags="-m -m" .`

```bash'
$ go run -gcflags="-m -m" .
# InliningStrategyLimitation
./op.go:3:6: can inline add as: func(float32, float32) float32 { return a + b }
./op.go:7:6: can inline sub as: func(float32, float32) float32 { return a - b }
./main.go:10:6: cannot inline sum: unhandled op RANGE
./main.go:16:11: inlining call to sub func(float32, float32) float32 { return a - b }
./main.go:14:11: inlining call to add func(float32, float32) float32 { return a + b }
./main.go:5:6: cannot inline main: function too complex: cost 148 exceeds budget 80
./main.go:7:12: inlining call to fmt.Printf func(string, ...interface {}) (int, error) { var fmt..autotmp_4 int; fmt..autotmp_4 = <N>; var fmt..autotmp_5 error; fmt..autotmp_5 = <N>; fmt..autotmp_4, fmt..autotmp_5 = fmt.Fprintf(io.Writer(os.Stdout), fmt.format, fmt.a...); return fmt..autotmp_4, fmt..autotmp_5 }
./main.go:10:10: sum s does not escape
./main.go:6:16: main []float32 literal does not escape
./main.go:7:40: sum(n) escapes to heap
./main.go:7:12: main []interface {} literal does not escape
./main.go:7:12: io.Writer(os.Stdout) escapes to heap
<autogenerated>:1: (*File).close .this does not escape
The total is -167.40
```

Go does not inline methods that use the range operation. Indeed, some operations block the inlining, such as closure calls, select, for, defer, and goroutine creation with go. However, this is not the only rule. When parsing the AST graph, Go allocates a budget of 80 nodes for the inlining. Each node consumes one of the budgets when functions call consumes the cost of their inlining. 

> 静态单赋值（Static Single Assigment, SSA）是中间代码的一个特性，如果一个中间代码具有静态单赋值的特性，那么每个变量就只会被赋值一次。

-- 摘自[Go 语言设计与实现](https://draveness.me/golang/docs/part1-prerequisite/ch02-compile/golang-compile-intro/)

More to read [How to read GoLang static single-assignment (SSA) form intermediate representation](https://sitano.github.io/2018/03/18/howto-read-gossa/)

```bash
$ env GOSSAFUNC=Aadd go build .
# InliningStrategyLimitation
dumped SSA to ./ssa.html

$ open ssa.html
```

As an example, the following instruction a = a + 1 represents five nodes: AS, NAME, ADD, NAME, LITERAL. Here is the SSA dump:

![image](https://user-images.githubusercontent.com/1940588/74205739-886e2c00-4cb3-11ea-8ba7-1bb0065dc9a6.png)

## Challenge

During the process of inlining, it removes some function calls, meaning the program is getting modified. However, when a panic occurs, the developers need to know the exact stack traces to get the file and the line where it happened.

Go maps the inlined functions in the generated code. It also maps the lines, you can visualize it with the flag `-gcflags="-d pctab=pctoline"`.

```bash
$ go run -gcflags="-d pctab=pctoline" .
...
    36        00054 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:12)  MOVSS   (AX)(DX*4), X1
    3b     13 00059 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:13)  XORPS   X2, X2
    3e        00062 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:13)  UCOMISS X0, X2
    41        00065 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:13)  JLS     78
    43        00067 (<unknown line number>)     NOP
    43      4 00067 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:14)  UCOMISS X1, X2
    46        00070 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:14)  JHI     100
    48      8 00072 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:14)  ADDSS   X1, X0
    4c     14 00076 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:14)  JMP     46
    4e        00078 (<unknown line number>)     NOP
    4e     12 00078 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:16)  SUBSS   X1, X0
...
```

The files are mapped as well, and can be displayed with the flag `-gcflags="-d pctab=pctofile"`. Here is the output:

```bash
$ go run -gcflags="-d pctab=pctofile" .
...
    41        00065 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:13)  JLS     78
    43        00067 (<unknown line number>)     NOP
    43      1 00067 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:14)  UCOMISS X1, X2
    46        00070 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:14)  JHI     100
    48        00072 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:14)  ADDSS   X1, X0
    4c      0 00076 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:14)  JMP     46
    4e        00078 (<unknown line number>)     NOP
    4e      1 00078 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:16)  SUBSS   X1, X0
    52      0 00082 (/Users/bingoo/GitHub/golang-gotcha/InliningStrategyLimitation/main.go:16)  JMP     46
...
```

We now have a proper mapping of each the generated instructions:

PC|Instruction|Func|File|Line
---|---|---|---|---
48|ADDSS   X1, X0|0 add | 1 op.go|3
4c|JMP 46|-1 sum|1 main.go|14
4e|SUBSS X1,X0|1 sub|1 op.go|11
52|JMP 46|-1 sum|1 main.go|16


## Impact

Inlining is important and be critical for applications that need high performance. A function call has an overhead — creation of a new stack frame, save and restore registers — and can be avoided with inlining. However, the copy of the body rather than a function call increases the binary size. Here is an example with the benchmark suite [go1](https://github.com/golang/go/tree/release-branch.go1.13/test/bench/go1) with and without inlining:

```bash
$ go test -bench=. ./... > with.bench
$ go test -gcflags '-N -l' -bench=. ./... > without.bench
$ benchcmp with.bench without.bench
benchmark                             old ns/op      new ns/op      delta
BenchmarkBinaryTree17-12              2194796388     2295813074     +4.60%
BenchmarkFannkuch11-12                1984853189     5954795767     +200.01%
```

More benchmark articles 

1. [Practical Go Benchmarks](https://stackimpact.com/blog/practical-golang-benchmarks/)
1. [Turning off optimization and inlining in Go gc compilers](https://gist.github.com/tetsuok/3025333)

    `go build -gcflags '-N -l' [code.go]` or `go install -gcflags '-N -l' [code.go]`

1. [Command compile](https://golang.org/cmd/compile/)

    ```
    -N
    	Disable optimizations.
    -l
    	Disable inlining.
    ```

The performance with inlining are `~5/6% better` than without for this benchmark suite.
