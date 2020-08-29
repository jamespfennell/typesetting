package main

import (
	"fmt"

	"github.com/jamespfennell/typesetting/pkg/knuthplass"
)

func test(item knuthplass.Item) {
	fmt.Println("inspecting item")
	fmt.Println("width:", item.Width())
	fmt.Println("shrinkability:", item.Shrinkability())
	fmt.Println("Is penalty?", knuthplass.IsPenalty(item))
}
func main() {
	box := knuthplass.NewBox(40)
	fmt.Println(box)
	fmt.Println(box.Width())
	test(box)

	penalty := knuthplass.NewPenalty(40, 1, false)
	test(penalty)

	var items = []knuthplass.Item{}
	items = append(items, box, penalty)
	fmt.Println(items)

	var criteria knuthplass.OptimalityCriteria
	criteria = &knuthplass.TexOptimalityCriteria{}
	fmt.Println(criteria)
}
