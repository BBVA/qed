Overview
========

What's Q.E.D
------------

``Q.E.D.`` is an open-source software that allows you to establish **trust relations with others**. It can be used in multiple scenarios:

- **Secure tamper-evident data transfers**,
- **Tamper-evident** (system/application/business)
- **Logging**
- ...

Why
---

There are multiple technologies to achieve a similar functionality as QED, such as signed data in a database, or block chain’s related structures.

The **advantages** of the data structure QED implements are:

- Scalability to reach **thousands of millions of events**.
- Proof of **membership** or non-membership generation in **logarithmic time**.
- Proofs of logarithmic size.
- Proof of **temporal consistency** related to QED insertion time.

How
---

``Q.E.D.`` implements a :ref:`forward-secure <forward_secure_glossary>` append-only persistent authenticated data structure. Each append operation produces as a result a cryptographic structure (a signed snapshot), which can verify:

- Whether or not a piece of data is on ``Q.E.D.``.
- Whether or not the appended **data is consistent**, in **insertion order**, to another entry.

How does it works in brief
--------------------------

To verify above statements, we need:

1. The :ref:`snapshot <snapshot_glossary>`,
2. The piece of data inserted and a QED proof.

Then ``Q.E.D.`` emits a :ref:`snapshot <snapshot_glossary>` on every :ref:`event <qed_event_glossary>` insertion, and they need to be accessible elsewhere.

Also ``Q.E.D.`` does not verify **nor emit unsolicited proofs**, it’s the user responsibility to know when and how to verify the data depending on their needs.

.. note::

    Last, ``Q.E.D.`` **does not store the data itself**, only a representation of it produced by a collision-resistant hash function.

    ``Q.E.D.`` **does not provide means to map a piece of data to a QED :ref:`event <qed_event_glossary>`**, so the semantic of the appended data and the relation between each item appended is also a client responsibility.

``Q.E.D.`` process explained in a picture:

.. image:: /_static/images/qed_whiteboard.png

.. note::

    All the scripts and code described live in the branch ‘tampering’ in the `main repository <https://github.com/BBVA/qed>`_.



``Q.E.D.`` is a software to test the scalability of authenticated data structures. Our mission is to design a system which, even when deployed into a non-trusted server, allows one to verify the integrity of a chain of events and detect modifications of single events or parts of its history.

This software is experimental and part of the research being done at BBVA Labs. We will eventually publish our research work, analysis and the experiments for anyone to reproduce.