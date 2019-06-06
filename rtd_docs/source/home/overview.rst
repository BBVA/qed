Overview
========

What's QED
------------

**QED** is an open-source software that allows you to establish
**trust relationships** by leveraging verifiable cryptographic proofs.

In real-life, there are countless scenarios where maintaining a chronological
record of events and operations is considered as a general principle for
good internal business controls. We usually refer to this record as an
audit trail, and it provides proof of compliance and operational integrity.

A paradigmatic example, with centuries of history, are the ledgers used by
accountants to register all financial and non-financial data of an
organization. But, the potential uses cases are not limited to that kind of
information, and can be extended to any sensitive activity that could happen
inside an organization, or exchanged between peers. For instance:

- Data transfers.
- System (or application or business) logging.
- Distributed business transactions.
- Etc.

Audit trails transitioned from manual to electronic records,
that make this historical information more accurate, easily accessible, and
usable. This also made easier the task of **auditing**, which is essential for
maintaining some grade of confidence with the integrity of the stored data.

But here is where a problem of **trust** appears: how can we assure that
nobody, either an insider or an outsider, tampered with such data?

QED comes to solve this lack of trust by adding **transparency** to the way
that different parties interact with some specific set of data. It provides
transparency by **making evident** any further non-authorized change either on
such data or on the data that QED stores itself, even when deployed into a
non-trusted server. And the way it achieves this capability is by using such a
extended technology as **verifiable cryptographic proofs**.

Why
---

In practice, there are multiple ways to achieve a similar functionality as
QED implements that range from very simple technologies, as might be the
case of storing signed data (by certificate signature) into a database, to
far more complicated approaches like blockchain-based technologies and smart
contracts. But QED has important **advantages**  over such alternatives:

- Works completely **detached** from the event source (database, logging
  system,...),
  and so from the usual way to interact with such data.
- Scales to reach **billions of events**.
- Generates proofs of **membership** or non-membership in **logarithmic time**.
  and with **logarithmic size**.
- Generates proofs of **temporal consistency** related to QED insertion
  time.

How
---

QED implements a forward-secure append-only persistent authenticated data
structure. Each append operation produces as a result a cryptographic
structure (a signed snapshot), which acts as a receipt for the operation,
and can be used later to verify the following statements:

- Whether or not a piece of data is on QED.
- Whether or not the appended data **is consistent**, in **insertion order**,
  to another entry.

QED can be requested to proof whether the above statements are true or
false for a specific piece of data. In response to that requests, QED
returns a cryptographic proof which, combined with the original piece of data,
can generate again the cryptographic value of the original snapshot.

Please refer to our :ref:`trust model <trust_model>` section to better
understand this point.


.. note::

    QED **does not store the data itself**, only a representation of it
    produced by a collision-resistant hash function.

    QED **does not provide means to map a piece of data to a QED event**,
    so the semantic of the appended data and the relation between each item
    appended is also a client responsibility.


.. note::

    This software is experimental and part of the research being done at
    BBVA Labs. We will eventually publish our research work, analysis and
    the experiments for anyone to reproduce.
