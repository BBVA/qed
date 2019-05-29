Installation
============

Pre-requisites:

- **docker** (see https://docs.docker.com/v17.12/install/)

- **docker-compose** (see https://docs.docker.com/compose/install/)


Once you have these pre-requisites installed, setting up the quickstart environment is as easy as:

.. code-block:: shell

    $ cd deploy/docker
    $ docker-compose up -d

This simple environment comprises 3 services: **QED server**, **publisher**, and **snapshot store**.
And you should be able to check them by typing:

.. code-block:: shell

    $ docker ps

finding there these 3 services.

Once finished the Quickstart section, don't forget to clean the environment:

.. code-block:: shell

    $ docker-compose down