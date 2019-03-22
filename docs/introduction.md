## Introduction

QED implements a forward-secure append-only persistent authenticated data structure. Each append operation produces as a result a cryptographic structure (a signed snapshot), which can be used to verify:

 * if the data was appended or not
 * whether the appended data is consistent, in insertion order, to another entry

To verify both statements we need the snapshot, the piece of data inserted and a QED proof. QED emits a snapshot every time an entry is inserted, and they need to be accessible and indexed elsewhere. Alsol QED does not verify nor emit unsolicited proofs, it’s the user responsibility to know when and how to verify the data depending on their needs.

Lastly, QED does not store the data itself, only a representation of it produced by a collision-resistant hash function. A system to send the data or a representation of it to QED is also not provided.

The semantic of the appended data and the relation between each item appended is also a client responsibility.

In this document we will describe use cases for the technology.

## Why?

There are multiple technologies to achieve simmilar functionality as QED, such as signed data in a database, or block chain’s related structures. The advantages of the data structure QED implements are:

 * scalability to thousands of millions of entries
 * proof of membership or non-membership generation in logarithmic time
 * proofs of logarithmic size
 * proof of temporal consistency related to QED insertion time

## Use cases

### Dependency management: authenticating a repository of software

A team of developers works on a software project stored in a repository. An automated software construction pipeline  builds and packages the work of the team. IT downloads multiple third-party repositories containing software or artifacts needed in the build pipeline

How can we check if a dependencie has been tampered?

 * the event source will be an artifacts download page
 * the application will be the pipeline construction software

Different stakeholders manage each component following their own regulations. Note that QED log and snapshot store are managed by different teams too.

The workflow we are assuming comprise the following steps:
 * developers commit code to the repository of code, including dependencies information such as location and version
 * on each commit, the pipeline build the software: it will fetch all the dependencies, compile them and build the project
 * each build will pass a set of tests and at the end of the pipeline, the generated code is stored in another repository or artifactory.
 
One question that can arise is: is the dependency downloaded legit or has been changed without our knowledge?

We can leverage QED to verify the history of an artifact ensuring our dependencies are correct, failing the construction in the pipeline in case one dependency has been changed.

In this scenario we can also contemplate multiple teams working on multiple software projects with overlapping dependencies which uses a single QED log and snapshot store.

Suppose for each dependency we have the following data in JSON format:

    {
        “pkg_name”: “rocksdb”,
         “pkg_version”: “vX.XX.X”,
        “pkg_digest”: “xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx”
    }


#### Tampering the dependency repository

The first scenario contemplates the dependency distribution has been compromised, using a corrupted version instead of a legit one with the same metadata as the original.

* Generate package data and add it to QED
* Publish the artifact in the repository for others to use when the data was already inserted in QED
* Users wanting to use that artifact as a dependency in their pipelines will need to
  - download the artifact
  - calculate the data that was inserted into QED
  - get a membership proof QED
  - verify the proof using the published snapshot
 
##### The happy path

    $ ./generate.sh
    usage:
    ./generate.sh pkg_name pkg_version pkg_url
    $ ./generate.sh qed 0.2 https://github.com/BBVA/qed/releases/download/v0.2-M3/qed_0.2-M3_linux_amd64.tar.gz > newentry.json
    Package
    {
        “pkg_name”: “qed”,
        “pkg_version”: “0.2”,
        “pkg_digest”: “5f6486ff5d366f14f9fc721fa13f4c3a4487c8baeb7a1e69f85dbdb8e2fb8ab2”
    }
    saved in /tmp/tmp.GS5kwBUYJu
    $ ./append.sh newentry.json
   
    Adding key [ {
        “pkg_name”: “qed”,
        “pkg_version”: “0.2”,
        “pkg_digest”: “5f6486ff5d366f14f9fc721fa13f4c3a4487c8baeb7a1e69f85dbdb8e2fb8ab2”
    } ]

    Received snapshot with values:

     EventDigest: f56c757b5403c71ced0773684a259c7a2dcde6e4232b251ceae5084d58ff356e
     HyperDigest: bcca38e67883f492b8dece031290a4b1b5cfa801d9917670f419b183487163be
     HistoryDigest: 784674a832f41ff7b9ddc13bdb2aef2093975319c5f23b69f04fbae163668975
     Version: 0
    
     $ ./membership.sh 0 newentry.json
    
    Querying key [ {
        “pkg_name”: “qed”,
        “pkg_version”: “0.2”,
        “pkg_digest”: “5f6486ff5d366f14f9fc721fa13f4c3a4487c8baeb7a1e69f85dbdb8e2fb8ab2”
    } ] with version [ 0 ]
   
    Received membership proof:
   
     Exists: true
     Hyper audit path: <TRUNCATED>
     History audit path: <TRUNCATED>
     CurrentVersion: 0
     QueryVersion: 0
     ActualVersion: 0
     KeyDigest: f56c757b5403c71ced0773684a259c7a2dcde6e4232b251ceae5084d58ff356e
   
    Please, provide the hyperDigest for current version [ 0 ]: bcca38e67883f492b8dece031290a4b1b5cfa801d9917670f419b183487163be
    Please, provide the historyDigest for version [ 0 ] : 784674a832f41ff7b9ddc13bdb2aef2093975319c5f23b69f04fbae163668975
   
    Verifying with Snapshot:
   
     EventDigest:f56c757b5403c71ced0773684a259c7a2dcde6e4232b251ceae5084d58ff356e
     HyperDigest: bcca38e67883f492b8dece031290a4b1b5cfa801d9917670f419b183487163be
     HistoryDigest: 784674a832f41ff7b9ddc13bdb2aef2093975319c5f23b69f04fbae163668975
     Version: 0
   
    Verify: OK

##### Lifecycle of a single dependency: tampering the source of a past release

Create a timeline of a single dependency, for example, for Facebook database Rocksdb, we add the following versions to QED:
   
    v5.13.3
    v5.12.5
    v5.13.4
    v5.14.2
    v5.14.3
    v5.15.10
    v5.16.6
    v5.17.2
    v5.18.3

in order from old to new.

Generate a data entry for each version and add it to QED:

    $ for i in v5.13.3 v5.12.5 v5.13.4 v5.14.2 v5.14.3 v5.15.10 v5.16.6 v5.17.2 v5.18.3; do
        ./generate.sh rocksdb ${i} https://github.com/facebook/rocksdb/archive/${i}.zip > rocksdb/${i}.json
    done

Append the data to a QED server:
   
    $ for i in v5.13.3 v5.12.5 v5.13.4 v5.14.2 v5.14.3 v5.15.10 v5.16.6 v5.17.2 v5.18.3; do ./append.sh rocksdb/$i.json; done

Now we have a QED with 9 rocksdb versions ordered by release date from older to newer.

Also we have 9 snapshots in a third party public store, so we can get anytime the snapshot of each version.

In this example we suppose we are using the version v5.16.6. In our pipeline we build our software depending on that version of rocksdb. The software distributor or repository has been compromised, a new package has been uploaded to the official repository, with an old version but new contents and new download hashes.

Ir our pipeline, every time we download a dependency, we generate its QED entry, but this time, because someone altered the package, we generate a different version of the entry:

    rocksdb/corrupted-v5.16.6.json
   
Before using it to build our software, we ask QED for a membership proof, so we can verify that our download verifies the QED membership proof:

    $ ./membership.sh 6 rocksdb/corrupted-v5.16.6.json

To verify the event, we need the correct snapshot information. Because we’re using the interactive client, it ask us the hyperDigest for the current version of the QED. We go to the snapshot store to get it:

    $ ./getsnapshot.sh 8
   
We use the hyperdigest presented here, and the client tries to verify the information, but it fails.

As we can see, the QED tells us that the information was not on QED and the client verified that there is no such event given the cryptographic information published in the insertion time of the event. With this information we can alert the alteration of one of our dependencies and stop the pipeline alerting the devops team of the issue.

We have authenticated a third party repository (a github release repository) which is the source of the events, then we have inserted into QED the information related to a set of verified releases of rocksdb. Later, the repository was the target of an attack and a changed version of a dependency was downloaded by the software construction pipeline. Our dependency check phase checked the information against QED and discovered a tamper in the remote software repository.

This simple scenario can be implemented by just storing the hashes into the repository and checking against them when downloading the dependency. Most package management tools like go mod, npm, cargo, etc. use dependency package files containing hashes of the dependency version that must be used in the construction.

Having an external validation tool is useful in situations when we need resistance to tampering. As a reference example, in the case of [the event-stream npm case](https://snyk.io/blog/malicious-code-found-in-npm-package-event-stream/), their source code repository was compromised, and a new version was generated with malicious code in it.

##### Tampering the source: github code repository

Using the last scenario as a starting point, we have now the situation on which our source code repository has been compromised and a new download URL and hash has been provided to our package manager.

Here, our pipeline will download the new dependency and will generate the entry for QED. But this entry was not in QED, so the check will fail the verification process.

This scenario assumes that only authenticated developers can insert entries to QED, and with their digital signature they provide a personal warranty the entry is legit.

QED do not detect vulnerabilities, nor validate the correctness of changes in software,  but can support a process involving multiple parties to be more resistant to them.

#####  Tampering the application: builder pipeline

Here, the builder system has been compromised, and instead of building the software as programmed, it will build a special release containing arbitrary code and dependencies. Also it will be changed to only ask QED non-modified dependencies.

Because of the pipeline tampering, QED will not be used in the pipeline, and it only can work if someone asks for the proofs, and if those proofs contain the metadata to discover the tampering.

We can leverage the gossip agents platform included in QED to build external tools. For example a special proxy to detect unwanted dependencies downloads. This proxy server will be in charge of outgoing HTTP connections to the internet and will check against QED all the package URLs before being downloaded.

#####  Tampering the QED log

The QED server log stores its cryptographic information in a local database that is replicated against the other nodes of the QED cluster using the RAFT protocol.

In our development version of QED we have included a special endpoint to change the database without using the QED API or stopping the server which would be the worst scenario in case of an attack against QED.

In QED we can change either the hyper or the history tree. The hyper tree is a sparse Merkle tree with some optimizations that maps a piece of data with is the position in the history tree. This history tree is an always growing complete binary tree and stores the entries ordered by its insertion time.

The hyper tree will tell us if an event changed or not and will give us a proof which we can use to verify it. The history tree will tell us if the event is consistent with the stored order of another entry, and will give us a verifiable proof.

Using as a starting point the rocksdb example, our history tree contains:

    v5.13.3  v5.12.5  v5.13.4  v5.14.2  v5.14.3  v5.15.10 v5.16.6  v5.17.2  v5.18.3   
    .________.________.________.________.________.________.________.________.


We create a fork if we insert three new events into QED, into the same positions of old events, and generate new snapshots for them.

It will not work changing the hyper tree only as the history will contain the correct hash and the membership query will not validate. Also, we can’t change just the history leaf in its storage because all the intermediate nodes will not be updated, and the generated audit paths will be unverifiable.

A way to rewrite the QED is to insert again new events resetting the balloon version to the one we want to change. In a RAFT cluster, the hijacked node needs to preserve the raft log order to tamper the data without corrupting the RAFT server making it crash. We also need to change all the events from that point in history to the end to have the same history size than before.

We insert into QED three new events, v5.16.6’ v5.17.2’ v5.18.3’, but instead of using the regular API, we will suppose we have a custom QED implementation that inserts in the hyper tree the version we want.

    $ ./resetVersion.sh 6
    $ ./append.sh rocksdb/v5.16.6p.json
    $ ./append.sh rocksdb/v5.17.2p.json
    $ ./append.sh rocksdb/v5.18.3p.json
   

Our forked history tree now looks like:

    v5.13.3  v5.12.5  v5.13.4  v5.14.2  v5.14.3  v5.15.10 v5.16.6’ v5.17.2’ v5.18.3’
    .________.________.________.________.________.________.________.________.

if we use the versions stored in the hyper tree.

In this situation we will download the compromised dependency, v5.16.6’ and ask QED for it, like we did before:

    $ ./membership.sh 6 rocksdb/v5.16.6p.json

And it will verify it.

To detect this,  we can check if we have two entries of the snapshot in the snapshot store. Also, we can implement processes attached to the gossip network to test if a snapshot n has been seen twice with the same version but different hashes.

#####  Tampering the QED snapshot store

QED does not provide an implementation for the snapshot store, just the HTTP API to be implemented to store and retrieve snapshots.

Because the tampering procedure explained before, we advise selecting an append only database with an auto-generated primary key, so it is possible to detect multiple snapshots for the same version.

In case someone deletes an entry from the snapshot store, the event inserted in that version won’t validate.

Also if someone can control the log and the snapshot store, it is possible to fake a verifiable history, but will fail the verification of the event source. The only way to avoid detection is to control all the parts of the system: the event source, the QED log and the snapshot store. This is equivalent to deploy a new QED with a custom event inserted.

### Transferences transparency: moving money between different institutions

Two banking entities transfer money between them, we will have:

 * two entities, each one with its own QED log
 * a common QED snapshot store
 * an application which will generate a QED entry for each money transfer order
 * an accountability process

A client of bank A makes a transfer order for user B of Bank B. How can those banks agree in what transactions have been made in a verifiable way?

 * the event source will be the transfer orders and its associated metadata
 * the application will be the related bank applications that operate the transfer orders

##### Tampering the source
#####  Tampering the application
#####  Tampering the QED log
#####  Tampering the QED snapshot store

### User activity non-repudiation: log the history of actions

A client of Bank A acquires new banking products. Those products are subject to risks. And at some point the client decides to accuse the Bank A of cheating in the operations he did. How can the Bank A prove he is lying?

* the event source will be the client signed contracts and the accounting changes and it reasons
* the application will be the related bank applications that operates and generates the accounting data and its changes

##### Tampering the source
#####  Tampering the application
#####  Tampering the QED log
#####  Tampering the QED snapshot store