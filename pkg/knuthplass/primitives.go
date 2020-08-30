package knuthplass

// Item is any element in the typesetting input: box, glue or penalty
type Item interface {
	Width() int64
	// EndOfLineWidth() int64
	Shrinkability() int64
	Stretchability() int64
	PenaltyCost() int64
	IsFlaggedPenalty() bool
}

// NewBox creates and returns a new box item
func NewBox(width int64) Item {
	return &box{width: width}
}

// IsBox returns true iff the item is a box item
func IsBox(item Item) bool {
	_, ok := item.(*box)
	return ok
}

type box struct {
	width int64
}

func (box *box) Width() int64 {
	return box.width
}

func (*box) Shrinkability() int64 {
	return 0
}

func (*box) Stretchability() int64 {
	return 0
}

func (*box) PenaltyCost() int64 {
	return 0
}

func (*box) IsFlaggedPenalty() bool {
	return false
}

// NewGlue creates and returns a new glue item
func NewGlue(width int64, shrinkability int64, stretchability int64) Item {
	return &glue{width: width, shrinkability: shrinkability, stretchability: stretchability}
}

// IsGlue returns true iff the item is a glue item
func IsGlue(item Item) bool {
	_, ok := item.(*glue)
	return ok
}

type glue struct {
	width          int64
	shrinkability  int64
	stretchability int64
}

func (glue *glue) Width() int64 {
	return glue.width
}

func (glue *glue) Shrinkability() int64 {
	return glue.shrinkability
}

func (glue *glue) Stretchability() int64 {
	return glue.stretchability
}

func (*glue) PenaltyCost() int64 {
	return 0
}

func (*glue) IsFlaggedPenalty() bool {
	return false
}

// NewPenalty creates and returns a new penalyu item
func NewPenalty(width int64, cost int64, flagged bool) Item {
	return &penalty{width: width, cost: cost, flagged: flagged}
}

// IsPenalty returns true iff the item is a penalty item
func IsPenalty(item Item) bool {
	_, ok := item.(*penalty)
	return ok
}

type penalty struct {
	width   int64
	cost    int64
	flagged bool
}

func (penalty *penalty) Width() int64 {
	return penalty.width
}

func (*penalty) Shrinkability() int64 {
	return 0
}

func (*penalty) Stretchability() int64 {
	return 0
}

func (penalty *penalty) PenaltyCost() int64 {
	return penalty.cost
}

func (penalty *penalty) IsFlaggedPenalty() bool {
	return penalty.flagged
}
