# Message passing.

QED receives events as input, and outputs signed snapshots. These snapshots
are inputs to any agent (publisher, monitor and auditor), so QED needs to
pass batchs of signed snapshots to the agents.

## Gossip

QED server and agents use the [memberslist](https://github.com/hashicorp/memberlist)
package from HashiCorp to create lists of servers, publishers, monitors, and
auditors.

Then, QED sends a batch of signed snapshots to a configurable number `N` of
each agent type vía memberlist `send reliable` tcp connection, adding a TTL
to each batch.

Agents receive a batch of signed snapshots, perform their particular task
using it, and send the batch again to other agents (not to QED), reducing the
message TTL. Message passing ends when TTL is equal to 0.

## Alternatives

Besides of Gossip, HTTP protocol can be used for passing messages, but
syncronous requests make QED to not perform as expected.