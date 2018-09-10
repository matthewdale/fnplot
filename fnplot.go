package fnplot

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/ALTree/bigfloat"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"github.com/pkg/errors"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type sortablePoints plotter.XYs

func (sp sortablePoints) Len() int           { return len(sp) }
func (sp sortablePoints) Swap(i, j int)      { sp[i], sp[j] = sp[j], sp[i] }
func (sp sortablePoints) Less(i, j int) bool { return sp[i].X < sp[j].X }

type ioPair struct {
	input  Values
	output Values
}

type ValuesSet struct {
	ios       []ioPair
	minInput  *big.Float
	maxInput  *big.Float
	minOutput *big.Float
	maxOutput *big.Float
	mu        sync.RWMutex
}

func NewValuesSet() *ValuesSet {
	return &ValuesSet{
		ios: make([]ioPair, 10),
	}
}

func (set *ValuesSet) Insert(input, output Values) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	set.ios = append(set.ios, ioPair{input: input, output: output})
	in, err := input.Scalar()
	if err != nil {
		return errors.WithMessage(err, "error converting input to int")
	}
	if set.minInput == nil || set.minInput.Cmp(in) == 1 {
		set.minInput = in
	}
	if set.maxInput == nil || set.maxInput.Cmp(in) == -1 {
		set.maxInput = in
	}
	out, err := output.Scalar()
	if err != nil {
		return errors.WithMessage(err, "error converting output to int")
	}
	if set.minOutput == nil || set.minOutput.Cmp(out) == 1 {
		set.minOutput = out
	}
	if set.maxOutput == nil || set.maxOutput.Cmp(out) == -1 {
		set.maxOutput = out
	}
	return nil
}

type Axis interface {
	Point(*big.Float) float64
	SetMaxValue(*big.Float)
}

type StdAxix struct{}

func (StdAxix) Point(p *big.Float) float64 {
	fp, _ := p.Float64()
	return fp
}

func (*StdAxix) SetMaxValue(*big.Float) {}

type ScaledAxis struct {
	Max   float64
	ratio *big.Float
}

func (sa ScaledAxis) Point(p *big.Float) float64 {
	scaled, _ := big.NewFloat(0).Mul(p, sa.ratio).Float64()
	return scaled
}

func (sa *ScaledAxis) SetMaxValue(v *big.Float) {
	sa.ratio = big.NewFloat(0).Quo(big.NewFloat(sa.Max), v)
	log.Printf("Scaling ratio: %s", sa.ratio.String())
}

type LnAxis struct{}

func (la LnAxis) Point(p *big.Float) float64 {
	if p.Cmp(big.NewFloat(0)) == 0 {
		return 0
	}
	scaled, _ := bigfloat.Log(p).Float64()
	return scaled
}

func (*LnAxis) SetMaxValue(*big.Float) {}

type LnScaledAxis struct {
	Max   float64
	ratio *big.Float
}

func (lsa LnScaledAxis) Point(p *big.Float) float64 {
	if p.Cmp(big.NewFloat(0)) == 0 {
		return 0
	}
	scaled, _ := big.NewFloat(0).Mul(bigfloat.Log(p), lsa.ratio).Float64()
	return scaled
}

func (lsa *LnScaledAxis) SetMaxValue(v *big.Float) {
	lsa.ratio = big.NewFloat(0).Quo(big.NewFloat(lsa.Max), bigfloat.Log(v))
	log.Printf("Ln scaling ratio: %s", lsa.ratio.String())
}

func (set *ValuesSet) PointsOn(x, y Axis) (plotter.XYs, error) {
	set.mu.RLock()
	defer set.mu.RUnlock()

	x.SetMaxValue(set.maxInput)
	y.SetMaxValue(set.maxOutput)

	points := make(plotter.XYs, len(set.ios))
	for i := range set.ios {
		inputScalar, err := set.ios[i].input.Scalar()
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("error converting input %d to int", i))
		}
		points[i].X = x.Point(inputScalar)

		outputScalar, err := set.ios[i].output.Scalar()
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("error converting output %d to int", i))
		}
		points[i].Y = y.Point(outputScalar)

		if pt := points[i]; math.IsInf(pt.X, 0) || math.IsInf(pt.Y, 0) {
			log.Printf(
				"Infinity found at value %d. Input: %s, output: %s, scaled X: %f, scaled Y: %f",
				i,
				inputScalar.String(),
				outputScalar.String(),
				pt.X,
				pt.Y)
		}
	}
	sort.Sort(sortablePoints(points))
	return points, nil
}

type Fn struct {
	p   gopter.Prop
	set *ValuesSet
}

func NewFn(fn func(*ValuesSet) interface{}, gens ...gopter.Gen) Fn {
	set := NewValuesSet()
	return Fn{
		p:   prop.ForAll(fn(set), gens...),
		set: set,
	}
}

func (fn Fn) Run(samples int) error {
	res := fn.p.Check(&gopter.TestParameters{
		MinSuccessfulTests: samples,
		MinSize:            100,
		MaxSize:            10000,
		MaxShrinkCount:     1000, // ?
		Seed:               time.Now().UnixNano(),
		Rng:                rand.New(gopter.NewLockedSource(time.Now().UnixNano())),
		Workers:            10,
		MaxDiscardRatio:    5,
	})
	if err := res.Error; err != nil {
		return err
	}
	return nil
}

func (fn Fn) ValuesSet() *ValuesSet {
	return fn.set
}

type FnPlot struct {
	Title    string
	Filename string
	Fn       Fn
	Samples  int
	X, Y     Axis
}

func (fp *FnPlot) Save() error {
	if err := fp.Fn.Run(fp.Samples); err != nil {
		return errors.WithMessage(err, "error running function")
	}
	p, err := plot.New()
	if err != nil {
		return errors.WithMessage(err, "error creating plot")
	}
	p.Title.Text = fp.Title
	p.X.Label.Text = " "
	p.Y.Label.Text = " "

	points, err := fp.Fn.ValuesSet().PointsOn(fp.X, fp.Y)
	if err != nil {
		log.Fatalf("Error generating X,Y points: %s", err)
	}
	err = plotutil.AddLinePoints(p, "Fn", points)
	if err != nil {
		panic(err)
	}

	// Save the plot to a file. The format is determined by the file extension.
	err = p.Save(20*vg.Inch, 4*vg.Inch, fp.Filename)
	return errors.WithMessage(err, "error writing plot image")
}
