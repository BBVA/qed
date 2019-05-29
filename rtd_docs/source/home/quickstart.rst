Quick start
===========


1. Adding events
----------------

.. code-block:: shell

    $ go run main.go client add --event "event 0" --api-key my-key

    Received snapshot with values:

        EventDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9
        HyperDigest: 90f257b0d905e47f48d769954a0df39affaf6f76a3a7b6880978ae61dbbb8d1e
        HistoryDigest: 163d06ec973f7c902d3ddf6bc10c08c03757004c085e02a3ca463e30ef7aca09
        Version: 0

    $ go run main.go client add --event "event 1" --api-key my-key
    ...
    $ go run main.go client add --event "event 2" --api-key my-key
    ...
    $ go run main.go client add --event "event 3" --api-key my-key

    Received snapshot with values:

        EventDigest: 4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce
        HyperDigest: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d
        HistoryDigest: 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537
        Version: 3

2. Proof of event insertion.
----------------------------

    2.1 Querying proof.

    Has "event 0" been inserted?

    .. code-block:: shell

        $ go run main.go client membership --event "event 0" --api-key my-key

        Querying event [ event 0 ]

        Received membership proof:

            **Exists: true**
            **Hyper audit path: <TRUNCATED>**
            **History audit path: <TRUNCATED>**
            CurrentVersion: 3
            QueryVersion: 3
            **ActualVersion: 0**
            KeyDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9

    Yes! It was inserted in version 0, and the last event inserted has been in version 3.

    And there is a proof for you to check it.

    2.2 Getting snapshots from the snapshot store.

    Ok, let's ask for the information I need ("hyperDigest" and "historyDigest") to check it.

    .. code-block:: shell

        $ go run main.go client get --version 3 --snapshot-store-url http://127.0.0.1:8888

        Retreived snapshot with values:

            EventDigest: 4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce
            **HyperDigest: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d**
            **HistoryDigest: 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537**
            Version: 3

    2.3 Verifying proof (manually).

    This interactive process will ask you the info previously retrived.

    .. code-block:: shell

        $ go run main.go client membership --event "event 0" --api-key my-key **--verify**

        Querying event [ event 0 ]

        Received membership proof:

            Exists: true
            Hyper audit path: <TRUNCATED>
            History audit path: <TRUNCATED>
            CurrentVersion: 3
            QueryVersion: 3
            ActualVersion: 0
            KeyDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9

        Please, **provide the hyperDigest** for current version [ 3 ]: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d
        Please, **provide the historyDigest** for version [ 3 ] : 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537

        **Verifying** event with:

            EventDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9
            HyperDigest: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d
            HistoryDigest: 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537
            Version: 3

        **Verify: OK**

    2.4 Auto-verifying proofs.

    This process is similar to the previous one, but getting snapshots from the snapshot store in a transparent way.

    .. code-block:: shell

        $ go run main.go client membership --event 0 --version 3 --api-key my-key **--auto-verify**

        Querying key [ 0 ] with version [ 3 ]

        Received membership proof:

            Exists: true
            Hyper audit path: <TRUNCATED>
            History audit path: <TRUNCATED>
            CurrentVersion: 3
            QueryVersion: 3
            ActualVersion: 0
            KeyDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9


        **Auto-Verifying** event with:

            EventDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9
            Version: 3

        **Verify: OK**


3. Incremental proof between 2 events.
--------------------------------------

    3.1 Querying proof.

    .. code-block:: shell

        $ go run main.go client incremental --start 0 --end 3 --api-key my-key

        Querying incremental between versions [ 0 ] and [ 3 ]

        Received incremental proof:

            Start version: 0
            End version: 3
            Incremental audit path: <TRUNCATED>

    3.2 Getting snapshots from the snapshot store.

    .. code-block:: shell

        $ go run main.go client get --version 0 --snapshot-store-url http://127.0.0.1:8888

        Retreived snapshot with values:

            EventDigest: 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9
            HyperDigest: 90f257b0d905e47f48d769954a0df39affaf6f76a3a7b6880978ae61dbbb8d1e
            HistoryDigest: 163d06ec973f7c902d3ddf6bc10c08c03757004c085e02a3ca463e30ef7aca09
            Version: 0

        $ go run main.go client get --version 3 --snapshot-store-url http://127.0.0.1:8888

        Retreived snapshot with values:

            EventDigest: 4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce
            HyperDigest: 28b2a8d7bfeedc61b988e5bddaf260f21aee96bfe88392a0af8a06d7129ab86d
            HistoryDigest: 9c577745b6979e1243b707d43f4ca3aa45859d5277bc37f63f4489322f1bf537
            Version: 3

    3.3 Verifying proofs.

        .. code-block:: shell

            $ go run main.go client incremental --start 0 --end 3 --api-key my-key--verify

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

        .. code-block:: shell

            $ go run main.go client incremental --start 0 --end 3 --api-key my-key--auto-verify

            Querying incremental between versions [ 0 ] and [ 3 ]

            Received incremental proof:

                Start version: 0
                End version: 3
                Incremental audit path: <TRUNCATED>


            Auto-Verifying event with:

                Start: 0
                End: 3

            Verify: OK
