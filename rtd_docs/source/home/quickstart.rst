Quick start
===========

Pre-requisites:

- First start QED server. For more information check our installation section.

This section will guide you through QED functionality.

Mainly, you can **add events** to QED, ask for the proof that an event
**has been inserted**, ask for the proof that two events are **consistent**
between each other, and verify (manual or automatically) each of both proofs.

For each step we will use the **QED CLI** facility.
The client will talk to the QED server and the snapshot store, so it must be
configured for that proposal.

The involved variables are the following ones, and we will use their default
values for this quickstart.

.. code-block:: shell

      --api-key             string  Set API Key to talk to QED Log service (default "my-key")
      --endpoints           string  REST QED Log service endpoint list http://ip1:port1,http://ip2:port2...  (default [http://127.0.0.1:8800])
      --snapshot-store-url  string  REST Snapshot store service endpoint http://ip:port  (default "http://127.0.0.1:8888")

1. Adding events.
-----------------

In this step the client only interact with the QED server (no snapshot store
info is required). The mandatory field here is the event to insert.

So, let's insert 4 simple events:

.. code-block:: shell

    $ go run main.go client add --event "event 0"

    Received snapshot with values:

        EventDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9
        HyperDigest: 90f257b0d905e47f48d769954a0df39affaf6f76a3a7b6880978ae61dbbb8d1e
        HistoryDigest: 163d06ec973f7c902d3ddf6bc10c08c03757004c085e02a3ca463e30ef7aca09
        Version: 0

    $ go run main.go client add --event "event 1"
    ...
    $ go run main.go client add --event "event 2"
    ...
    $ go run main.go client add --event "event 3"

    Received snapshot with values:

        EventDigest: 4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce
        HyperDigest: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d
        HistoryDigest: 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537
        Version: 3

This operation should return only if it has been completed successfully or not.
But currently, it returns certain info for debugging/testing purposes.
In fact, we will retrieve this information later from the right place.

.. note::

    Take a look at the add help section by typing:

    $ go run main.go client add -h


2. Proof of event insertion.
----------------------------

2.1 Querying proof.
+++++++++++++++++++

To get this proof we only need the original event.
So... has "event 0" been inserted?

    .. code-block:: shell

        $ go run main.go client membership --event "event 0"

        Querying event [ event 0 ]

        Received membership proof:

            Exists: true
            Hyper audit path: <TRUNCATED>
            History audit path: <TRUNCATED>
            CurrentVersion: 3
            QueryVersion: 3
            ActualVersion: 0
            KeyDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9

Yes! It was inserted in version 0 (ActualVersion), the last event inserted
has version 3 (CurrentVersion), and there is a proof for you to check it.

.. note::

    We print proofs as <TRUNCATED> due to these crypthographical proofs are too long and difficult to read.

2.2 Getting snapshots from the snapshot store.
++++++++++++++++++++++++++++++++++++++++++++++

This proof shows the version in which the event was inserted.
So, let's ask for the snapshot with that version
(it contains the information needed -"HyperDigest" and "HistoryDigest"- to verify proofs).

    .. code-block:: shell

        $ go run main.go client get --version 3

        Retreived snapshot with values:

            EventDigest: 4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce
            HyperDigest: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d
            HistoryDigest: 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537
            Version: 3

.. note::

    The snapshot store is the right place to look for digests, instead of using the result of the adding step.

    Take a look at the get help section by typing:

    $ go run main.go client get -h


2.3 Verifying proof (manually).
+++++++++++++++++++++++++++++++

Having the proof and the necessary information, let's verify the former.
The interactive process will ask you the info previously retrieved.

    .. code-block:: shell

        $ go run main.go client membership --event "event 0" --verify

        Querying event [ event 0 ]

        Received membership proof:

            Exists: true
            Hyper audit path: <TRUNCATED>
            History audit path: <TRUNCATED>
            CurrentVersion: 3
            QueryVersion: 3
            ActualVersion: 0
            KeyDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9

        Please, provide the hyperDigest for current version [ 3 ]: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d
        Please, provide the historyDigest for version [ 3 ] : 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537

        Verifying event with:

            EventDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9
            HyperDigest: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d
            HistoryDigest: 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537
            Version: 3

        Verify: OK

And yes! We can verify the membership of "event 0".

2.4 Auto-verifying proofs.
++++++++++++++++++++++++++

This process is similar to the previous one, but we get the snapshots from the
snapshot store in a transparent way.

    .. code-block:: shell

        $ go run main.go client membership --event "event 0" --auto-verify

        Querying key [ 0 ] with version [ 3 ]

        Received membership proof:

            Exists: true
            Hyper audit path: <TRUNCATED>
            History audit path: <TRUNCATED>
            CurrentVersion: 3
            QueryVersion: 3
            ActualVersion: 0
            KeyDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9


        Auto-Verifying event with:

            EventDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9
            Version: 3

        Verify: OK


3. Incremental proof between 2 events.
--------------------------------------

3.1 Querying proof.
+++++++++++++++++++

For this proof we don't need the events, but the QED version in which they
were added (you can get both versions by doing membership proofs as above).

    .. code-block:: shell

        $ go run main.go client incremental --start 0 --end 3

        Querying incremental between versions [ 0 ] and [ 3 ]

        Received incremental proof:

            Start version: 0
            End version: 3
            Incremental audit path: <TRUNCATED>

3.2 Getting snapshots from the snapshot store.
++++++++++++++++++++++++++++++++++++++++++++++

This process is similar to the one explained in section 2.2.
As we need 2 snapshots, we repeat the query for each version.

    .. code-block:: shell

        $ go run main.go client get --version 0

        Retreived snapshot with values:

            EventDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9
            HyperDigest: 90f257b0d905e47f48d769954a0df39affaf6f76a3a7b6880978ae61dbbb8d1e
            HistoryDigest: 163d06ec973f7c902d3ddf6bc10c08c03757004c085e02a3ca463e30ef7aca09
            Version: 0

        $ go run main.go client get --version 3

        Retreived snapshot with values:

            EventDigest: 4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce
            HyperDigest: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d
            HistoryDigest: 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537
            Version: 3

3.3 Verifying proofs (manually).
++++++++++++++++++++++++++++++++

To verify the proof manually, the process will ask you to enter the required
digests.

        .. code-block:: shell

            $ go run main.go client incremental --start 0 --end 3 --verify

            Querying incremental between versions [ 0 ] and [ 3 ]

            Received incremental proof:

                Start version: 0
                End version: 3
                Incremental audit path: <TRUNCATED>

            Please, provide the starting historyDigest for version [ 0 ]: 163d06ec973f7c902d3ddf6bc10c08c03757004c085e02a3ca463e30ef7aca09
            Please, provide the ending historyDigest for version [ 3 ] : 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537

            Verifying with snapshots:
                HistoryDigest for start version [ 0 ]: 163d06ec973f7c902d3ddf6bc10c08c03757004c085e02a3ca463e30ef7aca09
                HistoryDigest for end version [ 3 ]: 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537

            Verify: OK

3.4 Auto-verifying proofs.
++++++++++++++++++++++++++

This process is similar to the previous one, but we get the snapshots from the
snapshot store in a transparent way.

        .. code-block:: shell

            $ go run main.go client incremental --start 0 --end 3 --auto-verify

            Querying incremental between versions [ 0 ] and [ 3 ]

            Received incremental proof:

                Start version: 0
                End version: 3
                Incremental audit path: <TRUNCATED>


            Auto-Verifying event with:

                Start: 0
                End: 3

            Verify: OK
