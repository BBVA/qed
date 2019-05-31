Lie Detector in Tweeter feeds
=============================

.. image:: /_static/images/Uc3.png

Can we know instantly (to some extent) if a tweet is real or not?

Day after day we faced the necessity to detect inconsistencies between what was
published, was the original creation, and if is true what different lobbies
tries us to sell.

In this **Use case** we will discuss the high throughput that is required in
order to allow to detect inconsistencies in real time.

Logarithmic long tail
---------------------

Since QED is an cryptographic append-only storage, it has some design
capabilities and limitations. One of those it that the `Merkle Tree`_ has
