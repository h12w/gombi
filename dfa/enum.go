package dfa

const (
	invalidID      = -1
	trivialFinalID = -2
)

type finalLabel int

const (
	notFinal          finalLabel = 0
	defaultFinal      finalLabel = 1
	labeledFinalStart finalLabel = 2
)

func (l finalLabel) final() bool {
	return l >= defaultFinal
}

func (l finalLabel) labeled() bool {
	return l >= labeledFinalStart
}

func (l finalLabel) toInternal() finalLabel {
	return l + labeledFinalStart
}

func (l finalLabel) toExternal() int {
	if l >= labeledFinalStart {
		return int(l - labeledFinalStart)
	}
	panic("machine is not labeled")
}
