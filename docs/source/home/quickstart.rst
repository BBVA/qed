Quick start
===========

This section will guide you through QED functionality.

Mainly, you can **add events** to QED, ask for the proof that an event
**has been inserted**, ask for the proof that two events are **consistent**
between each other, and verify (manual or automatically) each of both proofs.

For each step we will use the **QED CLI** facility.
The client will talk to the QED server and the snapshot store, so it must be
configured for that proposal.

.. important::

    To use the ``qed_client`` command using docker (and forget about installing golang -among other stuff-), do the following:

    .. code::

        $ alias qed_client='docker run -it --net=docker_default bbvalabs/qed:v0.4.10-docs qed client --endpoints http://qed_server_0:8800 --snapshot-store-url http://snapshotstore:8888 --log info'

    Don't hesitate to check the ``qed_client`` help facility when necessary.

    .. code::

        $ qed_client -h
        $ qed_client <command> -h  # Where command=(add, get, membership, incremental)
        ...

.. note::

    In production deployments, the following variables are required, and you
    need to configure it. But, for this quickstart, we will use
    pre-defined values, so *you don't need to configure it for now*.

    .. code-block:: shell

          --api-key             string  Set API Key to talk to QED Log service (default "my-key")
          --endpoints           string  REST QED Log service endpoint list http://ip1:port1,http://ip2:port2...  (default [http://127.0.0.1:8800])
          --snapshot-store-url  string  REST Snapshot store service endpoint http://ip:port  (default "http://127.0.0.1:8888")


1. Environment set up
---------------------

Pre-requisites:

- **docker** (see https://docs.docker.com/v17.12/install/)

- **docker-compose** (see https://docs.docker.com/compose/install/)

Once you have these pre-requisites installed, setting up the quickstart
environment is as easy as:

.. code::

    $ git clone https://github.com/BBVA/qed.git
    $ cd qed/deploy/docker
    $ docker-compose up -d

This simple environment comprises 3 services: **QED Log server**,
**QED Publisher agent**, and **Snapshot store**. 
You should be able to list these 3 services by typing:

.. code-block:: shell

    $ docker ps

Once finished the Quickstart section, don't forget to clean the environment:

.. code::

    $ docker-compose down
    $ unalias qed_client


2. Adding events.
-----------------

In this step the client only interact with the QED server (no snapshot store
info is required). The mandatory field here is the event to insert.

So, let's insert 4 simple events:

.. code::

    $ qed_client add --event "event 0"

    Received snapshot with values:

        EventDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b
        HyperDigest: 6a050f12acfc22989a7681f901a68ace8a9a3672428f8a877f4d21568123a0cb
        HistoryDigest: b8fdd4b2146fe560f94d7a48f8bb3eaf6938f7de6ac6d05bbe033787d8b71846
        Version: 0

    $ qed_client add --event "event 1"
    ...
    $ qed_client add --event "event 2"
    ...
    $ qed_client add --event "event 3"

    Received snapshot with values:

        EventDigest: 6c5cd6775eb412207f7f71f11f09047f1475b2b7526063195b777a230fe4c2a6
        HyperDigest: 7bd6cee5eb0b92801ed4ce58c54a76907221bb4e056165679977b16487e5f015
        HistoryDigest: 4f95cd9fd828abe86b092e506bbffd4662d9431c5755d68eed1ba5e5156fdb13
        Version: 3

.. note::

    This operation should return only if it has been completed successfully or not.
    But currently it returns extra info for debugging/testing purposes.


3. Proof of event insertion.
----------------------------

3.1 Querying proof.
+++++++++++++++++++

To get this proof we only need the original event.
Therefore... has event "event 0" been inserted?

    .. code::

        $ qed_client membership --event "event 0"

        Querying event [ event 0 ] with latest version

        Received membership proof:

            Exists: true
            Hyper audit path: <TRUNCATED>
            History audit path: <TRUNCATED>
            CurrentVersion: 3
            QueryVersion: 3
            ActualVersion: 0
            KeyDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b

Yes! It was inserted in version 0 (ActualVersion), the last event inserted
has version 3 (CurrentVersion), and there is a proof for you to check it.

.. note::

    We print proofs as <TRUNCATED> due to these crypthographical proofs are too long and difficult to read.

3.2 Getting snapshots from the snapshot store.
++++++++++++++++++++++++++++++++++++++++++++++

To vefify the proof, we need to look for the right snapshot
(it contains **"HyperDigest"** and **"HistoryDigest"**, the information needed to verify proofs).

    .. code::

        $ qed_client get --version 3

        Retreived snapshot with values:

            EventDigest: 6c5cd6775eb412207f7f71f11f09047f1475b2b7526063195b777a230fe4c2a6
            HyperDigest: 7bd6cee5eb0b92801ed4ce58c54a76907221bb4e056165679977b16487e5f015
            HistoryDigest: 4f95cd9fd828abe86b092e506bbffd4662d9431c5755d68eed1ba5e5156fdb13
            Version: 3

.. note::

    The snapshot store is the right place to look for digests, instead of using the output of the step 2.


3.3 Verifying proof (manually).
+++++++++++++++++++++++++++++++

Having the proof and the necessary information, let's verify the former.
The interactive process will ask you to enter the info previously retrieved.

    .. code::

        $ qed_client membership --event "event 0" --verify

        Querying event [ event 0 ] with latest version

        Received membership proof:

            Exists: true
            Hyper audit path: <TRUNCATED>
            History audit path: <TRUNCATED>
            CurrentVersion: 3
            QueryVersion: 3
            ActualVersion: 0
            KeyDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b

        Please, provide the hyperDigest for current version [ 3 ]: 7bd6cee5eb0b92801ed4ce58c54a76907221bb4e056165679977b16487e5f015
        Please, provide the historyDigest for version [ 3 ] : 4f95cd9fd828abe86b092e506bbffd4662d9431c5755d68eed1ba5e5156fdb13

        Verifying event with:

            EventDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b
            HyperDigest: 7bd6cee5eb0b92801ed4ce58c54a76907221bb4e056165679977b16487e5f015
            HistoryDigest: 4f95cd9fd828abe86b092e506bbffd4662d9431c5755d68eed1ba5e5156fdb13
            Version: 3

        Verify: OK

And yes! We can verify the membership of "event 0".

3.4 Auto-verifying proofs.
++++++++++++++++++++++++++

This process is similar to the previous one, but we get the snapshots from the
snapshot store in a transparent way.

    .. code::

        $ qed_client membership --event "event 0" --auto-verify

        Querying key [ 0 ] with latest version

        Received membership proof:

            Exists: true
            Hyper audit path: <TRUNCATED>
            History audit path: <TRUNCATED>
            CurrentVersion: 3
            QueryVersion: 3
            ActualVersion: 0
            KeyDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b


        Auto-Verifying event with:

            EventDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b
            Version: 3

        Verify: OK


4. Incremental proof between 2 events.
--------------------------------------

4.1 Querying proof.
+++++++++++++++++++

For this proof we don't need the events, but the QED version in which they
were added (you can get both versions by doing membership proofs as above).

    .. code::

        $ qed_client incremental --start 0 --end 3

        Querying incremental between versions [ 0 ] and [ 3 ]

        Received incremental proof:

            Start version: 0
            End version: 3
            Incremental audit path: <TRUNCATED>

4.2 Getting snapshots from the snapshot store.
++++++++++++++++++++++++++++++++++++++++++++++

This process is similar to the one explained in section 2.2.
As we need 2 snapshots, we repeat the query for each version.

    .. code::

        $ qed_client get --version 0

        Retreived snapshot with values:

            EventDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b
            HyperDigest: 6a050f12acfc22989a7681f901a68ace8a9a3672428f8a877f4d21568123a0cb
            HistoryDigest: b8fdd4b2146fe560f94d7a48f8bb3eaf6938f7de6ac6d05bbe033787d8b71846
            Version: 0

        $ qed_client get --version 3

        Retreived snapshot with values:

            EventDigest: 6c5cd6775eb412207f7f71f11f09047f1475b2b7526063195b777a230fe4c2a6
            HyperDigest: 7bd6cee5eb0b92801ed4ce58c54a76907221bb4e056165679977b16487e5f015
            HistoryDigest: 4f95cd9fd828abe86b092e506bbffd4662d9431c5755d68eed1ba5e5156fdb13
            Version: 3

4.3 Verifying proofs (manually).
++++++++++++++++++++++++++++++++

To verify the proof manually, the process will ask you to enter the required
digests.

        .. code::

            $ qed_client incremental --start 0 --end 3 --verify

            Querying incremental between versions [ 0 ] and [ 3 ]

            Received incremental proof:

                Start version: 0
                End version: 3
                Incremental audit path: <TRUNCATED>

            Please, provide the starting historyDigest for version [ 0 ]: b8fdd4b2146fe560f94d7a48f8bb3eaf6938f7de6ac6d05bbe033787d8b71846
            Please, provide the ending historyDigest for version [ 3 ] : 4f95cd9fd828abe86b092e506bbffd4662d9431c5755d68eed1ba5e5156fdb13

            Verifying with snapshots:
                HistoryDigest for start version [ 0 ]: b8fdd4b2146fe560f94d7a48f8bb3eaf6938f7de6ac6d05bbe033787d8b71846
                HistoryDigest for end version [ 3 ]: 4f95cd9fd828abe86b092e506bbffd4662d9431c5755d68eed1ba5e5156fdb13

            Verify: OK

4.4 Auto-verifying proofs.
++++++++++++++++++++++++++

This process is similar to the previous one, but we get the snapshots from the
snapshot store in a transparent way.

        .. code::

            $ qed_client incremental --start 0 --end 3 --auto-verify

            Querying incremental between versions [ 0 ] and [ 3 ]

            Received incremental proof:

                Start version: 0
                End version: 3
                Incremental audit path: <TRUNCATED>


            Auto-Verifying event with:

                Start: 0
                End: 3

            Verify: OK
