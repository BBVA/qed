Commit certification
====================

In this use case we will show how to certify artifacts from source code,
building exactly what the developer intended to publish.

Theory and Operation
--------------------

Building **trust** around storage that is unattended it's a cumbersome task.
Create a certified untamperable repository it's near impossible to achieve.
Not to mention third-party storages, like github.

In order to create the transparency we will need some actors being the **event source** that **auditors** will need to detect tamperings.

In this Use Case QED allows transparency in that regard, by allowing developers
publish both code and a event ``F1(SOURCE)`` (more on this later...), and
the ``F2(BINARY)``.

.. image:: /_static/images/Uc1.png

Event Sources
+++++++++++++

The **DEVs** acts as event source for the transparency of the GIT REPO,
and the **BUILD** stage, creator of the artifact, acts as event source for
the transparency of the ARTIFACT REPO.

We will use two mapping functions ``F1`` and ``F2`` to identify the original
content as an unique **event**

Mapping Source Code to ``F1`` event
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

In orden to translate abstract content to a **QED event** we need to identify
what makes it unique, and how can we detect changes if the original content is
tampered.

For this event, the original commit hash, and a SHA256 of the files (excluding
.git folder) will provide a concise information that will change when even a
single character is changed.

.. note::

    ``F1`` event example:

    .. code:: json

        {
            "commit_hash": "4b1a0b7be7b5982dc778e76adacbb6348632ff4d",
            "src_hash": "b9261acdcc979434d37ed8211ad6014309752cb6a02705a40dc8dbaf9cdcd89b",
        }



Mapping Binaries to ``F2`` event
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

For this event, the SHA256 digest of the binary, will be simple and good to
detect changes.

.. note::

    ``F2`` event example:

    .. code:: json

        {
            "artifact_hash": "pcdcc979434d37e4b1a0b4309752cb6a0277c778e76adacbb6348632ff4d",
        }

Auditors
++++++++

The **BUILD** stage will act as an auditor before the creation of the artifact.
The **DEPLOY** stage will audit the binary in order to create the trust in
the ARTIFACT REPO.

Untrusted sources
+++++++++++++++++

Both **GIT REPO** and **ARTIFACT REPO** alongside with the **QED** are untrusted
sources. We create the trust with the auditors that verifies the original event
source as valid.


Creating transparency in a GIT repository
-----------------------------------------

.. warning::

    The following snippets are atop :ref:`Quick start`. please visit it to
    configure the required code.

Creating a event is crucial to allow **auditors** generate trust around
the repository that we need to rely on.

In this example we using the **commit_hash** and the **source_hash** to
univocally identify a source code.

.. code:: shell

    # Create the source code event
    commit_hash=$(git rev-parse HEAD)
    src_hash=$(echo $(find . -type f -not -path "./.git/*" -exec sha256sum {} \; | sort -k2) | sha256sum | cut -d' ' -f1)
    cat > event.json <<EOF
    {
        "commit_hash": "${commit_hash}",
        "src_hash": "${src_hash}",
    }
    EOF

Alonside publishing to the git repo (or using a githook) now you can push the
event to QED.

.. code:: shell

    # pushing the event to QED server
    qed_client \
        add \
        --event "$(cat event.json)"

Once the QED stores the event event ``F1(SOURCE)``, it will be verified
and proved only and only if the code retrieved is exactly the same. This will prove
untampered once the ``BUILD`` stage fetch the source code from the git repo.

.. code:: shell

    # Verify the proof
    # please note the --auto-verify flag, without this flag the operation
    # will returns the cryptographic proof
    qed_client \
        membership \
        --event "$(cat event.json)" \
        --auto-verify

Creating transparency in the Artifacts Repository
-------------------------------------------------

Once we create the ``BINARY`` in the BUILD stage we can create the event
``F2(BINARY)`` by using the content of the file.

.. code:: shell

    # Create the artifact event
    artifact_hash=$(sha256sum archived/gin | cut -d' ' -f1 )
    cat > bin_event.json <<EOF
    {
        "artifact_hash": "${artifact_hash}",
    }
    EOF

And push the binary event to QED alonside to push the binary to the Artifact
repo. Ad you can see there is a repeating pattern of ``event-source -> [QED|Untrusted-source] <- auditor`` in the
way QED creates the transparency.


.. code:: shell

    # pushing the artifact event to QED server
    qed_client \
        add \
        --event "$(cat bin_event.json)"

And Finally verify the proof.

.. code:: shell

    # Verify the proof
    qed_client \
        membership \
        --event "$(cat bin_event.json)" \
        --auto-verify
