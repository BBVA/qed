.. _trust_model:

Trust Model
===========

Before starting to use ``QED``, users need to translate their problem of trust
to a more suitable conceptual model to allow them to accurately identify which
are the actors that take part in the relationship and what are the pieces of
data that must be verified.

``QED`` defines a very simple but flexible trust model. It is composed of
three main components:

- The **information** itself to which the users want to add transparency.
- A set of **actors** that interacts with the information in different ways.
- A **mapping function** that translates the information space to a
  univocal event that can serve as input for ``QED``.

It is clear that the information depends on the nature of the problem and
likewise, the shape the mapping function takes is closely linked to it.
For their part, the actors can be grouped in three categories or roles:

- Sources of information.
- Information providers.
- Information consumers.

Let's see a brief example.

Suppose a scenario where bank customers want to ensure that every
money transfer related to their accounts can be verified later.

Here, the information takes the form of bank transfers which includes
references to the destination accounts, a timestamp, the amount of money
tranferred, a concept and probably a set of different internal metadata.

The involved actors are the bank and the customer. The customer plays the role
of the information consumer, and the bank plays both the roles of source and
provider (but might be divided in to different services: one for making
transfers and other one for querying them).

The mapping function then,
