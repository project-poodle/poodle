# Identity #

Poodle objects (e.g. Clusters, Nodes, Services, etc.) are identified by
[ECDSA](https://golang.org/pkg/crypto/ecdsa/) crypto keys.

- Each Poodle Cluster is identified by a cluster specific public / private
  ECDSA key.
- Each Poodle Node is identified by a node specific public / private
  ECDSA key.

A Poodle Node is added to a Poodle Cluster by a message containing the
__cluster:node__ scheme, the Poodle node public key, a 'UPDATE'
operation, and a timestamp, signed by the Poodle cluster private key.

A Poodle Node is removed from a Poodle Cluster by a message containing the
__cluster:node__ scheme, the Poodle node public key, a 'CLEAR'
operation, and a timestamp, signed by the Poodle cluster private key.



# Global Config #

Poodle global config is set by a message containing the specific config
information, a 'UPDATE' or 'CLEAR' operation, and a timestamp, signed by
the Poodle cluster private key.

Global config information are stored on all Poodle nodes.  All Poodle nodes
in the same cluster will replicate the entire global config with change logs.

Poodle global configs are associated with scheme: __cluster:conf__

Some global config Key examples are:

- raft.size
  - Suggested raft consensus size.
  - Actual raft consensus size is:

        min(21, max(3, raft.size))

- raft.quorum
  - Suggested raft quorum size.
  - Actual raft quorum size is:

        min(raft.size, max(ceil((raft.size + 2)/2), raft.quorum))



# Time Synchronization #

Poodle assumes all nodes are synced with each other on Unix time.  The assumption
does not require exact clock sync, and allows clock drift of 100s of milliseconds.

Each Poodle network message contains a timestamp of the source node.  When the
destination node received the message, it checks the timestamp against its own,
and if the timestamp difference is significantly high, the receiving node will
discard the message and log an error.

By default, poodle will accept time difference from another node with less than
300 milliseconds difference; reject time difference above 500 milliseconds; and
randomly chose to accept or reject packet from another node if time difference
is between 300 and 500 milliseconds.

These can be configured with following configs in __cluster:conf__ scheme:

- time.drift.min - effective drift min is:

      min(300, max(50, time.drift.min)

- time.drift.max - effective drift max is:

      min(500, max(100, time.drift.max, time.drift.min + 50))


### Leap Second ###

[Leap Second](https://en.wikipedia.org/wiki/Leap_second) is a one-second
adjustment that is occasionally applied to Coordinated Universal Time (UTC),
to accommodate the difference between precise time (as measured by atomic clocks)
and imprecise observed solar time.

For Distributed Ledger consensus, Poodle uses 30 seconds epoch, and can tolerate
leap second when it occurs.

For Raft consensus, Poodle uses [Unix monotonic clock](https://golang.org/pkg/time/),
and is not affected by the Leap Second.



# Consensus #

Poodle treats all nodes in a Poodle cluster as members on a hash ring. Poodle uses
the node public key to indicate the location of the node on the ring.

There are different types of consensus in Poodle.


### Distributed Ledger ###

Poodle cluster level consensus is a distributed ledger with following properties:

- Consensus by Proof of Stake, each valid member node has 1/N-th of voting share

- 30 seconds consensus time for each epoch

- epoch represented as unsigned int (4 bytes), possible life span ~4000 years

Poodle cluster level consensus keeps global state for the entire poodle cluster,
e.g.

- node membership
- global config parameters
- global status and stats
- current epoch
- global lookup schemes
- global compression schemes

Cluster level configs are published to all the nodes in the cluster, and are
replicated to all the nodes.


### Raft ###

Each consecutive segment of nodes (raft.size) on the hash ring in a Poodle
cluster forms a Raft consensus protocol.  Records are sharded, where each
Shard of data is stored in the corresponding Raft consensus data store.

The membership of each Raft consensus protocol is dynamically determined
by the location of the node on the hash ring.  E.g.

- if raft.size == 5, then for a specific node, itself, and 4 neighboring
  active nodes on the hash ring with location less than the current node
  are part of the same raft consensus

Records are distributed to the specific Shard on the hash ring by:

    SHA256( CONCAT(consensus_id, domain, table, key) )



# Proof of Stake #

Poodle distributed ledger consensus is established with Proof of Stake. 2/3
of the cluster members must sign a message for distributed ledger consensus.



# Raft Consensus #

### Raft Quorum ###

As Poodle cluster dynamically adds or removes nodes, the raft protocol for a data
shard will need to dynamically add and remove membership.  This poses additional
requirement to the raft consensus quorum.

If a raft consensus will need to add 1 or remove 1 member from the raft, it will
need more than the usual (N+1)/2 quorum in the consensus protocol.  E.g.

- to dynamically add 1 new nodes, a quorum will need (N+1+1)/2 nodes
- to dynamically add M more nodes, a quorum will need (N+1+M)/2 nodes

As raft node addition can be fast, adding and removing 1 node at a time can be
sufficient in most of the cases.  E.g. to add 2 new nodes and remove 2 existing
nodes, a raft consensus can sequentially add 1 new node, remove 1 existing node,
then add another new node, then remove another existing node.

When adding or removing max 1 node at a time in the raft consensus protocol, we
can derive the following:

| Raft Nodes | Quorum Size | Max Failure | Max Raft Nodes |
| :---: | :---: | :---: | :---: |
| 3 | 3 | 0 | 4 |
| 4 | 3 | 1 | 5 |
| 5 | 4 | 1 | 6 |
| 6 | 4 | 2 | 7 |
| 7 | 5 | 2 | 8 |
| 8 | 5 | 3 | 9 |
| 9 | 6 | 3 | 10 |

As in the above table, to tolerate 1 node failure, the minimum raft size is 4
nodes. To tolerate 2 node failures, minimum raft size is 6 nodes.


### Raft Membership ###

In Poodle, raft membership is dynamically formed.

All the nodes in a Poodle cluster knows all other nodes in the cluster as
the public keys of the other node.

Members in raft protocol can be in one of the 3 states:

* Leader
* Follower
* Candidate

To enable dynamically adding nodes to raft protocol, a new state is introduced:

* Learner

A learner learns from a raft consensus protocol all the historical state and
replicate the entire state and most recent change logs.

Leader sends raft messages to all the nodes, including the learner.  Learner
responds its status to the leader, indicating whether the Leander is up to date
with latest log replication from the Leader.

Once Leaner is up-to-date with the latest log replication, the Leader will decide
whether to turn the Leaner into a Follower.  In case there are more than on Leaner
up-to-date and ready, the Leader will only attempt to turn one of the Learners to
Follower.  The Leader sends a log message to itself and all the Followers about
the membership change, upon positive response by Quorum nodes, the Leaner is
formally a Follower.

If the Leader + Follower + Candidate size has reached Maximum Raft Nodes, the
Leader will choose a Follower to retire from the current raft consensus.  The
chosen node will be one of the node outside of the designated consensus for
the corresponding data shard.  A log message is appended to all the nodes
so that membership change is persisted.  If the Leader itself is to be removed,
the Leader will append the log to all nodes, waiting for the positive response,
and then stop itself from participating in the consensus.

The raft consensus will repeat the above steps, until the entire raft consensus
are running on the desired nodes for the data shard.



# Bootstrap #

During bootstrap, Poodle cluster may add new nodes, and potentially remove
existing nodes.


### Raft Consensus Identities ###

Node membership change directly impact Poodle cluster wrt how the hash ring is
splitted:

- When new node membership is added, Poodle will split corresponding
  Raft consensus group identity by introducing a new Raft consensus identity,
  then split node membership to serve the newly split Raft consensus identity.

- When existing node membership is removed, Poodle will merge corresponding
  Raft consensus identity, and merge the metadata from two Raft consensus
  identities into one.

Raft consensus group identity changes are processed 1 node at a time, and
is only processed after 3 confirmations of the cluster global consensus
protocol.

The Raft consensus identities change only when node membership changes
(config change). Node healthiness (status change) does not change the raft
consensus group identities.


### Raft Consensus Membership ###

Node healthiness status, when consistently detected in raft consensus group by
the leader after one full epoch, will be logged to the Raft consensus log, then
published to the cluster level consensus as __cluster.status:node__ scheme.

Poodle records node status as one of the 3 conditions:

- up
  - node is stable, and responds to more than 99% of the requests from
    leader.
- down
  - node is not responding, or responding to less than 1% of the requests
    from leader.
- flaky
  - node responds to between 1% and 99% of requests from leader.   

After 3 confirmation of 5 consecutive node healthiness as down or flaky, Poodle
will consider the node not eligible participating in the Raft consensus group,
and will be removed Raft consensus group membership.  The Raft consensus group
will pick the next node in the ring to form updated consensus group.

When a node come up again, it will announce itself to the corresponding Raft
consensus group, and act as learner.  After the node learned its knowledge,
Raft consensus identity leader will announce its healthiness to the global
cluster consensus.

After 3 confirmation of 5 consecutive node healthiness as up, Poodle will consider
the node eligible participating in the Raft consensus group, and will be
added to Raft consensus group membership for the corresponding Raft identity.

The healthiness detection, together with global cluster consensus build is
a 4-5 minutes process that avoids frequent Raft membership changes, and retains
Raft stability.



# Record #

Record consists of a scheme, key, value tuple encoded as the following.

A Record cannot exceed 64 KB, with following constraints:

- Key
  - Maximum key length is 4 KB
- Value
  - Maximum value length is 56 KB
- Scheme
  - Maximum scheme length is 2 KB


### Record Encode Magic ###

The first byte is a __magic__.

              Timestamp
                 and
              Signature
                 bit
    Key    Scheme |
    | |     | |   |
    7 6 5 4 3 2 1 0
        | |     |
       Value    |
              Clear
               bit

- Bit 7 and 6 are the key bits for encoding of key
  - 00 means no key
  - 01 means 1 byte to represent key length (up to 255 bytes)
  - 10 means 2 bytes to represent key length (up to 4 KB)
  - 11 means key is encoded with __data encoding__

- Bit 5 and 4 are the value bits for encoding of value
  - 00 means no value
  - 01 means 1 byte to represent value length (up to 255 bytes)
  - 10 means 2 bytes to represent value length (up to 56 KB)
  - 11 means value is encoded with __data encoding__

- Bit 3 and 2 are the value bits for encoding of scheme
  - 00 means no scheme
  - 01 means 1 byte to represent scheme length (up to 255 bytes)
  - 10 means 2 bytes to represent scheme length (up to 2 KB)
  - 11 means scheme is encoded with __data encoding__

- Bit 1 is the clear bit
  - 1 means 'CLEAR' operation
  - 0 means 'UPDATE' operation

- Bit 0 is the timestamp and signature bit
  - 1 means there is a timestamp and signature at the end of the record
  - 0 means no timestamp or signature at the end of the record
  - Signature is present only in Distributed Ledger Network Traffic and
    Distributed Ledger Consensus Block
  - Signature is __not__ present for SSTable (Distribute Ledger or Raft
    Consensus), nor Raft replication log.
    - In these cases, even if this bit is set to 1, only timestamp is present
      with the record
  - source content of the signature include binary consensus ID, concatenated
    with Record content, including Record Magic, Key, Value, Scheme, and
    Timestamp
  - If signature is present, the content of data are in raw format, and
    cannot be encoded with lookup scheme, or compression scheme


### Scheme Format ###

Scheme consists of the following components:

- Consensus ID
  - This is Consensus Identity
  - This ID is in binary, encoded as Consensus ID format
  - Consensus ID is required as part of full Scheme

- Domain
  - This is similar to database or NoSQL schema
  - Domain name is an alpha-numeric string separated by '.'
  - Domain name is required as part of the standard Scheme

- Table
  - This is similar to database or NoSQL table
  - Each row is identified by a unique key
  - Table name is an alpha-numeric string separated by '.'
  - Table name is required as part of the standard Scheme

- Attribute Group
  - This is similar to column group in a NoSQL table
  - Attribute Group name is an alpha-numeric string separated by '/'
  - Attribute Group name is optional part of the Scheme

A full Scheme is joined by Consensus ID, with Domain, Table, and optional
Attribute Group. E.g.

    <consensus_id>, <domain>:<table>[/<attribute-group>]

Scheme in Record encoding is encoded differently when transmitted via network,
or when stored on disk.  Not all Scheme components are stored in the Record
in all format.

- When transmitted via network, Consensus ID is NOT encoded as part of the
  Record.  Instead, Consensus ID is encoded as part of Consensus Block as
  defined in P-UDP Packet encoding.

- When stored on disk, Consensus ID, Domain, and Table are NOT encoded as
  part of the Record.  Instead, Consensus ID, Domain, and Table are encoded
  as part of SSTable header as defined in SSTable encoding.

Examples below:

| Scheme | Consensus ID | Domain | Table | Attribute Group |
| :--- | :--- | :--- | :--- | :--- |
| \<C\>, cluster:conf                           | Cluster ID | cluster consensus | cluster conf table | base attribute group |
| \<C\>, cluster:node                           | Cluster ID | cluster consensus | node table | base attribute group for membership |
| \<C\>, cluster:node/conf                      | Cluster ID | cluster consensus | node table | conf attribute group |
| \<C\>, cluster:spaceport                      | Cluster ID | cluster consensus | space port table | base attribute group for membership |
| \<C\>, cluster:spaceport/conf                 | Cluster ID | cluster consensus | space port table | conf attribute group |
| \<C\>, cluster.status:node                    | Cluster ID | cluster status | node table | base attribute group for status |
| \<C\>, cluster.status:node/stats              | Cluster ID | cluster status | node table | stats attribute group |
| \<C\>, cluster.status:spaceport               | Cluster ID | cluster status | space port table | base attribute group for status |
| \<C\>, cluster.status:spaceport/stats         | Cluster ID | cluster status | space port table | stats attribute group |
| \<C\>, \<S\>, \<E\>, raft:poodle              | Cluster ID, Shard Start, Shard End | Raft consensus | poodle table | base attribute group for poodle metadata service |
| \<C\>, \<S\>, \<E\>, raft:poodle.status       | Cluster ID, Shard Start, Shard End | Raft consensus | poodle status table | status attribute group |
| \<C\>, \<S\>, \<E\>, raft:poodle.status/stats | Cluster ID, Shard Start, Shard End | Raft consensus | poodle status table | stats attribute group |


### Record Encoding ###

A full record is encoded as following:

                                      Scheme Length
                      Value Length     | |
      Key Length       | |             | |             8 bytes timestamp
       | |             | |             | |             |     |
     X X X X ... ... X X X X ... ... X X X X ... ... X X ... X X ... ... X
     |     |         |     |         |     |         |         |         |
     |     Key Content     |         |     |         |         64 bytes signature
     |                    Value Content    |         |
     |                                    Scheme Content
    Record
    Magic

- Lead by a __magic__ byte

- Followed by key length, then key content (if applicable)

- Followed by value length, then value content (if applicable)

- Followed by scheme length, then scheme content (if applicable)

- Followed by timestamp (8 bytes) and signature (32 bytes) (if applicable)
  - The 64 bytes signature is a sizable data, and is present in:
    - Network traffic for Distributed Ledger Consensus
    - Consensus block data for Distributed Ledger Consensus
  - The 64 bytes signature is __not__ present in:
    - SSTable for Distributed Ledger or Raft
    - Replication log for Raft


### Data Encode Magic ###

Data encoding can significantly reduce size of the data by representing
data as a lookup, or in compressed format.

When the record encoding bits are __11__ for key, value, or scheme, this
indicates the data follows __data encoding__. In this case, the first byte
of the data encoding is another magic that represents __data encoding
magic__.

             Compression
                |
         Record |
          List  |
          | |   |
      7 6 5 4 3 2 1 0
      | |     |   | |
     Data     |   | |
     Array    |  Length
              |
            Lookup

- Bit 7 and 6 are Data Array bits
  - 00 means not a Data Array
  - 01 means 1 byte of array element count
  - 10 means 2 bytes of array element count
  - 11 is reserved
  - The Data Array bits cannot be set when Record List bits are also set

- Bit 5 and 4 are Record List bits
  - 00 means not a Record List
  - 01 means 1 byte of list element count
  - 10 means 2 bytes of list element count
  - 11 is reserved
  - The Record List bits cannot be set when Data Array bits are also set

- Bit 3 is lookup scheme bit
  - 0 means no lookup scheme
  - 1 means 2 bytes of lookup scheme

- Bit 2 is compression scheme bit
  - 0 means no compression scheme
  - 1 means 2 bytes off compression scheme

- Bit 1 and 0 are length of data length
  - When lookup bit is 1, this 2 bits represent lookup data length, not length
    of data length
  - 00 means 0 length
  - 01 means 1 byte length
  - 10 means 2 bytes length
  - 11 is reserved

Note:

- The following encoding schemes are mutually exclusive:
  - Data Array
  - Record List
  - Lookup Scheme
  - Compression Scheme

- A properly encoded Data can have only one of the encoding schemes from above.

- If none of these encoding scheme are set, and if Data is encoded inside a Record,
  then Data content will be normalized to become as part of Record encoding.


### Data Encoding ###

A full __data encoding__ is as following:

    Data
    Magic         Length
     |    Lookup   | |
     |     | |     | |
     |     | |     | |
     X X X X X X X X X X ... ... X
       | |     | |     |         |
       | |     | |     Data Content
      Data     | |
      Array    | |
       or    Compression
      Record
      List
      Count

When data size is relatively small (less than ~1k), and when possible
enumeration of data content is limited, lookup can be an effective
way of reducing the data size.

A poodle consensus keeps a list of cluster wide lookup schemes.
The list of lookup schemes are registered across the cluster, and is
specific to a consensus.  The cluster wide lookup scheme and can be
used to encode data.  E.g.

- A 256 bits ECDSA public key is 32 bytes long.  Sending 32 bytes
  over the wire, or store on disk can represent a significant overhead.

- Instead, if we have a lookup scheme that will lookup the encoded data
  for original content of the data, this can significantly reduce the
  data size to represent an ECDSA public key.

- Assume there are total 10k nodes (10k possible public keys), a perfect
  hash and 2 bytes lookup key will be enough to represent an ECDSA public
  key.

- Considering we will need to continuously evolving lookup schemes (e.g.
  when new nodes are added, and old removed, the lookup scheme will need
  to be updated), we will need to record a list of actively used schemes.

- Assume 1 byte to represent scheme, and 2 bytes to represent data content,
  total encoding length of a 32 bytes ECDSA public key is: 1 magic byte +
  1 lookup scheme byte + 0 compression scheme byte + 0 length byte + 2
  content bytes = 4 bytes.  This is 87.5% reduction of data size.

When data size is relatively large (larger than ~1k), and when data is not
already compressed, compression can be an effective way to reduce the
data size.

A poodle consensus keeps a list of cluster wide compression schemes.
The list of schemes are registered across the cluster, and can be
used to encode data.



# P-UDP Packet #

Poodle uses UDP Packet for fast metadata operations.

P-UDP stands for Poodle UDP.  A P-UDP packet consists of the sender's
node ID, followed by a list of Request(s) and Response(s), followed
by the sender's Timestamp and Signature.


### Packet Encoding ###

A P-UDP Packet is encoded below:

                   Consensus
    Node ID          Block                  8 bytes timestamp
    |     |         |     |                  |     |
    X ... X X ... X X ... X ... ...  X ... X X ... X X ... ... X
            |     |                  |     |         |         |
           Consensus                Consensus        32 bytes signature
             Block                    Block

Poodle Packet is constructed as UDP packet, a Poodle packet must be less
than 64KB.

Poodle Node gathers multiple requests and responses in a buffer during
a very short time period (e.g. 1-20ms), then send the aggregated requests
and responses to the destination Node and Port.  If destination buffer
exceeded a predefined threshold (default 8KB), Poodle will send the content
of the buffer without waiting for the timer.

- Node ID
  - Node ID is encoded with DATA

- A list of Consensus Blocks
  - a list of Consensus Blocks as in the Consensus Block encoding

- Timestamp
  - 8 bytes timestamp represent node own timestamp

- Signature
  - 32 bytes signature covers a list of requests and responses and the
    timestamp


### Consensus ID Magic ###

         Federation
            bit
             |
    Universe | Shard
        bit  |  bit
         |   |   |
         7 6 5 4 3 2 1 0
           |   |   |   |
       Cluster |  Reserved
          bit  |
               |
            Service
              bit

- Bit 7 is Universe bit
  - 1 means Universe ID is present
  - 0 means no Universe ID

- Bit 6 is Cluster bit
  - 1 means Cluster ID is present
  - 0 means no Cluster ID

- Bit 5 is Federation bit
  - 1 means Federation ID is present
  - 0 means no Federation ID

- Bit 4 is Service bit
  - 1 means Service ID is present
  - 0 means no Service ID

- Bit 3 is Shard bit
  - 1 means Shard Start and Shard End are present
  - 0 means no Shard start or Shard End

- Bit 2, 1, and 0 are reserved


### Consensus ID Encoding ###

                                           Shard
          Universe       Federation        Start
          |     |         |     |         |     |
        X X ... X X ... X X ... X X ... X X ... X X ... X
        |         |     |         |     |         |     |
    Consensus     Cluster         Service          Shard
       ID                                           End
      Magic

A Consensus ID is encoded as:

- Consensus ID Magic
- Followed by optional Universe ID
- Followed by optional Cluster ID
- Followed by optional Federation ID
- Followed by optional Service ID
- Followed by optional Shard Start and Shard End ID


### Consensus Block Encoding ###

                       Request
          Operation      or
           Count       Response
             |         |     |
     X ... X X X ... X X ... X ... ... X ... X
     |     |   |     |                 |     |
    Consensus  Request                 Request
       ID        or                      or
               Response                Response

A Consensus Block is encoded as:

- Consensus ID
- Followed by 1 byte Operation Count (maximum 255)
- Followed by a list of Request or Response


### Request and Response Magic ###

              test
               |
    request    | error
      bit  ops |   |
       |   | | |   |
       7 6 5 4 3 2 1 0
         |       |   |
      response   |  reserved
        bit      |
                test
               millis

- Bit 7 is Request bit
  - 1 means this is a request
  - 0 means this is not a request

- Bit 6 is Response bit
  - 1 means this is a response
  - 0 means this is not a response

- Bit 5 and 4 are ops bits
  - 00 means GET
    - this gets the specified key of specific Consensus ID, Domain, Table, and
      Attribute Group
  - 01 means SET
    - this sets value of the specified key of specific Consensus ID, Domain,
      Table, and Attribute Group. Both UPDATE and CLEAR Records are considered SET
  - 10 means GROUPS
    - this retrieves a list of Attribute Groups under the specified Key of
      specific Consensus ID, Domain and Table
  - 11 means KEYS
    - this retrieves a list of Keys with the specified Key as prefix, of specific
      Consensus ID, Domain and Table

- Bits 3 is test bit
  - Test bit enables atomic operation for handling of locked operation
  - When ops is POST (bits 4 and 5 are 01), this will test if a key
    matches the specified value and then UPDATE or CLEAR the key. e.g.
  - Set __value__ to __v1__ for __scheme=d1__, __key=k1__ if this value
    is not already set:
    - ops=POST, test=TEST, record=SET, scheme=d1, key=k1, value=v1
    - if a value is already set, this operation will return an error
    - if a value is not set, this operation will set the value, and will
      return success
  - Clear __value__ for __scheme=d1__, __key=k2__ if value is already
    set to __v2__:
    - ops=POST, test=1, record=SET, scheme=d1, key=k1, value=v2
    - if a value is not set, this operation will return success
    - if a value is set, but value is not __v2__, this operation
      will return an error
    - if a value is set, and value is __v2__, this operation will
      clear the value, and return success

- Bit 2 is test millis bit
  - This bit is valid only if both request bit and test bit are 1
  - 1 means a test milliseconds (4 bytes unsigned integer) is added to
    the end of the request
    - This test milliseconds is checked against record timestamp
    - If record timestamp not exist, this operation will return error
    - If record timestamp is newer than test milliseconds ago, this
      operation will perform the checks as normal test bit will do
      for UPDATE and CLEAR records
    - If record timestamp is older than test milliseconds ago, this
      operation will treat the value as if it is already cleared, and
      will not perform the test checks
  - 0 means no test milliseconds at the end of the request

- Bit 1 is error bit
  - This bit is valid only if response bit is 1
  - 1 means error occurred
    - When error occurred, the record value field is the error content
  - 0 means no error

- Bit 0 is reserved


### Request and Response Encoding ###

A Request and Response is encoded with a Magic, followed by record
content, followed by optional test millis:

     Request
       and
     Response
      Magic           Test Millis (optional)
        |             |     |
        X X ... ... X X ... X
          |         |
          Record Content



# SSTable #

Poodle stores Records as SSTable.

A few properties of SSTable file:

- SSTable is immutable - once written, it is never modified
  - SSTables are merged with other SSTables
  - SSTables can be removed once merged and not longer needed
  - SSTables content are never changed

- Each SSTable is limited to no more than 1048576 (1M) Records

- Each SSTable is limited to maximum 1 GB size

- Each Record is limited to less than 64 KB


### Record Scheme and SSTable ###

Poodle treats different portion of the Record Scheme separately:

- Consensus ID
  - Each Consensus ID are stored with its own directory structure
  - Different Consensus IDs are always stored separately
  - Consensus ID information determines directory name of the storage files
  - Consensus ID information is stored as the header of the storage files
  - Consensus ID information is removed from Scheme when the Record is
    stored in SSTable

- Domain
  - Each Domain is stored as separate directory structure under the
    Consensus ID directory
  - Different Domains are always stored separately
  - Domain information determines directory name of the storage files
  - Domain information is stored as the header of the storage files
  - Domain information is removed from Scheme when the Record is
    stored in SSTable

- Table
  - Each Table is stored as separate directory structure under the
    Domain directory
  - Like LevelDB and RocksDB, Poodle Table is a leveled structure of
    multiple SSTables at each level.
  - MemTables are flushed to L0 table
  - 10 L0 tables merges into one L1 table
  - 10 L1 tables merges into one L2 table
  - 10 L2 tables merges into one L3 table
  - 10 L3 tables merges into one L4 table
  - ...
  - Table information determines directory name of the storage files
  - Table information is stored as the header of the storage files
  - Table information is removed from Scheme when the Record is
    stored in SSTable

- Attribute Group
  - All Attribute Groups of the same Consensus ID, same Domain, and same
    Table are stored in the same groups of SSTable
  - Attribute Group information is stored in the Scheme field of a Record
    in SSTable


### SSTable Structure ###

A SSTable consists of:

- SSTable header:
  - Consensus ID
  - followed by domain name and table name
  - followed by table level (L0, L1, L2, L3, L4 ...)
  - followed by start and end time
    - for distributed ledger consensus, time is represented as Epoch #
    - for raft consensus, time is represented as Term + milliseconds
      elapsed + Record count in the same millisecond
  - followed by file start key
  - followed by crc32 of the header

- followed by Record Offset lookup:
  - the hash scheme for Record lookup
  - followed by the Record offset and length table
  - followed by crc32 of the offset lookup

- followed by a list of Records
  - each record followed by crc32 of the Records


### Record Offset Lookup ###

For each SSTable, Poodle generates a Perfect Hash for fast lookup
of Record in the file.

The Hash Key is composed by:

    <key_bytes> + 0x00 + <attribute_group_bytes>

Each SSTable file is less than 4GB, the record offset can be represented as
an uint32.  A 256 KB (64k * 4 bytes) offset table is enough to keep the offset
of all Records in an SSTable file.


### List of Records ###

SSTable uses Sort Key (same as Hash Key) to sort all the Records in a SSTable
and store the sorted Records in the file:

    <key_bytes> + 0x00 + <attribute_group_bytes>


### Compaction ###

SSTables are compacted to the next level when current level reaches 2 * 10
tables.

Compactions are performed on 10 tables - this keeps the time of compaction
stable after the compaction.

This compaction scheme is to enable speedy data replication from the client.



# Large Data Size #

While Poodle metadata service has __strict__ limitation on the Record size
and Data size.  Poodle metadata clients can support larger data size using
various techniques.

Note a Poodle Record cannot exceed 64 KB, with following constraints:

- Key
  - Maximum key length is 4 KB
- Value
  - Maximum value length is 56 KB
- Scheme
  - Maximum scheme length is 2 KB

Below are examples to support very large data size.


### Directory to File Mapping ###

For a highly scallable distributed file system, to support a directory
with millions of direct child files, a design as follows:

- __poodle.fs:inode__
  - this table keeps all inode information
    - table key is 8 bytes inode id (64 bits)
    - an inode can be a directory, or a file
  - [inode information](http://man7.org/linux/man-pages/man7/inode.7.html) includes:
    - file type (4 bits) and mode (12 bits)
    - UID, GID (2 * 4 bytes)
    - file size (8 bytes)
    - timestamps (4 * 8 bytes)
    - block information (optional)
    - extended attributes (optional)

- __poodle.fs:dir.filelist__
  - this table keeps all directories to files mappings

- each file metadata is 320 bytes (or less)
  - this keeps only core metadata information
  - the metadata is encoded as Record
    - Key
      - filename is Key
      - up to 255 bytes filename
    - Value
      - 8 bytes inode id
      - file type (4 bits) and mode (12 bits)
      - UID, GID (2 * 4 bytes)
      - file size (8 bytes)
      - timestamps (4 * 8 bytes)

- each attribute group stores metadata for up to 64 files
  - 320 * 64 = 20KB

- base table key is 8 bytes inode id for the directory
  - maximum 256 attribute groups from 0x00 to 0xff under each key
  - each shard can store up to 64 * 256 = 16K file metadata

- __one__ byte extended key can be appended to the table key
  - a bit map of 32 bytes (256 bits) is stored with base table key
    to indicate which child key is used
  - each extended key can have up to 256 attribute groups similar to
    the base key
  - each extended key adds metadata for another 16K files
  - one byte key extension supports up to 4M files directly under one
    directory

- extended key can be further appended to already extended keys
  - this further extends number of files directly under a same directory
    to unlimited.

By encoding with __keys__ and __attribute groups__, this file system can
support unlimited files directly under a directory.

Characteristics with this design:

- This design is efficient for directory scanning batch operations,
  such as __ls__, and __find__.

- Lookup by filename can suffer for very large directories. e.g. for a
  8M file directory (exteremly large), lookup a specific filename will
  require scanning 512 shards, or up to 2.4GB of metadata.

- Lookup by filename for small directories (e.g. less than 16K files)
  can be reasonably fast as the operation scans only one shard, with
  up to 5MB of metadata.


### File to Block Mapping ###

Another design is file to block mapping.  A very large file can have
millions of blocks.  The below design keeps one Record for each block
mapping, and store all block level data with this Record:

- __poodle.fs:file.block__
  - this table keeps all files to blocks mappings
  - file to block mapping is encoded to the block boundary
  - block index information is encoded as part of the key and attr groups
    - e.g. a 4KB block information will be stored in a Record with:
      - Key
        - 8 bytes file inode id + 6 bytes block prefix
      - Attribute Group
        - 4 bits block prefix as attribute group id (masked with 0xf0)

- each block stores metadata for __one__ block:
  - each block metadata can be up to 255 bytes
    - up to 16 container ids, each 8 bytes (up to 128 bytes)
    - 2 bytes block group id
    - 2 bytes block id
    - 8 bytes start offset
    - 4 bytes length
    - other block level attributes


| block size    | key size  | attr group size   |
| :---          | :---      | :---              |
| 1KB           | 6 bytes   | 6 bits            |
| 4KB           | 6 bytes   | 4 bits            |
| 16KB          | 6 bytes   | 2 bits            |
| 64KB          | 5 bytes   | 8 bits            |
| 256KB         | 5 bytes   | 6 bits            |
| 1MB           | 5 bytes   | 4 bits            |
| 4MB           | 5 bytes   | 2 bits            |



# Data Synchronization #

Poodle Metadata Service ensures atomic operation at Record level, but does
not guarantee data synchronization among multiple records.

A Poodle Metadata client can often make updates to multiple Poodle Metadata
Records to a complete transactional operation. A design for such use case
is to sequence the Record level operation properly, and only to complete
the transaction (write the final Record) when all the required steps have
completed.

This usage pattern can leave some unsynchronized metadata (and data) with
the system.  Some techniques to work with this pattern:


### Sequenced Operations ###

__Sequence__ the object creations: child first, then parent, then
grandparent. Writes root records after all children are created
successfully.

If a sequence has failed or interrupted in the middle, this will leave
__orphaned objects__.

In this case, running __garbage collection__ regularly will clean up
the orphaned objects.

This pattern can be used with __Copy on Write__ file system, where
garbage collection is a standard operation.

This can be a __common pattern__ for writing multiple Records to
fulfill a complete operation.

- One example is adding new blocks to a file, where two Records are
  required for the operation, one Record on __poodle.fs:block__, and
  another Record to __poodle.fs:file.blocklist__.

  - In this case, code will update __poodle.fs:block__ Record (child),
    then update __poodle.fs:file.blocklist__ (parent).
  - If code failed in bewteen, the __poodle.fs:block__ is orphaned,
    and will be cleaned up during next garbage collection.

- Another example use case is __link__ operation in a distributed file
  system that creates hard link between a directory and a file, where
  two Records, one on __poodle.fs:dir.filelist__, and another o
  __poodle.fs:file__ are required for completion of the operation.


### Caching ###

This use case is when a Cached copy of the Authoritative Data is kept
in alternative format for faster access purpose, where writing two or
more Records are needed to make the system consistent.

E.g. in a distributed file system:

- poodle.fs:file
  - This table keeps the authoritative data on a file's attributes
    - access permissions
    - timestamps (create, modify, access)
    - etc.
- poodle.fs:dir.filelist
  - This table keeps a copy of the file attributes for fast access
    - e.g. for __ls__, __find__, or similar operations

In this design, keep the cached copy in parent object, and regularly
sync between child object and parent object is needed.

This pattern should be use only __when necessary__.  Refresh of
cached data can be costly operation (when performed regularly),
or error prone to implement (when triggered by event).

A __alternate design__ can keep the file information at the parent
directory level, and do not keep the information at file level.  In
this design, __ls__ and __find__ do not incur additional cost, while
one extra access is required when trying to access the file itself.


### Journal Record ###

Write a journal Record prior to writing a set of separate Records that
can be represented by this one journal Record.

The set of separate Records are usually in a format that can be easily
accessed by direct access methods.

In this design, the failed updates to Records after a successful Journal
Record will repeatedly try until all Records are written successfully.

This is eventual consistency model, and can have ramifications when
the retries are handled across shards, with potential conflicting
write operations happening in between.

This should be use only with __extreme caution__.



# Service #

Like Poodle Node and Cluster, a Poodle Service is identified by crypto
keys.

A Poodle Service can consists a set of nodes together offering services
to its clients.  E.g.

- Poodle POSIX File System Service
  - This is a distributed POSIX compliant file system service

- Poodle Key-Value Store Service
  - This is a distributed Key-Value service

- Poodle Metadata Service
  - This service is provided as part of Poodle core

Each Poodle Cluster has one and only one Metadata Service.

A Poodle Cluster can have zero or more other services, such as POSIX
File System Service(s), and Key-Value Store Service(s).

While a Poodle Cluster offers Service(s), from time to time, the Cluster
may make changes to the Service(s), e.g.:

- Move a Poodle Service to from one Poodle Cluster another Poodle Cluster

- Enable Federated Services running across multiple Poodle Clusters

These operations further extend operability of Poodle Cluster(s) and
Poodle Service(s).  E.g.

- If an organization decided to segregate multiple services that were
  running on a single cluster, into two separate clusters for ease of
  future management, the operator can create another Poodle Cluster,
  and move the selected Poodle Service(s) to the other Poodle Cluster
  without disruption to a running production Poodle Service.

- Moving service from one set of hardware to another set of hardware.
  This use case can be supported similar to the earlier case, by creating
  separate Poodle cluster on new hardware, and move the service over
  to the new cluster running on new hardware.  The entire operation
  can happen with live production traffic.

- Setup Poodle Service Federation across Poodle Cluster(s).  All the
  nodes in a Poodle Cluster is usually co-located in the same data
  center.  There can be needs to run services across data centers.
  In this case, Poodle Service Federation can run across multiple
  Poodle Clusters that serves the clients from federated service(s).



# Nodes, Clusters, and Universe #

### Clusters and Nodes ###

Poodle Node and Cluster are related by many-to-many relationship.

Naturally, a Poodle Cluster consists of many nodes.  These nodes
forms consensus within the Cluster.

Similarly, a Poodle Node can belong to more than one Poodle Cluster.

- Node Membership
  - To add a Node to a Cluster, the Node sends a JOIN request,
    signed by Node private key.  Upon receiving the request, the
    Cluster can accept or reject by signing with Cluster private
    key.
  - To remove a Node from a Cluster, the Cluster sends a CLR
    request, signed by Cluster private key.

- Node Attributes
  - To update Node Attributes, such as IP Addr and Port Num, a
    Node signs a message with updated attributes, broadcast to
    the cluster.  Config is updated to other nodes once the
    Cluster Consensus accepts to the change.

A pre-requisite for a Poodle Node to belong to multiple cluster is:

- A Poodle Node can belong to multiple cluster if-and-only-if these
  clusters are in the same Poodle Universe.

- Neither Poodle Node, nor Poodle Cluster can cross multiple Poodle
  Universe.


### Universe ###

A Poodle Universe is formed from multiple Poodle Clusters.

Like Poodle Node and Poodle Cluster, a Poodle Universe is identified
by crypto keys.

Each Poodle Cluster can designate a set of nodes as Space-Port.
The Space-Port nodes from all Poodle Clusters in the same Universe
connects to each other and participates in Poodle Universe Consensus.

A Poodle Universe Consensus requires 2/3 of the Clusters to verify
the message.  For a Cluster to verify the message, 2/3 of the
Space-Port nodes must verify the message.

The Space-Port nodes will share the Poodle Universe consensus
with the Poodle Cluster, all of the Poodle Node keeps a copy
of Poodle Universe Consensus.

- Cluster Membership
  - To add a Cluster to a Universe, the Cluster sends a JOIN
    request, signed by Cluster private key.  Upon receiving the
    request, the Universe can accept or reject by signing with
    Universe private key.
  - To remove a Cluster from a Universe, the Universe sends a CLR
    request, signed by Universe private key.

- Space-Port Membership
  - To assign Space-Port, the Cluster signed a request to record
    Node as Space-Port.  The Space-Port Node broadcast the signed
    message to the Universe.  Space-Port Membership is accepted
    when Universe Consensus accepts the Node as Space-Port.
  - To un-assign Space-Port, the Cluster sign a request to un-assign
    Node as Space-Port.

- Trust Relationship
  - Poodle Clusters in the same Universe can establish trust
    relationship.
  - Trust relationship is established by Cluster 1 generate a TRUST
    request to Cluster 2, with Cluster 2 signing the request to
    accept the TRUST.  Trust is established when the accepted request
    is accepted by Universe Consensus.
  - Trust relationship is mutual - e.g. Cluster 1 trust Cluster 2
    means Cluster 2 also trust Cluster 1.
  - Trust relationship is not transferable. Cluster 1 trust Cluster 2,
    and Cluster 2 trust Cluster 3, this does not mean Cluster 1 trust
    Cluster 3  



# Multiverse #

Multiverse is not supported
