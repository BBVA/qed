********************************************************************
QED - Scalable, auditable and high-performance tamper-evident log
********************************************************************

.. image:: https://readthedocs.org/projects/qed/badge/?version=latest
   :target: https://qed.readthedocs.io
   :alt: User Documentation Status
.. image:: https://dev.azure.com/bbvalabs/qed/_apis/build/status/BBVA.qed?branchName=master
   :target: https://dev.azure.com/bbvalabs/qed/_build/latest?definitionId=1&branchName=master
   :alt: Build Status
.. image:: https://img.shields.io/azure-devops/coverage/bbvalabs/qed/1.svg
   :target: https://dev.azure.com/bbvalabs/qed/_build/latest?definitionId=1&branchName=master 
   :alt: Azure DevOps coverage
.. image:: https://goreportcard.com/badge/github.com/bbva/qed
   :target: https://goreportcard.com/report/github.com/bbva/qed
   :alt: GoReport
.. image:: https://godoc.org/github.com/bbva/qed?status.svg
   :target: https://godoc.org/github.com/bbva/qed
   :alt: GoDoc



.. figure:: https://raw.githubusercontent.com/BBVA/qed/master/docs/source/_static/images/qed_logo_small.png
   :align: center

**QED** is an open-source software that allows you to establish
**trust relationships** by leveraging verifiable cryptographic proofs.

It can be used in multiple scenarios:

- Data transfers.
- System (or application or business) logging.
- Distributed business transactions.
- Etc.

QED **guarantees** that the system itself, even when deployed
into a **non-trusted server**, cannot be modified without being
detected. It also provides **verifiable cryptographic proofs**
in logarithmic relation (time and size) to the number of entries.

QED is **scalable**, **resilient** and **ops friendly**:

- Designed to manage **billions of events** per shard
- Over **2000 operations per second** per shard under sustained load
- Consistent replication through RAFT
- Operable and instrumented with dozens of metrics
- **Zero config files**, fully documented single binary

Documentation
-------------

You can find the complete documentation at: `Documentation <https://qed.readthedocs.io>`_

Project code
------------

You can find the project code at `Github <https://github.com/BBVA/qed>`_

Authors
-------

QED was made by Hyperscale BBVA-Labs Team.

License
-------

QED is Open Source and available under the `Apache 2 license <https://github.com/BBVA/qed/blob/master/LICENSE>`_.

Contributions
-------------

Contributions are very welcome. See `docs/source/contributing/contributing.rst <https://github.com/BBVA/qed/blob/master/docs/source/contributing/contributing.rst>`_ or skim `existing tickets <https://github.com/BBVA/qed/issues>`_ to see where you could help out.
