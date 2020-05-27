package pdb

import (
	"../util"
)

type IJournal interface {
	// each pdb has one journal
	Version() uint32                // Version
	ConsensusID() util.IConsensusID // Consensus ID
	Domain() string                 // Domain Name
	// operations
	Append(r []util.IRecord) error     // append a list of records
	AppendRecord(r util.IRecord) error // append a record
}

type JournalV1 struct {
}
