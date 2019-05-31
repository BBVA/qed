Frequently Asked Questions
==========================

1. Why would anyone want to verify other's activities?
++++++++++++++++++++++++++++++++++++++++++++++++++++++

To ensure that significant information has not been modified without
being noticed. For example:

* As a user of a specific service, I want to ensure the provider
  of such service does not change the data I produce or use.
  Imagine a social network that rewrites some messages they don't
  like. Or a compromised software download page that make users
  download malware to their computers.
* As a service provider, I want to ensure my user's orders and
  service agreements cannot be repudiated or modified after signed.
* As a service provider participating in a network of services, which
  operates on behalf of their customers, I don't want someone to issue
  orders on behalf of my customers without being noticed.

There could be hundreds of situations on which you can leverage QED's
functionality to achieve tamper-evident security.

2. What is considered a user in QED?
++++++++++++++++++++++++++++++++++++

A QED user is anyone who can access to the QED Log, the Snapshot Store
and the QED events.

The user can be affiliated with the same organization where the QED
system is deployed or with another one, or even unaffiliated, depending
on the trust model you build with QED.

3. Can QED ensure an event is legit?
++++++++++++++++++++++++++++++++++++

No. QED is unable to guarantee the veracity of an inserted event, it can
only verify if an inserted event has not been modified and if its
insertion order has not been altered.

This means that if you inserted fake events into QED, you can only be
able to verify fake events.

4. But how can QED help me to achieve that guarantees?
++++++++++++++++++++++++++++++++++++++++++++++++++++++

QED provides cryptographic proofs that demonstrate:

* Whether or not a piece of data was inserted in QED.
* Whether or not the appended data is consistent, in its insertion
  order to another entry.

4.1 And with only those two proofs, can we achieve all those functionalities?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

No, QED is a software that helps you to implement processes and other pieces
of software which will enable you to build those capabilities.

5. How is QED different from digital signatures or blockchain?
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

In the following table we compare the main characteristics of each technology:

+---------------------+--------------------+------------+------------------+
| Feature             | Digital signature  | Blockchain | QED              |
|                     | + DB               |            |                  |
+=====================+====================+============+==================+
| Prove inclusion     | Logarithmic        | Linear     | Logarithmic      |
| (time)              |                    |            |                  |
+---------------------+--------------------+------------+------------------+
| Prove non-inclusion | Linear             | Linear     | Logarithmic      |
| (time)              |                    |            |                  |
+---------------------+--------------------+------------+------------------+
| Prove append-only   | Linear             | Linear     | Logarithmic (non |
| (time)              |                    |            | deletion proof   |
|                     |                    |            | can take linear  |
|                     |                    |            | time)            |
+---------------------+--------------------+------------+------------------+
| Consistency proof   | Linear             | Linear     |  Logarithmic     |
| size                |                    |            |                  |
+---------------------+--------------------+------------+------------------+
| Proof size          | Constant           | Constant   |  Logarithmic     |
+---------------------+--------------------+------------+------------------+
| Tampering detection | No, if the PK gets | Linear     |  Logarithmic     |
| (time)              | compromised        |            |                  |
+---------------------+--------------------+------------+------------------+

Either entries digitally signed in a database or a blockchain network can
provide for proofs like QED, and similar functionality can be achieved by all
of them.

QED really shines when it is used to build a lot of inclusion proofs or
consistency proofs, because its performance allows you to save a lot of space
and computing power, which can be transformed into scalability.

QED is designed to handle **billions of entries** at over 2000 operations per
second.

6. Is it secure?
++++++++++++++++

The security model of QED is based on three pillars:

* Strong cryptographic hash functions (SHA256 or BLAKE2).
* Separated source data stores from the proof store and the snapshot store.
* Active and decentralized monitoring.

It is fundamental for QED to use a fast, reliable and unbroken hash function.
This allows you to avoid collisions and ensure event information cannot leak.

Also, in order to verify any of the QED issued proofs a user will need three
items:

* The original event inserted into QED.
* The proof issued by QED.
* The authentication token (snapshot) published by QED when the event was
  inserted.

And lastly, QED includes active monitoring that throw alerts if something
goes wrong.

But be aware, we intrinsically trust the append operation to QED. If you insert
fake data, you verify fake data. There is no way to fully prevent that, in this
system or in any other.

7. Will QED alert me from changes or tampering attempts?
++++++++++++++++++++++++++++++++++++++++++++++++++++++++

No. QED will never issue proofs proactively nor be aware of tampering. It
is the user responsibility to actively monitor QED to detect those
modification attempts.

Besides, QED will also help you detect tampering in itself.

8. Is QED a data store? Can I save my data into QED to secure it?
+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

No. QED does not store any data, it only stores a fingerprint of
such data using a strong hashing function. It only supports three operations:

* Append a new entry.
* Ask for an inclusion proof.
* Ask for a consistency proof.
