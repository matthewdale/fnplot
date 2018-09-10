# fnplot [![Build Status](https://travis-ci.org/matthewdale/fnplot.svg?branch=master)](https://travis-ci.org/matthewdale/fnplot) [![codecov](https://codecov.io/gh/matthewdale/fnplot/branch/master/graph/badge.svg)](https://codecov.io/gh/matthewdale/fnplot) [![Go Report Card](https://goreportcard.com/badge/github.com/matthewdale/fnplot)](https://goreportcard.com/report/github.com/matthewdale/fnplot) [![GoDoc](https://godoc.org/github.com/matthewdale/fnplot?status.svg)](https://godoc.org/github.com/matthewdale/fnplot)

Package `fnplot` provides a way to plot input/output values for arbitrary functions on a 2D cartesian coordinate plot. The input/output values are converted to scalar values (arbitrary precision floating point) in a way that attempts to preserve the relative scale of the original values.

# Examples

### Numeric Functions
Plot the `math.Sin` function between 0 and 100 on the X axis.

```go
import (
    "math"
    "github.com/leanovate/gopter/gen"
    "github.com/matthewdale/fnplot"
)

func main() {
    p := fnplot.FnPlot{
        Filename: "sin.png",
        Title:    "math.Sin",
        Fn: fnplot.NewFn(
            func(set *fnplot.ValuesSet) interface{} {
                return func(x float64) bool {
                    y := math.Sin(x)
                    set.Insert(fnplot.Values{x}, fnplot.Values{y})
                    return true
                }
            },
            gen.Float64Range(0, 100),
        ),
        Samples: 2000,
        X:       &fnplot.StdAxix{},
        Y:       &fnplot.StdAxix{},
    }
    if err := p.Save(); err != nil {
        panic(err)
    }
}
```

### Binary Functions
Plot the `md5.Sum` function using a natural log X axis and scaled Y axis.

```go
import (
    "crypto/md5"
    "github.com/leanovate/gopter/gen"
    "github.com/matthewdale/fnplot"
)

func main() {
    p := fnplot.FnPlot{
        Filename: "md5.png",
        Title:    "md5.Sum",
        Fn: fnplot.NewFn(
            func(set *fnplot.ValuesSet) interface{} {
                return func(s string) bool {
                    sum := md5.Sum([]byte(s))
                    set.Insert(fnplot.Values{s}, fnplot.Values{sum})
                    return true
                }
            },
            gen.AnyString()),
        Samples: 2000,
        X:       &fnplot.LnAxis{},
        Y:       &fnplot.ScaledAxis{Max: 1000},
    }
    if err := p.Save(); err != nil {
        panic(err)
    }
}
```
