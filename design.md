# Identity #

Poodle clusters and nodes are identified by crypto keys.

- Each Poodle cluster is identified by a cluster specific public / private
  ECDSA key.
- Each Poodle node is identified by a node specific public / private
  ECDSA key.

A Poodle node is added to a Poodle cluster by a message containing the
__poodle.cluster.membership__ domain, the Poodle node public key, a 'UPDATE'
operation, and a timestamp, signed by the Poodle cluster private key.

A Poodle node is removed from a Poodle cluster by a message containing the
__poodle.cluster.membership__ domain, the Poodle node public key, a 'CLEAR'
operation, and a timestamp, signed by the Poodle cluster private key.


# Global Config #

Poodle global config is set by a message containing the specific config
information, a 'UPDATE' or 'CLEAR' operation, and a timestamp, signed by
the Poodle cluster private key.

Global config information are stored on all Poodle nodes.  All Poodle nodes
in the same cluster will replicate the entire global config with change logs.

Poodle global configs are associated with domain: __poodle.config__
 
Some global config examples are:

- poodle.raft.size
  - Suggested raft consensus size.
  - Actual raft consensus size is:
    - min(21, max(4, poodle.raft.size))
- poodle.raft.quorum
  - Suggested raft quorum size.
  - Actual raft quorum size is:
    - min(poodle.raft.size, max(ceil((poodle.raft.size + 2)/2), poodle.raft.quorum))


# Time Synchronization #

Poodle assumes all nodes are synced with each other on unix time.  The assumption
does not require exact sync, and allow drift of 100s of milliseconds of drift.

Each Poodle network message contains a timestamp of the source node.  When the
destination node received the message, it checks the timestamp against its own,
and if the timestamp difference is significant high, the receiving node will
discard the message and log the error.

By default, poodle will accept time difference from another node with less than
400 milliseconds difference; reject time difference above 600 milliseconds; and
randomly chose to accept or reject packet from another node if time difference
is between 400 and 600 milliseconds.

These can be configured with following configs in __poodle.config__ domain:

- poodle.time.drift.min
  - min(800, max(100, poodle.time.drift.min)
- poodle.time.drift.max
  - min(1000, max(300, poodle.time.drift.max))


# Consensus #

Poodle treats all nodes in a Poodle cluster as members on a hash ring. Poodle uses
the node public key to indicate the location of the node on the ring.

There are two types of consensus in a Poodle cluster:

### Cluster Consensus ###

Poodle cluster level consensus keeps global state for the entire poodle cluster, e.g.

- node membership
- global config parameters
- current epoch
- compaction epoch
- global lookup schemes
- global compression schemes

Cluster level configs are published to all the nodes in the cluster, and are
replicated to all the nodes.

Poodle cluster level consensus is a distributed ledger with following properties:

- Consensus by Proof of Stake, each valid member node has 1/N-th of voting share
- 30 seconds consensus time for each epoch
- epoch represented as unsigned int (4 bytes), possible life span ~4000 years
 
### Data Segment Consensus ###

Each segment (poodle.raft.size) of the Poodle cluster on the hash ring
forms a Raft consensus protocol, and keeps a segment of data in a
distributed key/value store.

The membership of each raft consensus protocol is dynamically determined
by the location of the node on the cluster.  E.g.

- if poodle.raft.size == 4, then for a specific node, itself, and 3 neighbor
  active nodes on the ring with location less than the current node are part
  of the same raft consensus

The key value store is distributed to the ring and specific segmented by: 

- (hash value of the key) XOR (hash value of the domain)


# Proof of Stake #

Poodle cluster wide consensus is established with Proof of Stake. 2/3 of the
cluster members must sign a message for cluster wide consensus.


# Raft #

### Raft Quorum ###

As Poodle cluster dynamically adds or removes nodes, the raft protocol for data
segments will need dynamically add and remove membership.  This poses additional
requirement to the raft consensus quorum.

If a raft consensus will need to add 1 or remove 1 member from the raft, it will
need more than the usual (N+1)/2 quorum in the consensus protocol.  E.g.

- to dynamically add 1 new nodes, a quorum will need (N+1+1)/2 nodes
- to dynamically add M more nodes, a quorum will need (N+1+M)/2 nodes

As raft node addition can be fast, adding and removing 1 node at a time can be
sufficient for most of the cases.  E.g. to add 2 new nodes and remove 2 existing
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
the corresponding data segment.  A log message is appended to all the nodes
so that membership change is persisted.  If the Leader itself is to be removed,
the Leader will append the log to all nodes, waiting for the positive response,
and then stop itself from participating in the consensus.

The raft consensus will repeat the above steps, until the entire raft consensus
are running on the desired nodes for the data segment.


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
is only processed after 3 confirmations of the cluster global consensus protocol.

The Raft consensus identities change only when node membership changes (config change).
Node healthiness (status change) does not change the raft consensus group identities.

### Raft Consensus Membership ###

Node healthiness change, when consistently detected in raft consensus group by
the leader after one full epoch, will be logged to the Raft consensus log, then
published to the cluster level consensus as __poodle.node.healthiness__ domain.

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

Record a domain, key, value tuple encoded as the following.

### Record Magic ###

The first byte is a __magic__.

                  signature bit
    key    domain | 
    | |     | |   |
    7 6 5 4 3 2 1 0
        | |     |
       value    |
                clear bit

- Bit 7 and 6 are the key bits for encoding of a key
  - 00 means no key
  - 01 means 1 byte to represent key length (up to 255 bytes)
  - 10 means 2 bytes to represent key length (up to 65565 bytes)
  - 11 means key is encoded with __data encoding__
- Bit 5 and 4 are the value bits for encoding of a value 
  - 00 means no value
  - 01 means 1 byte to represent value length (up to 255 bytes)
  - 10 means 2 bytes to represent value length (up to 65565 bytes)
  - 11 means value is encoded with __data encoding__
- Bit 3 and 2 are the value bits for encoding of a value 
  - 00 means no domain
  - 01 means 1 byte to represent domain length (up to 255 bytes)
  - 10 means 2 bytes to represent domain length (up to 65565 bytes)
  - 11 means domain is encoded with __data encoding__
- Bit 1 is the clear bit
  - 1 means 'CLEAR' operation
  - 0 means 'UPDATE' operation
- Bit 0 is the signature bit
  - 1 means there is a timestamp and signature at the end of the record
  - 0 means no timestamp or signature at the end of the record
  - If signature bit is 1, the content of data are in raw format,
    and cannot be encoded with lookup scheme, or compression scheme
  
### Record Encoding ###

A full record is encoded as following:

                                     Domain Length
                     Value Length     | |
      Key Length      | |             | |             8 bytes timestamp
      | |             | |             | |             |     |
    X X X X ... ... X X X X ... ... X X X X ... ... X X ... X X ... ... X
    |     |         |     |         |     |         |         |         |
    |     Key Content     |         |     |         |         32 bytes signature
    |                    Value Content    |         |
    |                                    Domain Content
    Magic
                                      
- Lead by a __magic__ byte
- Followed by key length, then key content (if applicable)
- Followed by value length, then value content (if applicable)
- Followed by domain length, then domain content (if applicable)
- Followed by timestamp (8 bytes) and signature (32 bytes) (if applicable)

### Data Magic ###

Data encoding can significantly reduce size of the data by representing
data as a lookup, or in compressed format.

When the record encoding bits are __11__ for key, value, or domain, this
indicates the data follows __data encoding__. In this case, the first byte
of the data encoding is another magic that represents __data encoding
magic__.

         lookup
          | |
    array | |
     bit  | |    length
      |   | |     | |
      7 6 5 4 3 2 1 0
        |     | |
        |     | |
    composite | |
       bit    | |
             compression

- Bit 7 is array bit
  - 1 means the content is an array 
  - 0 means content is not array
  - when this bit is 1, length value means # of elements in the array
- Bit 6 is composite bit
  - 1 means content is a composite (key/value pairs)
  - 0 means content is not composite
  - when this bit is 1, length value means # of elements in the composite
  - bit 7 and bit 6 cannot be 1 at the same time. When both bit 7 and bit 6
    are 1, this has no defined behavior
- Bit 5 and 4 are encoding for lookup scheme
  - 00 means no lookup scheme
  - 01 means 1 byte lookup scheme
  - 10 means 2 byte lookup scheme
  - 11 is reserved
- Bit 3 and 2 are encoding for compression scheme
  - 00 means no compression scheme
  - 01 means 1 byte compression scheme
  - 10 means 2 bytes compression scheme
  - 11 is reserved
- Bit 1 and 0 are encoding for length of data length
  - When lookup bits are not 00, this 2 bits represent data length, not length
    of data length
  - 00 means 0 length
  - 01 means 1 byte length
  - 10 means 2 bytes length
  - 11 is reserved

### Data Encoding ###

A full __data encoding__ is as following:

              Length
      Lookup  | |
      | |     | | Data Content
      | |     | | |         |
    X X X X X X X X ... ... X
    |     | |
    |     Compression
    |
    Magic
    
When data size is relatively small (less than ~1k), and when possible
enumeration of data content is limited, lookup can be an effective
way of reducing the data size.
 
A poodle consensus keeps a list of cluster wide lookup schemes.
The list of schemes are registered across the cluster, and is specific
to a domain.  The cluster wide lookup schema and can be used to encode
data.  E.g.

- A 256 bits ECDSA public key is 32 bytes long.  Sending 32 bytes
  over the wire, or store on disk can represent a significant overhead.
- Instead, if we have a lookup scheme that will lookup the encoded data
  for original content of the data, this can significantly reduce the
  data size to represent an ECDSA public key.
- Assume there are total 10k nodes (10k possible public keys), a perfect
  hash and 2 bytes lookup key will be enough to represent an ECDSA public
  key.
- Considering we will need to continuously evolving lookup schemas (e.g.
  when new nodes are added, and old removed, the lookup scheme will need
  to be updated), we will need to record a list of actively used schemes.
- Assume 1 byte to represent schema, and 2 bytes to represent data content,
  total encoding length of a 32 bytes ECDSA public key is: 1 magic byte +
  1 lookup schema byte + 0 compression scheme byte + 0 length byte + 2
  content bytes = 4 bytes.  This is 87.5% reduction of data size.

When data size is relatively large (larger than ~1k), and when data is not
already compressed, compression can be an effective way to reduce the
data size.

A poodle consensus keeps a list of cluster wide compression schemes.
The list of schemes are registered across the cluster, and can be
used to encode data.


# Packet #

Poodle uses UDP Packet for fast metadata operations.  A UDP packet
consists of a list of Request(s) and Response(s)) followed by the
Packet Timestamp and Signature

### Packet Encoding ###

A Poodle Packet is encoded below:

                    request
                      or
    Node ID         response                 8 bytes timestamp
    |     |         |     |                  |     |
    X ... X X ... X X ... X ... ...  X ... X X ... X X ... ... X
            |     |                  |     |         |         |
            request                  request         32 bytes signature
              or                       or
            response                 response

Poodle Packet is constructed as UDP packet, a Poodle packet must be less
than 64KB.

Poodle Node gathers multiple requests and responses in a buffer during
a very short time period (e.g. 1-20ms), then send the aggregated requests
and responses to the destination Node and Port.  If destination buffer
exceeded a predefined threshold (default 8KB), Poodle will send the content
of the buffer without waiting for the timer.

- Node ID
  - Node ID is encoded with DATA
- A list of requests and responses
  - a list of requests and responses as in the request response encoding
- Timestamp
  - 8 bytes timestamp represent node own timestamp
- Signature
  - 32 bytes signature covers a list of requests and responses and the
    timestamp

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
    - this gets the specified key of specific domain
  - 01 means SET
    - this sets value of the specified key of specific domain. Both
      UPDATE and CLEAR record are considered SET
  - 10 means KEYS
    - this retrieves a list of keys under the specified key of specific
      domain
  - 11 means VALUES
    - this retrieves a list of records (key/value) under the specified
      key of specific domain
- Bits 3 is test bit
  - Test bit enables atomic operation for handling of locked operation
  - When ops is POST (bits 4 and 5 are 01), this will test if a key
    matches the specified value and then UPDATE or CLEAR the key. e.g.
  - Set __value__ to __v1__ for __domain=d1__, __key=k1__ if this value
    is not already set:
    - ops=POST, test=TEST, record=SET, domain=d1, key=k1, value=v1
    - if a value is already set, this operation will return an error
    - if a value is not set, this operation will set the value, and will
      return success
  - Clear __value__ for __domain=d1__, __key=k2__ if value is already
    set to __v2__:
    - ops=POST, test=1, record=SET, domain=d1, key=k1, value=v2
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