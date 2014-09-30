package cnc

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

func TestGridStep(t *testing.T) {
	grid := [][]int{
		[]int{0, 1, 0, 0, 0, 0},
		[]int{0, 0, 1, 0, 0, 0},
		[]int{1, 1, 1, 0, 0, 0},
		[]int{0, 0, 0, 0, 0, 0},
		[]int{0, 0, 0, 0, 0, 0},
	}

	w, h := GridSize(grid)
	checkEq(w, 6)
	checkEq(h, 5)

	next_grid := [][]int{
		[]int{0, 0, 0, 0, 0, 0},
		[]int{1, 0, 1, 0, 0, 0},
		[]int{0, 1, 1, 0, 0, 0},
		[]int{0, 1, 0, 0, 0, 0},
		[]int{0, 0, 0, 0, 0, 0},
	}
	checkEq(GridStep(grid), next_grid)
}
