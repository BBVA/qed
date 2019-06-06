Commit certification
====================

In this use case we will show how to add transparency to a simple software
deployment pipeline, starting from the source code a developer commits
to a source repository, and ending with the deployment of the corresponding
built artifact.

This way, the developer can ensure that what he intented to publish is what
it was finally deployed.

Theory and operation
--------------------

In order to add transparency to the process we will need to identify firstly
what are the elements of our trust problem and then try to adapt them to the
components defined in our :ref:`QED's trust model<trust_model>`: information,
actors and mapping function(s).

.. image:: /_static/images/Uc1.png

As we can see from the figure, there are two kinds of information to which we
need to add transparency: the original source code commited by the developer
and the binary artifact built by the CI tool.

In this case, the actors are multiple (developer, repositories and pipeline
processes) and some of them take different roles depending on the step of
the pipeline being executed.

Let's explain the process in detail.

First step: source committing
+++++++++++++++++++++++++++++

We have an actor, the **developer**, that takes the role of source of
information. He makes some changes to the source code and commits them to
the Git repository. The repository will therefore be the infomation provider
in our trust model and the first component we want to add transparency to.
Every consumer of that repository will need some kind of proof to verify its
integrity.

To achive this, the developer can use a particular mapping function ``F1`` that
translates the resulting source code to a unique QED event. But first, we need
to identify what makes it unique.

For this event, the original commit hash and a SHA256 digest of all files
(excluding the .git folder) will provide a concise information that will vary
whenever even a single character gets changed.

.. note::

    ``F1`` output example:

    .. code:: json

        {
            "commit_hash": "4b1a0b7be7b5982dc778e76adacbb6348632ff4d",
            "src_hash": "b9261acdcc979434d37ed8211ad6014309752cb6a02705a40dc8dbaf9cdcd89b",
        }

Then, the developer can take the event resulting after applying the function to
the source code ``F1(SOURCE)``, and insert it into the QED Log.

Second step: artifact building
++++++++++++++++++++++++++++++

Once the source code has been committed to the repository, a hook fires the
**build** phase of the pipeline which downloads the source code and generates a
binary artifact. The build process acts here as the consumer actor in the trust
model and thus, needs to have confidence in the integrity of the repository.

To do that, it could use the same mapping function ``F1`` to generate again the
QED event and then request a membership query to the QED Log. With the
resulting cryptographic proofs and the QED event, it could verify the
original information, the source code, as valid.

Third step: uploading artifact
++++++++++++++++++++++++++++++

Now, the build process comes from acting as a consumer to take the role of
source of information. The binary artifact is now the information we want
to verify and the artifact repository becomes the new information provider.

Thus, the build process has to use a new mapping function ``F2`` to
translate the resulting artifact to a unique QED event ``F2(BINARY)``,
and then, insert such event into the QED Log.

For this function, the SHA256 digest of the binary file, will be simple
and good to detect changes.

.. note::

    ``F2`` output example:

    .. code:: json

        {
            "artifact_hash": "pcdcc979434d37e4b1a0b4309752cb6a0277c778e76adacbb6348632ff4d",
        }


Fourth step: artifact deployment
++++++++++++++++++++++++++++++++

Once the binary artifact has been uploaded to the artifact repository,
a new hook fires the **deploy** phase of the pipeline which downloads the
binary file and deploys it to the corresponding environment. Now, the deploy
process acts as the consumer actor in the trust model that needs to have
confidence in the integrity of the artifact repository.

To achieve that, it must use the same mapping function ``F2`` to generate
the corresponding QED event in order to request a membership proof from
the QED Log. Again, combining the resulting cryptographic proofs with the QED
event, the process could verify the original information as valid.


Working example
---------------

Adding transparency to a GIT repository
+++++++++++++++++++++++++++++++++++++++

.. warning::

    The following snippets assume a working QED installation. Please refer
    to the :ref:`Quick start` page.


The following snippet simulates the creation of a QED event starting from
the source code recently committed. As mentioned before, we are using the
**commit_hash** and the **source_hash** as the output of the mapping function
``F1(SOURCE)`` to unambiguously identify a source code.

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

Alongside pushing the code to the git repo, the developer (or a githook) adds
the event to the QED Log.

.. code:: shell

    # pushing the event to QED server
    qed_client \
        add \
        --event "$(cat event.json)"

Once the QED stores the event, the ``BUILD`` stage will fetch the source code
from the git repo and, just before building the binary artifact, generate
again the QED event to request a membership proof to QED Log. After verifying
the integrity of the source code at the repository, it will continue with
the next step.

.. code:: shell

    # Verify the proof
    # please note the --auto-verify flag, without this flag the operation
    # will returns the cryptographic proof
    qed_client \
        membership \
        --event "$(cat event.json)" \
        --auto-verify

Adding transparency to the artifacts repository
+++++++++++++++++++++++++++++++++++++++++++++++

Once the BUILD stage creates the ``BINARY`` file, it applies the mapping
function ``F2(BINARY)`` to the file and obtains a new QED event.

.. code:: shell

    # Create the artifact event
    artifact_hash=$(sha256sum archived/gin | cut -d' ' -f1 )
    cat > bin_event.json <<EOF
    {
        "artifact_hash": "${artifact_hash}",
    }
    EOF

Alongside pushing the binary artifact to the repository it adds the event to
the QED Log. As you can see, there is a repeating pattern of
``source -> [QED|Untrusted-source] <- auditor`` in the way QED creates the
transparency.

.. code:: shell

    # pushing the artifact event to QED server
    qed_client \
        add \
        --event "$(cat bin_event.json)"

And finally, the DEPLOY stage can request again a proof from the QED Log
and verify the integrity of the artifact before deploying it.

.. code:: shell

    # Verify the proof
    qed_client \
        membership \
        --event "$(cat bin_event.json)" \
        --auto-verify
