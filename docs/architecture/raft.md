## Raft
QED uses `raft` for a leader election within a cluster of servers.
Agents do not use raft.

The leader is able to add new events or verify proofs, while
followers only perform the verifiying option, to cope with read
scalability requirements.

Once there is a leader and some followers, QED leader replicate the 
finite state machine (FMI) to the followers before performing the 
insert operation.
Only insert operations (not query operations) are stored in the FMI,
since QED uses them to recover from server failures.
