Contributing
============

You can contribute in a few different ways:

- Submit issues through our issue tracker_ on Github.
.. _tracker: https://github.com/bbva/qed/issues

- If you wish to make code changes, please check out above our guidelines about **Pull Requests** and the GitHub Forks/PullRequests model_.
.. _model: https://help.github.com/articles/fork-a-repo/

Pull requests
=============

We have stablished a **work agreement** to provide a linear history, with at most one branch in parallel.
We also require all commits in master to pass the tests.

For that to happen, this steps will enable you to get your pull request ready for being merged into the **master branch**.
TL;DR: Always rebase to master before attempting to merge into master.

    .. code::

        # download repo
        git clone git@github.com:bbva/qed.git

        # enter project dir
        cd qed

        # create your fork
        hub fork

        # create a to branch to hack on
        git checkout -b my-cool-feature-branch

        # ... do some groovy changes
        git commit -am 'some explanatory although a bit cryptic msg ;-P'

        # ensure your changes are in your github fork
        git push my-user my-cool-feature-branch

        # create PR
        hub pull-request --base bbva:master

        # wait for feedback (possibly master will advance in the meantime)
        git commit ...
        git commit ...
        git commit ...
        git push ...

        # once it's approved and ready to merge, rebase to master and resolve all conflicts.
        git fetch origin master
        git rebase origin/master

        # push rebased branch to your fork (the PR will be updated automatically)
        git push --force-with-lease my-user my-cool-feature-branch

        # check again that tests are ok, and then merge (this step can only be performed by developers with write access to the repo)
        hub merge https://github.com/bbva/pr/pull/42
