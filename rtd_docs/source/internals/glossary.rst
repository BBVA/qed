Glossary
========

The purpose of this section is to equip the reader with necessary background
about the most common keywords and concepts used in the development of
verifiable (or authenticated) data structures.

Cryptographic primitives
------------------------

Cryptographic hash functions and digital signatures are the fundamental
building blocks for creating authenticated data structures.

Hash functions
++++++++++++++

A cryptographic hash function compresses an arbitrary large message *m* into a
fixed size digest *h*. Due to the large space of messages mapped, collisions
are inevitable but they must be computationally hard to find. A cryptographic
hash function must conform with the following properties:

- **Preimage resistance:** given a digest `*h* <- H(*m*)` for message *m*, it
  must be computationally hard to find a preimage *m'* generating *h* without
  knowledge of *m*.

- **Second preimage resistance:** given a fixed preimage *m*, it must be
  computationally hard to find another preimage *m' != m* such that
  `H(*m*) = H(*m'*)`.

- **Collision resistance:** it must be computationally hard to find any
  distinct preimages *m1* and *m2* such that `H(*m1*) = H(*m2*)`.

Digital signatures
------------------

A digital signature is a mathematical scheme for demonstrating the
*authenticity*, *non-repudiation* and *integrity* of a message. So a valid
digital signature gives a recipient a reason to believe that the message was
created by a known sender, that the sender cannot deny having sent the message
and that the message was not altered in transit.

Tree-based data structures
--------------------------

A tree is an (un)ordered collection of entities, not necessarily unique, that
has a hierarchical parent-child relationship between pairs of entities. Every
tree has a single *root* node designating the start of the tree, and each
*descendant* is recursively defined as a tree. A node is said to be an
*ancestor* to all its descendants, and a parent to its concrete *children*.
All children that have the same parent are referred to as *siblings*, and
every node without children is referred to as a *leaf*. The root is said to be
found at *level* one, the *height *is the number of levels in the tree, and
the *depth* of a subtree rooted at a leaf is zero.

Binary tree
+++++++++++

A binary tree is a tree where each node is restricted to at most a *left
child* and a *right child*.

Perfect binary tree
+++++++++++++++++++

A binary tree of height *h* that must contain exactly 2^h - 1 nodes.

Full binary tree
++++++++++++++++

A binary tree which all nodes must have two or no children.

Complete binary tree
++++++++++++++++++++

A binary tree which must be filled left-to-right at the lowest level, and
entirely at the level above.

Merkle tree
-----------

A binary tree that stores values at the lowest level of the tree and uses
cryptographic hash functions. While leaves compute the hash of their own
attributes, parents derive the hash of their children’s hashes concatenated
left-to-right. Therefore the hash rooted at a particular subtree is
recursively dependent on all its descendants, effectively serving as a
succinct summary for that subtree.

Membership proof
++++++++++++++++

A Merkle tree can prove values to be present by constructing efficient
*membership proofs*. Each proof must include a *Merkle audit path*, and it is
verified by recomputing all hashes, bottom up, from the leaf that the proof
concerns towards the root. The proof is believed to be valid if the recomputed
root hash matches that of the original Merkle tree, but to be convincing it
requires a trustworthy root (e.g., signed by a trusted party or published
periodically in a newspaper).

Merkle audit path
+++++++++++++++++

A Merkle audit path for a leaf is the list of all additional nodes in the
Merkle tree required to compute the Merkle Tree Hash for that tree. If the
root computed from the audit path matches the true root, then the audit path
is proof that the leaf exists in the tree.

History tree
------------

An append-only Merkle tree that stores *events* left-to-right at the lowest
level of the tree. It is not lexicographically sorted, and unable to generate
efficient non-membership proofs, but it is *naturally persistent*, supports
efficient membership proofs and allows to generate *incremental proofs*.

Persistent nature
+++++++++++++++++

A history tree is naturally persistent, in the sense that past versions of the
tree can be efficiently reconstructed and queried for membership.

Incremental proof
+++++++++++++++++

A history tree can show consistency between root hashes for different views,
and that requires proving all events in the earlier view present in the newer
view. It is achieved by returning just enough information to reconstruct both
root hashes checking if expected roots are obtained.

Binary search tree
------------------

A binary tree that requires the value of each node to be greater (or lesser)
that the value of its left (or right) child. This property, referred to as the
*BST property*, implies a lexicographical order and allows every look-up
operation to use a divide-and-conquer technique known as *binary search*.
Because the time required to complete a binary search is bounded by the height
of the BST, it is important that the tree structure remains *balanced*.

Heap
----

A specialized tree-based data structure used in the context of priority
queues. It associates each node a priority and preserves, at all times, two
properties: the *shape property*, requiring that the heap is a complete binary
tree; and the *heap property*, requiring that every node has a lower or equal
priority with respect to its parent.

Treap
-----

A randomized search tree associating with each entity a *key* and a randomly
selected priority. Treaps enforce the BST property with respect to keys, the
heap property with respect to priorities, and are also *set-unique*.
Set-uniqueness ensures the tree structures of identical collections to be
equivalent, thereby implying *history independence* if priorities are assigned
deterministically.


Hash treap
----------

A lexicographically sorted history independent key-value store combining a
regular Merkle tree and a deterministic treap. Each node is associated with an
entity and every (non-)member has a unique position, therefore hash treaps
support efficient (non-)membership proofs.

Sparse Merkle tree
------------------

A Merkle tree which depth is fixed in advance with respect to the underlying
hash function H, meaning there are always 2^|H(.)| leaves.  These are referred
left-to-right by indices, and are associated with either *default* or
*non-default* values. In the latter case the hash of a key determines the
index, which implies there is a unique leaf reserved for every conceivable
digest H(*k*). This allows generation of (non-)membership proofs using regular
Merkle audit paths. The SMT is *sparse* because the large majority of all
leaves will be empty, and consequently most nodes rooted at lower levels of
the tree derive identical default hashes.
