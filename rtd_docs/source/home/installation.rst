Installation
============

Pre-requisites:

- **docker** (see https://docs.docker.com/v17.12/install/)

- **docker-compose** (see https://docs.docker.com/compose/install/)


Once you have these pre-requisites installed, setting up the quickstart
environment is as easy as:

.. code-block:: shell

    $ cd deploy/docker
    $ docker-compose up -d

This simple environment comprises 3 services: **QED Log server**,
**QED Publisher agent**, and **Snapshot store**. You should be able
to check them by typing:

.. code-block:: shell

    $ docker ps

Listing there these 3 services.

.. note::

    To enable connectivity from your host to these services, ensure that you have in your /etc/hosts the following line:

    .. code:: shell

        127.0.0.1   localhost   qed_server_0    snapshotstore


Once finished the Quickstart section, don't forget to clean the environment:

.. code-block:: shell

    $ docker-compose down
