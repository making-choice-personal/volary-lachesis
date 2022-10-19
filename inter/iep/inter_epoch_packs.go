package iep

import (
	"github.com/making-choice-personal/volary-lachesis/inter"
	"github.com/making-choice-personal/volary-lachesis/inter/ier"
)

type LlrEpochPack struct {
	Votes  []inter.LlrSignedEpochVote
	Record ier.LlrIdxFullEpochRecord
}
