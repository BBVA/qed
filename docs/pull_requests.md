# Why this guide

We follow a *work agreement* to provide a linear history, with at most one branch in parallel.
We also require all commits in master to pass the tests.

For that to happen this guide enable you get your pull request ready for merge into the **master branch**.

# How to make a Pull Request to QED

TL;DR: always rebase to master before attempting to merge into master.
```bash

# download repo
git clone git@github.com:bbva/qed.git

# create your fork
hub fork

# add main organization
git remote add -f bbva git@github.com:bbva/qed.git

# create branch
git checkout -b <my-cool-branch>

# ... do some groovy changes
git commit -am "msg"

# create PR
$ hub pull-request --base bbva:master

# wait for feedback (surely master will advance)

# once it's approved and ready to merge rebase and resolve conflicts.
git rebase master

# check again tests are ok, and then merge
hub merge https://github.com/bbva/pr/pull/13
```
