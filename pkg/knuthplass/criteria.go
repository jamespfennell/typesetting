package knuthplass

import (
	"fmt"
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

// TexOptimalityCriteria is the optimiality criteria developed for Tex and described
// in the Knuth-Plass paper
type TexOptimalityCriteria struct {
	MaxAdjustmentRatio            float64 // ro in the paper
	Looseness                     int     // q in the paper
	ConsecutiveFlaggedPenaltyCost float64 // alpha in the paper
	MismatchingFitnessClassCost   float64 // gamma in the paper
}

func (texOptimalityCritera TexOptimalityCriteria) GetMaxAdjustmentRatio() float64 {
	return texOptimalityCritera.MaxAdjustmentRatio
}

func (texOptimalityCritera TexOptimalityCriteria) GetLooseness() int {
	return texOptimalityCritera.Looseness
}

// CalculateDemerits does...
func (texOptimalityCritera TexOptimalityCriteria) CalculateDemerits(
	adjustmentRatio float64,
	fitnessClass FitnessClass,
	prevFitnessClass FitnessClass,
	penaltyCost int64,
	isFlaggedPenalty bool,
	isPrevFlaggedPenalty bool) (demerits float64) {
	if penaltyCost >= 0 {
		demerits = math.Pow(1+100*math.Pow(adjustmentRatio, 3)+float64(penaltyCost), 2)
	} else if penaltyCost > NegInfBreakpointPenalty {
		demerits = math.Pow(1+100*math.Pow(adjustmentRatio, 3), 2) - float64(penaltyCost*penaltyCost)
	} else {
		demerits = math.Pow(1+100*math.Pow(adjustmentRatio, 3), 2)
	}
	if isFlaggedPenalty && isPrevFlaggedPenalty {
		demerits = demerits + texOptimalityCritera.ConsecutiveFlaggedPenaltyCost
	}
	if fitnessClass-prevFitnessClass > 1 || fitnessClass-prevFitnessClass < -1 {
		fmt.Println("Adding penalty for mismatching fitness class", adjustmentRatio)
		demerits = demerits + texOptimalityCritera.MismatchingFitnessClassCost
	}
	return
}

func (TexOptimalityCriteria) CalculateFitnessClass(adjustmentRatio float64) FitnessClass {
	// Binary search esentially...
	if adjustmentRatio <= 0.5 {
		if adjustmentRatio <= -0.5 {
			return 0
		}
		return 1
	}
	if adjustmentRatio <= 1 {
		return 2
	}
	return 3

}
