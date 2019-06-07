Lie Detector for Tweeter feeds
==============================

Nowadays, with the boom of fake news, it could be interesting to detect
inconsistencies between the contents that were originally published in a
particular media and what is currently accessible to the public.

In this use case, we will show how to add transparency to messages
published in a social network like Twitter, by allowing the users to
verify that already published tweets have not been altered.

Theory and Operation
--------------------

First of all, we need to identify what are the elements of the problem
to address and how we can adapt them to the components defined in our
:ref:`QED's trust model<trust_model>`: information, actors and
mapping function(s).

.. image:: /_static/images/Uc3.png

As we can see from the figure, the information we want to add transparency to,
is the set of tweets published by one or multiple users. Those users are
the actors interested in keeping track of the contents of their own
publications, so they take the role of **sources of information**.

The tweets get inserted into the internal storage operated by Twitter, Inc.
This storage acts as the **information provider** in out trust model and,
by definition, is considered unstrusted. The way users interact with this
provider is through its public API.

In order to push events to QED, a tool like a ``STREAMING PUBLISHER`` becames
necessary to drain messages from Twitter.

.. note::

    See golang `go-twitter <https://github.com/dghubble/go-twitter/blob/master/examples/streaming.go>`_ module,
    python's `tweeepy <http://docs.tweepy.org/en/v3.4.0/streaming_how_to.html>`_ library, or
    npm `twitter <https://www.npmjs.com/package/twitter#streaming-api>`_ package `streaming-api` capabilities,
    to create your own tool.


Such ``STREAMING PUBLISHER`` tool would use a mapping function ``F`` to
translate the tweets contents to a unique QED event ``F(TWEET)``. Tweets
data and metadata (like username, date and text) could serve to identify
each tweet unambiguously.

.. note::

   ``F`` output example:

    .. code:: json

        {
            "user_screen_name": "TwitterDev",
            "date": "22:01 - 6 may. 2019",
            "text": "Today's new update means that you can finally add Pizza Cat to your Retweet with comments! Learn more about this ne… https://t.co/Rbc9TF2s5X",
        }


Finally, the ``LIE DETECTOR`` service would act as the
**information consumer** in the trust model, and will audit the
information provided by Twitter's public APIs.

To do that, it could use the same mapping function ``F`` to generate
again the QED event and then, ask for a membership proof to the QED Log.
Combining the resulting cryptographic proofs with the QED
event, the ``LIE DETECTOR`` could verify the original information as valid.

Working example
---------------

.. warning::

    The following snippets assume a working QED installation. Please refer
    to the :ref:`Quick start` page.

The following snippet simulates the creation of a QED event starting from
a particular tweet recently published. As mentioned before, we are applying
a mapping function ``F(TWEET)`` to some data and metadata from the tweet.

.. code:: shell

    # Create the tweet event
    $ cat > tweet_event.json <<EOF
    {
        "user_screen_name": "TwitterDev",
        "date": "22:01 - 6 may. 2019",
        "text": "Today's new update means that you can finally add Pizza Cat to your Retweet with comments! Learn more about this ne… https://t.co/Rbc9TF2s5X",
    }
    EOF

Then, we insert the event into QED Log:

.. code:: shell

    # pushing the tweet event to QED server
    qed_client \
        add    \
        --event "$(cat tweet_event.json)"

Finally, we can generate again the QED event to request a membership
proof from QED Log and verify the proof.

.. code:: shell

    # Verify the proof
    qed_client                            \
        membership                        \
        --event "$(cat tweet_event.json)" \
        --auto-verify
