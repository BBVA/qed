Certification of Documents, Emails, Contracts, etc.
===================================================

How can we create transparency about a transaction that somebody as interest to
certify that a ``DOCUMENT`` is emitted that way?


Theory and Operation
--------------------

.. tip::

    for the sake of clarity **Document** is *anything* that could be suitable
    to keep track of the original content, such as Emails, Contracts, Dues,
    etc...

Allowing QED to store an event of the transaction ``F(DOCUMENT)``, will
prove that it was as intended.

Furthermore the proof returned by the QED server it is a cryptographic proof
``WARRANT`` with legal validity.

QED is a **tamper evident** storage, that is that QED it can be deployed in
untrusted servers, because of the way QED stores the transactions.

In this Use case we will try to explain in mundane terms why QED is worth the
effort to be used as warranteer.

.. image:: /_static/images/Uc2.png

Event Source
++++++++++++

Any **petitioner** interested to keep track of the tracked **DOCUMENT** Is
considered an event source in the model.

Mapping Documents to Events
^^^^^^^^^^^^^^^^^^^^^^^^^^^
To create the event we can use the SHA256 digest of the content to prove
inconsistencies across proofs.

.. note::

    Mapping example:

    .. code:: json

        {
            "digest": "4b1a0b7be7b5982dc778e76adacbb6348632ff4d",
        }


Auditor
+++++++

To check if the retrieved document is the same as the originally emitted the
**warranteer** service will ask QED for proofs that is untampered.

Untrusted Sources
+++++++++++++++++

Any **storage** that kept the document will be considered unsafe, and QED
event proofs will provide transparency to them.

Creating transparency receiving a Document
------------------------------------------

.. warning::

    The following snippets are atop :ref:`Quick start`. please visit it to
    configure the required code.

Once we emit the ``DOCUMENT`` can create the event ``F1(DOCUMENT)`` by using
the content of the file.

.. code:: shell

    # Create the document event
    document_hash=$(sha256sum <document> | cut -d' ' -f1 )
    cat > document_event.json <<EOF
    {
        "document_hash": "${document_hash}",
    }
    EOF

Push the document event to QED.

.. code:: shell

    # pushing the document event to QED server
    qed_client \
        add \
        --event "$(cat document_event.json)"

And Finally retrieve and verify the proof.

.. code:: shell

    # Verify the proof
    qed_client \
        membership \
        --event "$(cat document_event.json)" \
        --auto-verify
