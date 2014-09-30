package main

import (
	"fmt"
	"reflect"
	"testing"
)

func checkEq(actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		fmt.Println("Expected:", expected)
		fmt.Println("Actual:  ", actual)
		panic("checkEq failed")
	}
}

func TestDominantNeighbor(t *testing.T) {
	checkEq(GetDominantNeighbor(&[...]int{1, 1, 2, 0, 0, 0, 0, 0}), 1)
	checkEq(GetDominantNeighbor(&[...]int{0, 0, 0, 0, 1, 2, 2, 0}), 2)
	checkEq(GetDominantNeighbor(&[...]int{1, 2, 0, 0, 0, 0, 0, 3}), 1)
}

func TestGetNeighbors(t *testing.T) {
	grid := [][]int{
		[]int{0, 1, 2, 3},
		[]int{4, 5, 6, 7},
		[]int{8, 9, 10, 11},
	}

	var neighbors [8]int
	GetNeighbors(grid, 1, 0, &neighbors)

	checkEq(neighbors, [...]int{
		3, 0, 1,
		7 /**/, 5,
		11, 8, 9})
}

func TestGridStep(t *testing.T) {
	grid := [][]int{
		[]int{0, 1, 0, 0, 0, 0},
		[]int{0, 0, 1, 0, 0, 0},
		[]int{2, 2, 2, 0, 0, 0},
		[]int{0, 0, 0, 0, 0, 0},
		[]int{0, 0, 0, 0, 0, 0},
	}

	w, h := GridSize(grid)
	checkEq(w, 6)
	checkEq(h, 5)

	checkEq(GridStep(grid), [][]int{
		[]int{0, 0, 0, 0, 0, 0},
		[]int{2, 0, 1, 0, 0, 0},
		[]int{0, 2, 2, 0, 0, 0},
		[]int{0, 2, 0, 0, 0, 0},
		[]int{0, 0, 0, 0, 0, 0},
	})
}
