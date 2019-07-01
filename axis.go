package fnplot

import (
	"math/big"

	"github.com/ALTree/bigfloat"
)

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
}
