package util

const (
	// Each EPOCH is 30 seconds
	POODLE_EPOCH_MILLIS = 30 * 1000

	DEFAULT_DRIFT_MILLIS_LOW  = 300
	DEFAULT_DRIFT_MILLIS_HIGH = 500

	DEFAULT_ETC_DIR = "/etc/poodle"
	DEFAULT_LIB_DIR = "/var/lib/poodle"
	DEFAULT_LOG_DIR = "/var/log/poodle"

	DEFAULT_SECRET = "poodle"

	DEFAULT_PUDP_PORT = 31415
	DEFAULT_QUIC_PORT = 31416

	MAX_KEY_LENGTH    = 4 * 1024    // Maximum  4 KB Key Length
	MAX_VALUE_LENGTH  = 56 * 1024   // Maximum 56 KB Value Length
	MAX_SCHEME_LENGTH = 1 * 1024    // Maximum  1 KB Scheme Length
	MAX_ATTR_GROUPS   = 256         // maximum 256 Attribute Groups per Key
	MAX_DATA_LENGTH   = 64*1024 - 1 // Maximum 64 KB - 1 Data Length
	MAX_PACKET_LENGTH = 64*1024 - 1 // Maximum 64 KB - 1 Packet Length

	CLS_NODE       = 1
	CLS_CLUSTER    = 2
	CLS_UNIVERSE   = 3
	CLS_SERVICE    = 4
	CLS_FEDERATION = 5

	TRIE_NODE_TYPE_KEY        = byte(0x01) // trie node type key
	TRIE_NODE_TYPE_ATTR_GROUP = byte(0x02) // trie node type attr group
)
