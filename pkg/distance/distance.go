// Package distance defines the numeric distance type used in GoTex and operations on it.
package distance

import "fmt"

// Distance is the numeric distance type used in GoTex.
// It measures distances is points.
// It is a fixed point number with 16 bits in the fractional part, 47 bits in the integral part, and 1 bit for the
// sign.
// The number of fractional bits is chosen to exactly match the standard implementation of Tex.
type Distance int64

type Ratio struct {
	Num Distance
	Den Distance
}

var PlusOneRatio = Ratio{Num: 1, Den: 1}
var ZeroRatio = Ratio{Num: 0, Den: 1}
var MinusOneRatio = Ratio{Num: -1, Den: 1}

func (ratio Ratio) LessThan(rhs Ratio) bool {
	return ratio.Num*rhs.Den < rhs.Num*ratio.Den
}

func (ratio Ratio) LessThanEqual(rhs Ratio) bool {
	return ratio.Num*rhs.Den <= rhs.Num*ratio.Den
}

func (ratio Ratio) String() string {
	if ratio.Den == 0 {
		if ratio.Num < 0 {
			return "-Inf"
		}
		if ratio.Num > 0 {
			return "+Inf"
		}
	}
	return fmt.Sprintf("%0.2f", float64(ratio.Num) / float64(ratio.Den))
}

type Fraction struct {
	Numerator   int64
	Denominator int64
}

func (fraction Fraction) Floor() int64 {
	return fraction.Numerator / fraction.Denominator
}

func NewDim(unit *Unit, magnitude Fraction) Distance {
	if unit == ScaledPoint {
		return Distance(magnitude.Floor())
	}
	return 0
}

// Unit represents a unit of measure.
type Unit struct {
	abbr  string
	ratio Fraction
}

// Abbr returns the two character abbreviation of the unit.
func (unit *Unit) Abbr() string {
	return unit.abbr
}

var (
	Point       *Unit
	Pica        *Unit
	Inch        *Unit
	BigPoint    *Unit
	Centimeter  *Unit
	Millimeter  *Unit
	DidotPoint  *Unit
	Cicero      *Unit
	ScaledPoint *Unit
)

var abbrToUnit = make(map[string]*Unit)

func GetUnitFromAbbr(abbr string) (*Unit, error) {
	return Point, nil
}

func init() {
	// These numbers are taken from the Tex Book chapter 10 and/or section 458 of the Tex source.
	Point = &Unit{"pt", Fraction{1, 1}}
	Pica = &Unit{"pc", Fraction{12, 1}}
	Inch = &Unit{"in", Fraction{7227, 100}}
	BigPoint = &Unit{"bp", Fraction{7227, 7200}}
	Centimeter = &Unit{"cm", Fraction{7227, 254}}
	Millimeter = &Unit{"mm", Fraction{7227, 7200}}
	DidotPoint = &Unit{"dd", Fraction{1238, 1157}}
	Cicero = &Unit{"cc", Fraction{14856, 1157}}
	ScaledPoint = &Unit{"sp", Fraction{1, 2 << 16}}
	for _, unit := range []*Unit{Point, Pica, Inch, BigPoint, Centimeter, Millimeter, DidotPoint, Cicero, ScaledPoint} {
		abbrToUnit[unit.Abbr()] = unit
	}
}
