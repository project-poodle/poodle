package pdb

import (
	"../util"
)

type IPdb interface {
	// each pdb has one journal
	Version() uint32                // Version
	ConsensusID() util.IConsensusID // Consensus ID
	Domain() string                 // Domain Name
	// operations
	Get(table string, key util.IData, group string) util.IRecord
	Set(table string, record util.IRecord) error
	Groups(table string, key, scheme util.IData) util.IData
	Keys(table string, key, scheme util.IData) util.IData
}

type PdbV1 struct {
}
