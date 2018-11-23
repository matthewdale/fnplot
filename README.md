# fnplot [![Build Status](https://travis-ci.org/matthewdale/fnplot.svg?branch=master)](https://travis-ci.org/matthewdale/fnplot) [![codecov](https://codecov.io/gh/matthewdale/fnplot/branch/master/graph/badge.svg)](https://codecov.io/gh/matthewdale/fnplot) [![Go Report Card](https://goreportcard.com/badge/github.com/matthewdale/fnplot)](https://goreportcard.com/report/github.com/matthewdale/fnplot) [![GoDoc](https://godoc.org/github.com/matthewdale/fnplot?status.svg)](https://godoc.org/github.com/matthewdale/fnplot)

Package `fnplot` provides a way to plot input/output values for arbitrary functions on a 2D cartesian coordinate plot. The input/output values are converted to scalar values (arbitrary precision floating point) in a way that attempts to preserve the relative scale of the original values.

# Examples

### Numeric Functions
Plot the `math.Sin` function between 0 and 100 on the X axis.

```go
import (
    "math"
    "github.com/matthewdale/fnplot"
)

func main() {
    err := fnplot.FnPlot{
        Fn: fnplot.NewFn(
            math.Sin,
            fnplot.Float64Range(0, 100),
        ),
        Filename: "sin.png",
        Title:    "math.Sin",
        Samples:  2000,
        X:        &fnplot.StdAxix{},
        Y:        &fnplot.StdAxix{},
    }.Save()

    if err != nil {
        panic(err)
    }
}
```

### Binary Functions
Plot the `md5.Sum` function using a natural log X axis and scaled Y axis.

```go
import (
    "crypto/md5"
    "github.com/matthewdale/fnplot"
)

func main() {
    err := fnplot.FnPlot{
        Fn: fnplot.NewFn(
            func(s string) [md5.Size]byte {
                return md5.Sum([]byte(s))
            },
            fnplot.AnyString()),
        Filename: "md5.png",
        Title:    "md5.Sum",
        Samples:  2000,
        X:        &fnplot.LnAxis{},
        Y:        &fnplot.ScaledAxis{Max: 1000},
    }.Save()

    if err != nil {
        panic(err)
    }
}
```
