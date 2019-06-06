Lie Detector in Tweeter feeds
=============================

Nowadays, with the boom of fake news, it could be interesting to detect 
inconsistencies between what was been published, and the original creation.

Theory and Operation
--------------------

In order to push tweets to QED, a tool like a stream publisher becames neccesary to
drain messages from Twitter and insert them into QED.

(see golang `go-twitter <https://github.com/dghubble/go-twitter/blob/master/examples/streaming.go>`_ module,
python's `tweeepy <http://docs.tweepy.org/en/v3.4.0/streaming_how_to.html>`_ library, or
npm `twitter <https://www.npmjs.com/package/twitter#streaming-api>`_ package `streaming-api` capabilities,
to create your own tool)

.. image:: /_static/images/Uc3.png

Event Source
++++++++++++

The **Streaming Publisher** tool would act as the event source.
It will need to process tweets data (username, date, text), and create the events that identifies individual tweets
univocally.

.. note::

    Mapping example of event ``F(TWEET)``:

    .. code:: json

        {
            "user_screen_name": "TwitterDev",
            "date": "22:01 - 6 may. 2019",
            "text": "Today's new update means that you can finally add Pizza Cat to your Retweet with comments! Learn more about this ne… https://t.co/Rbc9TF2s5X",
        }


Auditor
+++++++

The **Lie Detector** service would act as an auditor, giving transparency to 
any tweet source.

Untrusted Sources
+++++++++++++++++

Any source where you can fetch a tweet is susceptible to be used as untrusted
source. 

Creating transparency in tweets
-------------------------------

.. warning::

    The following snippets are atop :ref:`Quick start`. please visit it to
    configure the required code.

First, it is neccesary to process a tweet and create the event ``F(TWEET)`` to send it to QED:

.. code:: shell

    # Create the tweet event
    $ cat > tweet_event.json <<EOF
    {
        "user_screen_name": "TwitterDev",
        "date": "22:01 - 6 may. 2019",
        "text": "Today's new update means that you can finally add Pizza Cat to your Retweet with comments! Learn more about this ne… https://t.co/Rbc9TF2s5X",
    }
    EOF

Then, insert the tweet event to QED:

.. code:: shell

    # pushing the tweet event to QED server
    qed_client \
        add    \
        --event "$(cat tweet_event.json)"

And finally, retrieve and verify the proof:

.. code:: shell

    # Verify the proof
    qed_client                            \
        membership                        \
        --event "$(cat tweet_event.json)" \
        --auto-verify
