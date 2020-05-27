package pdb

import (
	"../util"
)

type IMemTable interface {
	// SSTable Level Attributes
	Version() uint32                // Version
	ConsensusID() util.IConsensusID // Consensus ID
	Domain() string                 // Domain Name
	Table() string                  // Table Name
	// operations
	Get(key, scheme util.IData) util.IRecord
	Set(record util.IRecord) error
	Groups(key, scheme util.IData) []string
	Keys(key, scheme util.IData) []util.IData
}
