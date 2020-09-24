package knuthplass

import (
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
	GetMaxAdjustmentRatio() float64
	GetLooseness() int
	CalculateDemerits(
		adjustmentRatio float64,
		fitnessClass FitnessClass,
		prevFitnessClass FitnessClass,
		penaltyCost int64,
		isFlaggedPenalty bool,
		isPrevFlaggedPenalty bool) float64
	CalculateFitnessClass(adjustmentRatio float64) FitnessClass
}

// TexOptimalityCriteria is the optimality criteria developed for Tex and described
// in the Knuth-Plass paper
type TexOptimalityCriteria struct {
	MaxAdjustmentRatio            float64 // ro in the paper
	Looseness                     int     // q in the paper
	ConsecutiveFlaggedPenaltyCost float64 // alpha in the paper
	MismatchingFitnessClassCost   float64 // gamma in the paper
}

// GetMaxAdjustmentRatio returns the largest legal adjustment ratio.
func (criteria TexOptimalityCriteria) GetMaxAdjustmentRatio() float64 {
	return criteria.MaxAdjustmentRatio
}

// GetLooseness returns the looseness parameter.
func (criteria TexOptimalityCriteria) GetLooseness() int {
	return criteria.Looseness
}

// CalculateDemerits calculates the demerits of the line following the formulas in the Knuth-Plass paper.
func (criteria TexOptimalityCriteria) CalculateDemerits(
	adjustmentRatio float64,
	fitnessClass FitnessClass,
	prevFitnessClass FitnessClass,
	penaltyCost int64,
	isFlaggedPenalty bool,
	isPrevFlaggedPenalty bool) (demerits float64) {
	// Section 858 of the Tex source
	if penaltyCost >= 0 {
		demerits = math.Pow(1+100*math.Pow(adjustmentRatio, 3)+float64(penaltyCost), 2)
	} else if penaltyCost > NegInfBreakpointPenalty {
		demerits = math.Pow(1+100*math.Pow(adjustmentRatio, 3), 2) - float64(penaltyCost*penaltyCost)
	} else {
		demerits = math.Pow(1+100*math.Pow(adjustmentRatio, 3), 2)
	}
	if isFlaggedPenalty && isPrevFlaggedPenalty {
		demerits = demerits + criteria.ConsecutiveFlaggedPenaltyCost
	}
	if fitnessClass-prevFitnessClass > 1 || fitnessClass-prevFitnessClass < -1 {
		demerits = demerits + criteria.MismatchingFitnessClassCost
	}
	return
}

// calculateBadness calculates the badness of an adjustment ratio.
//
// The badness is essentially what the Knuth-Plass algorithm is trying to minimize, and hence to get identical results
// to Tex "all implementations of TeX should use precisely this method" as Knuth says in section 108. Through some
// reverse engineering we are able to provide explanations of the magic constants as code comments.
func calculateBadness(numerator int64, denominator int64) int64 {
	// quotient is an approximation to (alpha * num / den) where alpha^3 ~= 100 * 2^18
	var quotient int64
	switch true {
	case numerator == 0:
		return 0
	case denominator <= 0:
		return 10000
	case numerator <= 7230584:
		// 7230584 is the smallest integer less than 2^31/297. Knuth presumably chooses it so that that the following
		// multiplication doesn't overflow on a 32 bit machine.
		quotient = (numerator * 297) / denominator
	case denominator >= 1663497:
		// 1663497 is the smallest integer such that quotient is less than or equal to 1290, and hence
		// that the final result (quotient^3/2^18 rounded) is less than or equal to 8192. Any number bigger than
		// 8192=2^13 yields an infinite badness of 10000 and no computation is needed - we just return the infinite
		// badness in the following case.
		quotient = numerator / (denominator / 297)
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
func (TexOptimalityCriteria) CalculateFitnessClass(adjustmentRatio float64) FitnessClass {
	// Binary search essentially...
	if adjustmentRatio <= 0.5 {
		if adjustmentRatio <= -0.5 {
			return -1
		}
		return 0
	}
	if adjustmentRatio <= 1 {
		return 1
	}
	return 2

}
