package knuthplass

import "testing"

var paramsList = []struct {
	preceedingItem    Item
	item              Item
	isValidBreakpoint bool
}{
	{nil, &box{}, false},
	{&box{}, &box{}, false},
	{&glue{}, &box{}, false},
	{&penalty{}, &box{}, false},
	{nil, &glue{}, false},
	{&box{}, &glue{}, true},
	{&glue{}, &glue{}, false},
	{&penalty{}, &glue{}, false},
	{nil, &penalty{}, true},
	{&box{}, &penalty{}, true},
	{&glue{}, &penalty{}, true},
	{&penalty{}, &penalty{}, true},
	{nil, &penalty{cost: 20000}, false},
	{&box{}, &penalty{cost: 20000}, false},
	{&glue{}, &penalty{cost: 20000}, false},
	{&penalty{}, &penalty{cost: 20000}, false},
}

func TestIsValidBreakpoint(t *testing.T) {
	for _, params := range paramsList {
		t.Run("", func(t *testing.T) {
			if IsValidBreakpoint(params.preceedingItem, params.item) != params.isValidBreakpoint {
				t.Errorf(
					"IsValidBreakpoint(%T, %T) != %t",
					params.preceedingItem,
					params.item,
					params.isValidBreakpoint,
				)
			}
		})
	}
}
