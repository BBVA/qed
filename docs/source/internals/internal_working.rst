How does it works (long version)
================================

TODO

.. Tampering
.. ---------

.. Event source
.. ++++++++++++

.. * QED can emit a verifiable proof to check if an event is on QED.
.. * QED can emit a verifiable proof to check if two events are consistent with each other in the order of insertion.
.. * The user has the responsibility to ask for these proofs and verify them and can user the QED gossip network to build auditors and monitors adapted to its use case.
.. * The user should use a secret unknown to QED for the QED event mapping function.
.. * QED does not audit nor emits proof or verify proactively any event.
.. * QED does not alert in real time about event source changes.

.. Application
.. +++++++++++

.. - We cannot guarantee an application will use QED.
.. - We can use QED capabilities to build external tooling to check the application expected behavior.

.. Third party
.. +++++++++++

.. * We can use QED to verify changes in third-party data source using a QED client which must implement a mapping function between the third-party data to QED events.
.. * We can use QED to check the history of changes of a third party ordered data source. Also, the source of the order could be build from another means.

.. QED log
.. +++++++

.. QED is resistant to naive attempts to tamper with its database. A modification of a single leaf of a tree, or path is detected easily. This kind of tampering tries to harm the credibility of the system by making it complain or to avoid the validation of a correct event. *Once the QED is tampered with, there is no rollback. Only a complete rebuild can ensure its coherence.*

.. We can alter the events stored in QED in a way that the proofs will verify only if the QED version is reset to an old version and we insert events from that version again using the QED append algorithm to regenerate all the intermediate nodes of the trees:

.. TODO

.. v0————>v1————>v2————>v3 ————>v4 ————> v5              original history
..                      |                                version reset
..                      |—>v3’————>v4’————>v5’————>v6    forked history


.. This a theoretical attack, in practice it is unfeasible to do such an attack without being detected, as it requires modifying a running program which replicates on real time, without being noticed.
.. Also, even if the attack happens, it can be detected doing a full audit checking all events against the event source and the snapshot store.

.. *QED will not know which component was tampered, only that an event being check has either its event source, its snapshot, or its QED event altered. We will not establish the source of truth unless we do a full audit which comprises the insertion of all QED events again to regenerate the source, the log and the snapshots to check the differences.*

.. To further protect a QED deployment against such tampering, we recommend salting the QED events with a secret (which QED does not know) verifiable by the event stakeholders and recommends implementing a monitoring agent that check the snapshot store searching for duplicate QED versions.

.. Another recommendation is to make QED clusters to block any arbitrary non-authenticated joins, replications or from-disk recoveries.

.. Last, the teams or companies in charge of the QED log, agents and snapshot store should be different to avoid collusion.

.. QED agents
.. ++++++++++

.. The agent's mission is to check the QED activities to identify anomalous behaviors and also publish the snapshots into the store.

.. They can be tampered as any other application, making them publish altered snapshots or to omit any functionality.
.. But the QED proofs verification process will detect modifications regarding the events being checked as long as the event source or QED log are untampered.

.. The gossip agents can use certificates and/or user/password to authenticate against each other.

.. QED snapshot store
.. ++++++++++++++++++

.. The snapshot store can be compromised to change stored snapshots, but like in the QED agents case, the QED proofs verification process will fail as long as the event source or QED log are untampered.
