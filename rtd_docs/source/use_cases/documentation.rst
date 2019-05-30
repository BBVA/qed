Certification of Documents, Emails, Contracts, etc.
===================================================

How can we create transparency about a transaction that somebody as interest to
certify that a ``DOCUMENT`` is emitted that way?

Allowing QED to store a fingerprint of the transaction ``F(DOCUMENT)``, will
prove that it was as intended.

Furthermore the proof returned by the QED server it is a cryptographic proof
``WARRANT`` with legal validity.

QED is a **tamper evident** storage, that is that QED it can be deployed in
untrusted servers, because of the way QED stores the transactions.

In this Use case we will try to explain in mundane terms why QED is worth the
effort to be used as warranteer.

.. image:: /_static/images/Uc2.png


Trust the untrustable
---------------------

Using today's technologies are by far trusted by default. A myriad of problems
can emerge in the neccessity to protect sensitive data, and even the when the
maximum level of isolation and protection are in place, you can always ask
**Who whatches the Watchmen?**

This trust dilemma is called in *distribution systems* a `Bizantine Fault
Tolerant`_ service.

.. _`Bizantine Fault Tolerant`: https://en.wikipedia.org/wiki/Byzantine_fault

Cryptographic proofs
--------------------

QED address this problem by using a internal tree storage that are
*statistically impossible* to alter without detection.

This is mainly because some inherent properties of the cryptographic
algorithms we use. From the original event source is fast and coherent to
create a *cryptographic hash* but **statistically impossible** to find other
input that could create the same output.

The other interesting property of the cryptographic hashers are the
**sparsity** of the hashes. This mean that similar inputs provides completely
different results, and the *distance* between those results are really wide.

This both properties are *abused* in QED in order to create a tamper evident
storage, even on untrustable environments.


Understanding the QED storage
+++++++++++++++++++++++++++++

QED stores all the transactions in a append-only tree. this allows us to track
the previous and future transactions that where sent to the QED server.

In order to prevent tamperings, we use a `Merkle tree`_. Which is a
cryptographic sum between adjacent elements in a tree fashion. This allows us
to make a lot of cryptographic hashes, between the last inserted elements and
all the previous ones.

.. image:: /_static/images/Hash_Tree.svg

Since the append-only storage can grow really fast, we need a way to find
previously inserted transactions, so we use another cryptographic tree, to
prevent tampering in finding the stored transactions.

How a Proof can be used
+++++++++++++++++++++++

Once a transaction is stored, we publish the final sum of all the cryptographic
nodes in a public, distributed storage.

If the need to prove that some transaction exists we return and audit path of
the current QED storage. Any alteration of the history will be evident and we
can't determine if the transaction is the same as it was included in the QED
server in first place. Only if the history is coherent the proof will be verified.

A final note along the auditable proofs is that it must be verified outside the
QED server in order to allow transparency.
