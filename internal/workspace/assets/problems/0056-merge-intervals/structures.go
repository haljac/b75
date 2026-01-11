package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

// --- Content of Interval.go ---
// Interval 提供区间表示
type Interval struct {
	Start	int
	End	int
}

// Interval2Ints 把 Interval 转换成 整型切片
func Interval2Ints(i Interval) []int {
	return []int{i.Start, i.End}
}

// IntervalSlice2Intss 把 []Interval 转换成 [][]int
func IntervalSlice2Intss(is []Interval) [][]int {
	res := make([][]int, 0, len(is))
	for i := range is {
		res = append(res, Interval2Ints(is[i]))
	}
	return res
}

// Intss2IntervalSlice 把 [][]int 转换成 []Interval
func Intss2IntervalSlice(intss [][]int) []Interval {
	res := make([]Interval, 0, len(intss))
	for _, ints := range intss {
		res = append(res, Interval{Start: ints[0], End: ints[1]})
	}
	return res
}

// QuickSort define
func QuickSort(a []Interval, lo, hi int) {
	if lo >= hi {
		return
	}
	p := partitionSort(a, lo, hi)
	QuickSort(a, lo, p-1)
	QuickSort(a, p+1, hi)
}





var _ = fmt.Printf
var _ = math.Abs
var _ = sort.Ints
var _ = strconv.Atoi

