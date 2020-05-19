
### Node request to join cluster ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.cluster.node.request |
| Key | Node Public Key |
| Value | Cluster Public Key |
| Operation | SET |
| Timestamp | 8 bytes |
| Signature | Signed with Node Private Key |

### Cluster approval of Node join request ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.cluster.membership |
| Key | Node Public Key |
| Value | Original Request RECORD |
| Operation | SET - approved; CLEAR - rejected |
| Timestamp | 8 bytes |
| Signature | Signed with Cluster Private Key |
  
### Cluster request to remove Node ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.cluster.membership |
| Key | Node Public Key |
| Value | nil |
| Operation | CLEAR |
| Timestamp | 8 bytes |
| Signature | Signed with Cluster Private Key |

### Cluster config change ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.cluster.conf |
| Key | poodle cluster conf key |
| Value | poodle cluster conf value |
| Operation | SET; CLEAR |
| Timestamp | 8 bytes |
| Signature | Signed with Cluster Private Key |

### Node config change ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.cluster.node |
| Key | Node Public Key |
| Value | Node Configs: IP Addr(s), Port, etc. encoded as DATA |
| Operation | SET; CLEAR |
| Timestamp | 8 bytes |
| Signature | Signed with Node Private Key |

### Node status change ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.status.node |
| Key | Node Public Key |
| Value | Node Status: Up/Down/Flappy, etc. encoded as DATA |
| Operation | SET; CLEAR |
| Timestamp | 8 bytes |
| Signature | Signed with Node Private Key |

### Propose Cluster Consensus Block ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.consensus.cluster.proposal |
| Key | Block Hash, Block Height, encoded as DATA |
| Value | Block Content + Block Signature of Proposing Node. encoded as DATA |
| Operation | SET |
| Timestamp | 8 bytes |
| Signature | Signed with Node Private Key |

### Sign Cluster Consensus Block ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.consensus.cluster.sign |
| Key | Block Hash, Block Height, encoded as DATA |
| Value | Block Signature by Signing Node, encoded as DATA |
| Operation | SET |
| Timestamp | 8 bytes |
| Signature | Signed with Node Private Key |

### Propose Status Consensus Block ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.consensus.status.proposal |
| Key | Block Hash, Block Height, encoded as DATA |
| Value | Block Content + Block Signature of Proposing Node. encoded as DATA |
| Operation | SET |
| Timestamp | 8 bytes |
| Signature | Signed with Node Private Key |

### Sign Status Consensus Block ###

| Field | Value |
| :--- | :--- |
| Domain | poodle.consensus.status.sign |
| Key | Block Hash, Block Height, encoded as DATA |
| Value | Block Signature by Signing Node, encoded as DATA |
| Operation | SET |
| Timestamp | 8 bytes |
| Signature | Signed with Node Private Key |


