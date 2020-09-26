package knuthplass

import (
	d "github.com/jamespfennell/typesetting/pkg/distance"
	"math"
)

// FitnessClass can be thought of as an enum type: it classifies lines into a class
// represented by a value. The classification is only as a function of the adjustment ratio.
// The demerits function has access to the previous class.
// Note that the Knuth-Plass algorithm is O(# of fitness classes)
// so be conservative! The default optimiality criteria divides lines into 4 fitness
// classes
type FitnessClass int8

// OptimalityCriteria contains all of the data to optimize
type OptimalityCriteria interface {
	GetMaxAdjustmentRatio() d.Ratio
	GetLooseness() int
	CalculateDemerits(
		adjustmentRatio d.Ratio,
		fitnessClass FitnessClass,
		prevFitnessClass FitnessClass,
		penaltyCost int64,
		isFlaggedPenalty bool,
		isPrevFlaggedPenalty bool) float64
	CalculateFitnessClass(adjustmentRatio d.Ratio) FitnessClass
}

// TexOptimalityCriteria is the optimality criteria developed for Tex and described
// in the Knuth-Plass paper
type TexOptimalityCriteria struct {
	MaxAdjustmentRatio            d.Ratio // ro in the paper
	Looseness                     int     // q in the paper
	ConsecutiveFlaggedPenaltyCost float64 // alpha in the paper
	MismatchingFitnessClassCost   float64 // gamma in the paper
}

// GetMaxAdjustmentRatio returns the largest legal adjustment ratio.
func (criteria TexOptimalityCriteria) GetMaxAdjustmentRatio() d.Ratio {
	return criteria.MaxAdjustmentRatio
}

// GetLooseness returns the looseness parameter.
func (criteria TexOptimalityCriteria) GetLooseness() int {
	return criteria.Looseness
}

// CalculateDemerits calculates the demerits of the line following the formulas in the Knuth-Plass paper.
func (criteria TexOptimalityCriteria) CalculateDemerits(
	adjustmentRatio d.Ratio,
	fitnessClass FitnessClass,
	prevFitnessClass FitnessClass,
	penaltyCost int64,
	isFlaggedPenalty bool,
	isPrevFlaggedPenalty bool) (demerits float64) {
	// Section 858 of the Tex source
	// TODO: don't cast to float
	badness := float64(calculateBadness(adjustmentRatio))
	if penaltyCost >= 0 {
		demerits = math.Pow(1+badness+float64(penaltyCost), 2)
	} else if penaltyCost > NegInfBreakpointPenalty {
		demerits = math.Pow(1+badness, 2) - float64(penaltyCost*penaltyCost)
	} else {
		demerits = math.Pow(1+badness, 2)
	}
	if isFlaggedPenalty && isPrevFlaggedPenalty {
		demerits = demerits + criteria.ConsecutiveFlaggedPenaltyCost
	}
	if fitnessClass-prevFitnessClass > 1 || fitnessClass-prevFitnessClass < -1 {
		demerits = demerits + criteria.MismatchingFitnessClassCost
	}
	return
}

// calculateBadness calculates the badness of an adjustment ratio. It is approximately 100 * ratio^3.
//
// The badness is essentially what the Knuth-Plass algorithm is trying to minimize, and hence to get identical results
// to Tex "all implementations of TeX should use precisely this method" as Knuth says in section 108. Through some
// reverse engineering we are able to provide explanations of the magic constants as code comments.
func calculateBadness(ratio d.Ratio) int64 {
	// quotient is an approximation to (alpha * num / den) where alpha^3 ~= 100 * 2^18
	var quotient int64
	switch true {
	case ratio.Num == 0:
		return 0
	case ratio.Den <= 0:
		return 10000
	case ratio.Num <= 7230584:
		// 7230584 is the smallest integer less than 2^31/297. Knuth presumably chooses it so that that the following
		// multiplication doesn't overflow on a 32 bit machine.
		quotient = int64((ratio.Num * 297) / ratio.Den)
	case ratio.Den >= 1663497:
		// 1663497 is the smallest integer such that quotient is less than or equal to 1290, and hence
		// that the final result (quotient^3/2^18 rounded) is less than or equal to 8192. Any number bigger than
		// 8192=2^13 yields an infinite badness of 10000 and no computation is needed - we just return the infinite
		// badness in the following case.
		quotient = int64(ratio.Num / (ratio.Den / 297))
	default:
		// In this case num/den > 7230584/1663497 > 4.346, in which case 100(num/den)^3 > 8200 > 8192, and so the
		// badness is infinite. Knuth's code returns this value, but the way it's laid out is confusing because
		// the quotient is set to be the numerator which breaks the scaling. Our code is equivalent because if we set
		// quotient = numerator, the next if statement would evaluate to true and return 10000 anyway.
		return 10000
	}
	if quotient > 1290 {
		return 10000
	}
	return (quotient * quotient * quotient + 2 << 16) / (2 << 17)
}

// CalculateFitnessClass calculates the fitness class of the line following the formulas in the Knuth-Plass paper.
func (TexOptimalityCriteria) CalculateFitnessClass(ratio d.Ratio) FitnessClass {
	// Binary search essentially...
	if ratio.LessThanEqual(d.Ratio{1,2}) {
		if ratio.LessThanEqual(d.Ratio{-1, 2}) {
				return -1
		}
		return 0
	}
	if ratio.LessThanEqual(d.Ratio{Num: 1, Den: 1})  {
		return 1
	}
	return 2

}
