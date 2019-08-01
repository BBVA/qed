Backup and Restore
==================

This section will guide you through QED backup and restore functionalities.

Backup
------

Here, you will **create backups**, **list backups**, and **delete backups** by interacting to the QED management API
(not the same API as the one used in the QuickStart section).

For the backup functionality we will use the backup **QED CLI** facility.
The backup client will talk to the QED server, so it must be configured for that proposal.
To **add events**, we will use the same client as in QuickStart.

.. important::

    To use ``qed_backup`` and ``qed_client`` command using docker (and forget about installing golang -among other stuff-), do the following:

    .. code::

        $ alias qed_client='docker run -it --net=docker_default bbvalabs/qed:v0.4.10-docs qed client --endpoints http://qed_server_0:8800 --snapshot-store-url http://snapshotstore:8888 --log info'

        $ alias qed_backup='docker run -it --net=docker_default bbvalabs/qed:v0.4.10-docs qed backup --endpoint http://qed_server_0:8700 --log info'

    Don't hesitate to check both ``qed_backup`` and ``qed_client`` help commands when necessary.

    .. code::

        $ qed_backup -h
        $ qed_backup <command> -h  # Where command=(create, list, delete)
        ...


1. Environment set up
+++++++++++++++++++++

Pre-requisites:

- **docker** (see https://docs.docker.com/v17.12/install/)

- **docker-compose** (see https://docs.docker.com/compose/install/)

Once you have these pre-requisites installed, setting up the required 
environment is as easy as:

.. code::

    $ git clone https://github.com/BBVA/qed.git
    $ cd qed/deploy/docker
    $ docker-compose -f backup-restore.yml up -d

This environment is not similar to the QuickStart's one. It comprises 1 service: **QED Log server**.
To test backup/restore functionality we do not need any other service but this one.
Moreover, now the DB folder of Qed Log server is mapped to a host temporal folder, to be used
later in the restore section.
You should be able to list this service by typing:

.. code-block:: shell

    $ docker ps

Once finished the backup&restore section, don't forget to clean the environment:

.. code::

    $ docker-compose -f backup-restore.yml down
    $ unalias qed_client
    $ unalias qed_backup

2. Adding events.
+++++++++++++++++

Similarly to QuickStart guide, let's insert 2 events:

.. code::

    $ for i in {0..1}; do qed_client add --event "event $i"; done

    Received snapshot with values:

    EventDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b
    HyperDigest: 6a050f12acfc22989a7681f901a68ace8a9a3672428f8a877f4d21568123a0cb
    HistoryDigest: b8fdd4b2146fe560f94d7a48f8bb3eaf6938f7de6ac6d05bbe033787d8b71846
    Version: 0


    Received snapshot with values:

    EventDigest: fb378474af5953bec611fcb2602c5b61271c1f233b60c0adba76d5d6f47a50c4
    HyperDigest: 15814ee2f820da9c126fc740d5b4de034d250a3f5fe6e58ab5616026cb65b3dd
    HistoryDigest: ae6fe0b70e09b12eeea3bc2cb923d239d184a2b30a578e201ad952e2e9a405f2
    Version: 1

3. Creating backups.
++++++++++++++++++++

.. code::

    $ qed_backup create

    Backup created!

The version of the last inserted event is stored into the backup metadata.

4. Listing backups.
+++++++++++++++++++

.. code::

    $ qed_backup list

    Backup list:
    Id: 1	Timestamp: 2019-07-17T13:13:26	Version: 1	Size(GB): 0	Num.Files: 4

5. Repeat steps 2-4 several times.
++++++++++++++++++++++++++++++++++

.. code::

    $ for i in {2..3}; do qed_client add --event "event $i"; done 
    $ qed_backup create
    $ for i in {4..5}; do qed_client add --event "event $i"; done
    $ qed_backup create
    $ qed_backup list

    Backup list:
    Id: 1	Timestamp: 2019-07-17T13:13:26	Version: 1	Size(GB): 0	Num.Files: 4
    Id: 2	Timestamp: 2019-07-17T13:13:40	Version: 3	Size(GB): 0	Num.Files: 4
    Id: 3	Timestamp: 2019-07-17T13:13:54	Version: 5	Size(GB): 0	Num.Files: 4

6. Deleting backups.
++++++++++++++++++++

.. code::

    $ qed_backup delete --backup-id=1

    Backup deleted!

    $ qed_backup list

    Backup list:
    Id: 2	Timestamp: 2019-07-17T13:13:40	Version: 3	Size(GB): 0	Num.Files: 4
    Id: 3	Timestamp: 2019-07-17T13:13:54	Version: 5	Size(GB): 0	Num.Files: 4

Restore
-------

Here, you just will **restore** a QED log server state **from a previous backup**, being able to choose 
the latest backup (by default) or a certain backup ID to recover from (see IDs above).

1. Environment set up.
++++++++++++++++++++++

To simulate a new QED log server, let's destroy the current environment and create a new one
from scratch. To destroy the environment, just do:

.. code::

    $ cd qed/deploy/docker
    $ docker-compose down
    ...

Remember that we saved the backups folder in a host path.
So let's check that the folder has backup information.

.. code::

        $ tree /tmp/backups/

        /tmp/backups/
        ├── meta
        │   ├── 2
        │   └── 3
        ├── private
        │   ├── 2
        │   │   ├── 000003.log
        │   │   ├── CURRENT
        │   │   ├── MANIFEST-000004
        │   │   └── OPTIONS-000014
        │   └── 3
        │       ├── 000003.log
        │       ├── CURRENT
        │       ├── MANIFEST-000004
        │       └── OPTIONS-000014
        └── shared

There are information of backups 2 and 3 as expected (we deleted backup 1 before).

To create a new environment from scratch, just do: 

.. code::

    $ docker-compose -f backup-restore.yml up -d

Finally, let's check that the "event 0" is not present in the new QED log server.

.. code::

    $ qed_client membership --event "event 0"

    Querying event [ event 0 ] with latest version

    Received membership proof:

        Exists: false
        Hyper audit path: <TRUNCATED>
        History audit path: <TRUNCATED>
        CurrentVersion: 18446744073709551615
        QueryVersion: 18446744073709551615
        ActualVersion: 18446744073709551615
        KeyDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b
        
Notice that the event does not exist.

2. Restore process.
+++++++++++++++++++

Get into the QED log server:

.. code::

    $ docker exec -it qed_server_0 /bin/bash

Restore backup 2, from the interal docker backup folder, to the interal docker path where the DB is:

.. code::

    $ qed restore --backup-dir "/var/tmp/qed0/db/backups/" --restore-path "/var/tmp/qed0/db/" --backup-id 2 --log info

Exit the QED server, and restart the container to make QED server aware of the restored DB.

.. code::

    $ exit
    $ docker restart qed_server_0


3. Check event membersip.
+++++++++++++++++++++++++

Event 0 (and up to event 3) should be there:

.. code::

    $ qed_client membership --event "event 0"

    Querying key [ event 0 ] with latest version

    Received membership proof:

    Exists: true
    Hyper audit path: <TRUNCATED>
    History audit path: <TRUNCATED>
    CurrentVersion: 3
    QueryVersion: 3
    ActualVersion: 0
    KeyDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b

But event 4 should not:

.. code::

    $ qed_client membership --event "event 4"

    Querying key [ event 4 ] with latest version

    Received membership proof:

    Exists: false
    Hyper audit path: <TRUNCATED>
    History audit path: <TRUNCATED>
    CurrentVersion: 3
    QueryVersion: 3
    ActualVersion: 3
    KeyDigest: 2d245d477b973c0895afc098b46762967f728e5aec8555d81ceaf1996d4c33e0

.. important::

    Try restoring other backups and checking the membership of other events.

    (repeat step 2 and 3 with different values)
