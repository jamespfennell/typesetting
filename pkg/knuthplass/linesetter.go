package knuthplass

type FixedItem struct {
	Visible bool
	Width   int64
}

type SetLineError struct {
	TargetLineLength int64
	ActualLineLength int64
}

func (err *SetLineError) Error() string {
	return "Item could not be set"
}

func (err *SetLineError) IsOverfull() bool {
	return err.TargetLineLength < err.ActualLineLength
}

func (err *SetLineError) IsUnderfull() bool {
	return !err.IsOverfull()
}

func SetLine(itemList *ItemList, lineLength int64) ([]FixedItem, *SetLineError) {
	return nil, nil
}
