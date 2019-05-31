Commit certification
====================

In this Use Case we will discuss how to certify artifacts from source code,
building exactly what the developer intended to publish.

Theory and Operation
--------------------

Building **trust** around storage that is unattended it's a cumbersome task.
Create a certified untamperable repository it's near impossible to achieve.
Not to mention third-party storages, like github.

In order to create the transparency we will need some actors being the **event source** that **auditors** will need to detect tamperings.

In this Use Case QED allows transparency in that regard, by allowing developers
publish both code and a fingerprint ``F1(SOURCE)`` (more on this later...), and
the ``F2(BINARY)``.

.. image:: /_static/images/Uc1.png

Event Sources
+++++++++++++

The **DEVs** acts as event source for the transparency of the GIT REPO,
and the **BUILD** stage, creator of the artifact, acts as event source for
the transparency of the ARTIFACT REPO.

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

.. important::

    Visit 'Installation' and 'Quick start' sections before typing the following code snippets.

Creating a fingerprint is crucial to allow **auditors** generate trust around
the repository that we need to rely on.

In this example we using the **commit_hash** and the **source_hash** to
univocally identify a source code.

.. code:: shell

    # Create the source code fingerprint
    commit_hash=$(git rev-parse HEAD)
    src_hash=$(echo $(find . -type f -not -path "./.git/*" -exec sha256sum {} \; | sort -k2) | sha256sum | cut -d' ' -f1)
    cat > fingerprint.json <<EOF
    {
        "commit_hash": "${commit_hash}",
        "src_hash": "${src_hash}",
    }
    EOF

Alonside publishing to the git repo (or using a githook) now you can push the
fingerprint to QED.

.. code:: shell

    # pushing the fingerprint to QED server
    qed client \
        add \
        --event "$(cat fingerprint.json)"

Once the QED stores the fingerprint event ``F1(SOURCE)``, it will be verified
and proved only and only if the code retrieved is exactly the same. This will prove
untampered once the ``BUILD`` stage fetch the source code from the git repo.

.. code:: shell

    # Verify the proof
    # please note the --auto-verify flag, without this flag the operation
    # will returns the cryptographic proof
    qed client \
        membership \
        --event "$(cat fingerprint.json)" \
        --auto-verify

Creating transparency in the Artifacts Repository
-------------------------------------------------

Once we create the ``BINARY`` in the BUILD stage we can create the fingerprint
``F2(BINARY)`` by using the content of the file.

.. code:: shell

    # Create the artifact fingerprint
    artifact_hash=$(sha256sum archived/gin | cut -d' ' -f1 )
    cat > bin_fingerprint.json <<EOF
    {
        "artifact_fingerprint": "${artifact_hash}",
    }
    EOF

And push the binary fingerprint to QED alonside to push the binary to the Artifact
repo. Ad you can see there is a repeating pattern of ``event-source -> [QED|Untrusted-source] <- auditor`` in the
way QED creates the transparency.


.. code:: shell

    # pushing the artifact fingerprint to QED server
    qed client \
        add \
        --event "$(cat bin_fingerprint.json)"

And Finally verify the proof.

.. code:: shell

    # Verify the proof
    qed client \
        membership \
        --event "$(cat bin_fingerprint.json)" \
        --auto-verify
