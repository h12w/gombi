package dfa

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

type stateID int

const (
	invalidID      stateID = -1
	trivialFinalID stateID = -2
)

func (id stateID) valid() bool {
	return id >= 0
}
