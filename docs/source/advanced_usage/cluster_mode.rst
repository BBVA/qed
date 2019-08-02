Cluster mode
==================

This section will guide you through QED cluster features.

Here, you will **check cluster information**, **add events**, and **query proofs**
in a cluster environment (against more than one QED server).

For this functionality we will use the **QED CLI** facility.
The client will talk to the QED servers, so it must be configured for that proposal.

.. important::

    To use ``qed_client`` command using docker (and forget about installing golang -among other stuff-), do the following:

    .. code::

        $ alias qed_client='docker run -it --net=docker_default bbvalabs/qed:v1.0.0-rc1 qed client --log info'

    Don't hesitate to check ``qed_client`` help command when necessary.


1. Environment set up
---------------------

Pre-requisites:

- **docker** (see https://docs.docker.com/v17.12/install/)

- **docker-compose** (see https://docs.docker.com/compose/install/)

Once you have these pre-requisites installed, setting up the required 
environment is as easy as:

.. code::

    $ git clone https://github.com/BBVA/qed.git
    $ cd qed/deploy/docker
    $ docker-compose -f cluster-mode.yml up -d

This environment comprises three **QED Log server** services:
**qed_server_0** will be the cluster leader, while **qed_server_1** and 
**qed_server_2** will be followers.
You should be able to list these service by typing:

.. code-block:: shell

    $ docker ps

Once finished the cluster-mode section, don't forget to clean the environment:

.. code::

    $ docker-compose -f cluster-mode.yml down
    $ unalias qed_client

2. Checking cluster information.
--------------------------------

QED servers have a shard information endpoint that returns how is the cluster formed.
Here we use **curl** to ask for this information, since there is no command for this.

Here we will ask **qed_server_0** (notice that "nodeID: server0"), but you can try another nodes:

.. code::

    $ curl -sS -H "Api-key:my-key" http://localhost:8800/info/shards | python -m json.tool

    {
        "nodeId": "server0",
        "leaderId": "server0",
        "uriScheme": "http",
        "shards": {
            "server0": {
                "nodeId": "server0",
                "httpAddr": "qed_server_0:8800"
            },
            "server1": {
                "nodeId": "server1",
                "httpAddr": "qed_server_1:8800"
            },
            "server2": {
                "nodeId": "server2",
                "httpAddr": "qed_server_2:8800"
            }
        }
    }
 
    $ curl -sS -H "Api-key:my-key" http://localhost:8801/info/shards | python -m json.tool  # For qed_server_1
    $ curl -sS -H "Api-key:my-key" http://localhost:8802/info/shards | python -m json.tool  # For qed_server_2

Servers information is shared between QED servers via Raft. 
Once a server joins the cluster (vía cluster leader), it shares its information and receive others.
Servers interchange information also when a server leaves the cluster, or when leader changes.


3. Adding events.
-----------------

Only QED cluster leader accepts insertions.

**qed_client** is configured by default to discover the cluster topology (using the above information), 
identify which server is the cluster leader, and send requests directly to this server.

.. code::

    $ qed_client --endpoints http://qed_server_0:8800,http://qed_server_1:8800,http://qed_server_2:8800 add --event "event 0"

    Received snapshot with values:

    EventDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b
    HistoryDigest: b8fdd4b2146fe560f94d7a48f8bb3eaf6938f7de6ac6d05bbe033787d8b71846
    HyperDigest: 6a050f12acfc22989a7681f901a68ace8a9a3672428f8a877f4d21568123a0cb
    Version: 0

Notice that given just 1 endpoint is enough to discover the cluster topology, and the cluster leader (qed_server_0).

.. code::

    $ qed_client --endpoints http://qed_server_1:8800 add --event "event 1"

    $ qed_client --endpoints http://qed_server_2:8800 add --event "event 2"


4.  Querying membership proof.
------------------------------

Proofs can be asked to any cluster member.

.. code::

    $ qed_client --endpoints http://qed_server_0:8800 membership --event "event 0" 

    Querying key [ event 0 ] with latest version

    Received membership proof:

    Exists: true
    Hyper audit path: <TRUNCATED>
    History audit path: <TRUNCATED>
    CurrentVersion: 2
    QueryVersion: 2
    ActualVersion: 2
    KeyDigest: 5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b

    $ qed_client --endpoints http://qed_server_1:8800 membership --event "event 0"
    $ qed_client --endpoints http://qed_server_2:8800 membership --event "event 0"

5. Shutting down a server
-------------------------

Here we will stop the QED cluster leader to force a leader election.

.. code::

    $ docker stop qed_server_0 


6. Repeat steps 2-4 several times.
----------------------------------

Check that shard information has been modified, and remember that **qed_server_0** will not work.
