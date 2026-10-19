package main

import "math"

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	if len(nums) < 2 {
		return Stats{}
	}
	resStats := Stats{
		Count: len(nums) - 1,
		Min:   math.MaxInt64,
		Max:   math.MinInt64,
	}
	for i := 1; i < len(nums); i++ {
		n := nums[i] - nums[i-1]
		resStats.Sum += n
		resStats.Min = min(resStats.Min, n)
		resStats.Max = max(resStats.Max, n)
	}

	return resStats
}
