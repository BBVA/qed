Lie Detector in Tweeter feeds
=============================

Can we know instantly (to some extent) if a tweet is real or not?

Day after day we faced the necessity to detect inconsistencies between what was
published, was the original creation, and if is true what different lobbies
tries us to sell.

In this **Use case** we will discuss the high throughput that is required in
order to allow to detect inconsistencies in real time.


Theory and Operation
--------------------

In Order to push to QED the tweets, we will need a stream publisher which can be
created with any third-party library that has `streaming-api` capabilities like
golang `go-twitter <https://github.com/dghubble/go-twitter/blob/master/examples/streaming.go>`_ module,
python's `tweeepy <http://docs.tweepy.org/en/v3.4.0/streaming_how_to.html>`_ library or,
npm `twitter <https://www.npmjs.com/package/twitter#streaming-api>`_ package.

.. image:: /_static/images/Uc3.png

Event Source
++++++++++++

Since the **Streaming Publisher** will act as the event source it will need to
process all the tweets and creating the events that identifies individual tweets
univocally. And since tweets are small in size, there's no need to digest it 
beforehand.

.. note::

    Mapping example of event ``F(TWEET)``:

    .. code:: json

        {
            "user_screen_name": "TwitterDev",
            "text": "Today's new update means that you can finally add Pizza Cat to your Retweet with comments! Learn more about this ne… https://t.co/Rbc9TF2s5X",
        }


Auditor
+++++++

The **LIE DETECTOR** service will act as the auditor giving transparency to 
any tweet source. 

Untrusted Sources
+++++++++++++++++

Any source where you can fetch a tweet are susceptible to be used as untrusted
source to give transparency by auditing the event tweets. 

Creating transparency in tweets
-------------------------------

.. warning::

    The following snippets are atop :ref:`Quick start`. please visit it to
    configure the required code.

We will process a tweet and creating his event ``F(TWEET)`` to send it to QED.

.. code:: shell

    # Create the tweet event
    cat > tweet_event.json <<EOF
    {
        "user_screen_name": "TwitterDev",
        "text": "Today's new update means that you can finally add Pizza Cat to your Retweet with comments! Learn more about this ne… https://t.co/Rbc9TF2s5X",
    }
    EOF

Push the tweet event to QED.

.. code:: shell

    # pushing the tweet event to QED server
    qed_client \
        add \
        --event "$(cat tweet_event.json)"

And Finally retrieve and verify the proof.

.. code:: shell

    # Verify the proof
    qed_client \
        membership \
        --event "$(cat tweet_event.json)" \
        --auto-verify
