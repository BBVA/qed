## Introduction

QED is an implementation of a forward-secure append-only persistent authenticated data structure. Each append operation produces as a result a cryptographic structure (a signed snapshot), which can be used to verify:
 *  if the data was appended or not
 * whether the appended data is consistent, in insertion order, to another entry

The snapshot, the piece of data inserted and a QED proof are needed to verify both statements. The snapshots are emited proactively by QED and must be stored, accesible and indexed elsewhere. The proofs are not verified or emited by QED proactively, its the user responsibility to know when and how the data must be verified depending on their needs.

The data is not stored in QED, only a representation of it produced by a collision-resistant hash function. A system to send the data or a representation of it to QED is also not provided.

The semantic of the appended data and the relation between each item appended is also a client responsibility.

In this document we are going to describe some use cases for the technology.

## Why 

The functionality described can be achieved using other tecnologies such as digitally signed data in a database, or block chain's related structures. The advantages of the data structure implemented in QED are:
 * scalability to thousands of millions of entries
 * proof of membership or non-membership generation in logarithmic time
 * proofs of logarithmic size
 * proof of temporal consistency related to QED insertion time

## Use cases

### Dependency management: authenticating a repository of software

In this scenario we have:

 * a team of developers working on a software project stored in a repository
 * an automated software construction pipeline which builds and packages the work of the team
 * multiple third-party repositories containing software or artifacts needed in the build pipeline
 * a QED log
 * a QED snapshot store

Each system is managed by different tenants following their own regulations, and each team is only responsible for one of each asset. QED log ans snapsot store are managed by different teams too.

The workflow we are assuming consist of the following steps:
 * developers commit code to the respository of code, including dependencies information such as location and version
 * on each commit, the software is built by the automated system: this system will fetch all the dependencies, compile them and build the project
 * each build wil pass a set of tests and the the pipeline will be finished, generating an artifact that will be managed elsewhere
 
One question that can arise is: is the dependency downloaded legit or has been modified without my knowledge? 

We can leverage QED to verify the history of an artifact ensuring our dependencies are correct automatically, failing the construction in the pipeline in case one of the dependencies have been modified.

In this scenario we can also contemplate multiple teams working on multiple software projects with overlapping dependencies which uses a single QED log and snapshot store.

All the files generated during this test are present in the respository as an example to guide a possible production implementation, but the scripts and data are not meant to be used in production.


#### Tampering the dependency repository

The first scenario contemplates the dependency distribution has been compromised, using a corrupted version instead of a legit one with the same metadata as the original.

* Generate package data and add it to QED
* Publish the artifact in the artifactory for others to use when the data was already inserted in QED
* Users wanting to use that artifact as a dependency in their pipelines will need to
  -  download the artifact
  - calculate the data that was inserted into QED 
  - get a membership proof QED
  - verify the proof using the published snapshot

For this example, we have created some scripts to easy the operation:
 * generate.sh download an url and generates the output to be inserted into QED
 * append.sh reads from stdin the generate output and insert it into QED
 * membershit.sh read from stdin the generate output and insert it into QED
 
##### The happy path

    $ ./generate.sh 
    usage:
    ./generate.sh pkg_name pkg_version pkg_url
    $ ./generate.sh qed 0.2 https://github.com/BBVA/qed/releases/download/v0.2-M3/qed_0.2-M3_linux_amd64.tar.gz > newentry.json
    Package
    {
    	"pkg_name": "qed",
    	"pkg_version": "0.2",
    	"pkg_digest": "5f6486ff5d366f14f9fc721fa13f4c3a4487c8baeb7a1e69f85dbdb8e2fb8ab2"
    }
    saved in /tmp/tmp.GS5kwBUYJu
    $ ./append.sh newentry.json
	
    Adding key [ {
    	"pkg_name": "qed",
    	"pkg_version": "0.2",
    	"pkg_digest": "5f6486ff5d366f14f9fc721fa13f4c3a4487c8baeb7a1e69f85dbdb8e2fb8ab2"
    } ]

    Received snapshot with values:

     EventDigest: f56c757b5403c71ced0773684a259c7a2dcde6e4232b251ceae5084d58ff356e
     HyperDigest: bcca38e67883f492b8dece031290a4b1b5cfa801d9917670f419b183487163be
     HistoryDigest: 784674a832f41ff7b9ddc13bdb2aef2093975319c5f23b69f04fbae163668975
     Version: 0
     
     $ ./membership.sh 0 newentry.json
     
    Querying key [ {
    	"pkg_name": "qed",
    	"pkg_version": "0.2",
    	"pkg_digest": "5f6486ff5d366f14f9fc721fa13f4c3a4487c8baeb7a1e69f85dbdb8e2fb8ab2"
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

Create a timeline of a single dependency, for example for Facebook database Rocksdb, we add the following versions to QED:
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

Generate a data entry  for each version and add it to QED. Each data entry will have the form of a json message:

    {
    	"pkg_name": "rocksdb",
     	"pkg_version": "vX.XX.X",
    	"pkg_digest": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    }

    $ for i in v5.13.3 v5.12.5 v5.13.4 v5.14.2 v5.14.3 v5.15.10 v5.16.6 v5.17.2 v5.18.3; do
    	./generate.sh rocksdb ${i} https://github.com/facebook/rocksdb/archive/${i}.zip > rocksdb/${i}.json
    done
    
Append the data to a QED server:
    
    $ for i in v5.13.3 v5.12.5 v5.13.4 v5.14.2 v5.14.3 v5.15.10 v5.16.6 v5.17.2 v5.18.3; do ./append.sh rocksdb/$i.json; done

Now we have a QED with 9 rocksdb versions ordered by release date from older to newer.

Also we have 9 snapshots in a third party public store, so we can get anytime the snapshot of each version.

In this example we suppose we are using the version v5.16.6. In our pipeline we build our software depending on that version of rocksdb. The software distributor or repository has been compromised, a new package has been uploaded to the official repository, with an old version but new contents and new download hashes.

Ir our pipeline, every time we download a dependency, we generate its QED entry, but this time, because the package was altered by a third party, we generate a different version of the entry:

    rocksdb/corrupted-v5.16.6.json
    
Before using it to build our software we ask QED for a membership proof, so we can verify that our download was already verified and inserted into the QED:

    $ ./membership.sh 6 rocksdb/corrupted-v5.16.6.json

To verify the event we need the correct snapshot information. Because we're using the interactive client, it ask us the hyperDigest for the current version of the QED. We go to the snapshot store to get it:

    $ ./getsnapshot.sh 8
    
We use the hyperdigest presented here, and the client tries to verify the information:

    $ ./membership.sh rocksdb/corrupted-v5.16.6.json 6
    $ 

As we can see, the QED tells us that the information was not on QED and the client verified that there is no such event given the cryptographic information published in the insertion time of the event. With this information we can alert one of our dependencies was altered and stop the pipeline alerting the devops team of the issue.

In this scenario we have authenticated a third party repository (a github release repository) which is the source of the events, then we have inserted into QED the information related to a set of  verified releases of rocksdb. Later, the artifactory was the target of an attack and a modified version of a dependency was downloaded automatically by the software construction pipeline. Our dependency check phase checked the information against QED and discovered a tamper in the remote software repository.

This simple scenario can be implemented by just storing the hashes into the repository and checking against them when downloading the dependency. Furthermore, most package management tools like go mod, npm, cargo, etc. use dependency package files containng hashes of the depndency version tha must be used in the construction.

##### Tampering the source code repository

Using the last scenario as a starting point, we have now the situation on which our source code respository has been compromised and a new download url and hash has been provided to our package manager.

In this case, our building pipeline will download the new dependency and will generate the entry for QED. But this entry was not inserted into QED, so the check will again fail.

This scenario assumes that only autheticated developers can insert entries to QED, and with their digital signature they provide a personal warranty the entry is legit.

#####  Tampering the builder pipeline

In this case, the builder system has been compromised, and instead of building the software as programmed, it will build a special release containing arbitrary code and dependencies. Also it will be modified to only ask QED non-modified dependencies.

In this case, the help from QED will be limited as it only can work if someone ask for the proofs, and if those proofs contain the appropriate metadata to discover the tampering.

We can leverage the gossip agents platform included in QED to build a special proxy to detect such behaviour. This proxy server will be in charge of outgoing HTTP connections to the internet and will check against QED all the package urls before being downloaded.

#####  Tampering the QED log

The QED server log stores its cryptographic information in a local database that is replicated against the other nodes of the QED cluster using the RAFT protocol.

In our development version of QED we have included a special endpoint to modify the database without using the QED API or stopping the server, which would be the worst scenario possible in case of an attack agains QED.

In QED we can modify either the hyper or the history tree. The hyper tree is a sparse merkle tree with some optimizations that maps a given piece of data with is position in the history tree. This history tree is an always growing complete binary tree and stores the entries ordered by its insertion time.

The hyper tree will tell us if an event has been inserted or not and will give us a proof which we can use to verify it. The history tree will tell us if the event is consistent with the stored order of another entry, and will give us a also a verifiable proof.

Given this, and using as a starting point the rocksdb example, our history tree contains:

    v5.13.3  v5.12.5  v5.13.4  v5.14.2  v5.14.3  v5.15.10 v5.16.6  v5.17.2  v5.18.3    
    .________.________.________.________.________.________.________.________.


This fork is possible if we insert three new events into QED, into the same positions of old events, and generate new snapshots for them.

It will not work modifying the hyper tree only, as the history will contain the correct hash and the membership query will not validate. Also, we can't modify just the history leaf in its storage because all the intermediate nodes will not be updated, and the generated audit paths will be unverfiable.

A way to rewrite the QED is to insert again new events reseting the ballon version to the one we want to modify. In a RAFT cluster, the hijacked node needs to preserve the raft log order to be able to tamper the data without corrupting the RAFT server making it crash. We also need to modify all the events from the given point in history to the end to have the same history size than before.

In this case we are going to insert the new compromised events in the same position as the old ones:
 * event v5.16.6 becomes v5.16.6'
 * event v5.17.2 becomes v5.17.2'
 * event v5.18.3 becomes v5.18.3'


We insert into QED three new events, v5.16.6' v5.17.2' v5.18.3', but instead of using the regfular API, we are going to suppose we have a custom QED implementation that inserts in the hyper tree the version we want.

    $ ./resetVersion.sh 6
    $ ./append.sh rocksdb/v5.16.6p.json
    $ ./append.sh rocksdb/v5.17.2p.json
    $ ./append.sh rocksdb/v5.18.3p.json
    

Our forked history tree now looks like:

    v5.13.3  v5.12.5  v5.13.4  v5.14.2  v5.14.3  v5.15.10 v5.16.6  v5.17.2  v5.18.3 
    .________.________.________.________.________.________.________.________.

                                                          |
                                                          .________.________.
                                                          v5.16.6' v5.17.2' v5.18.3' 

if we use the versions stored in the hyper tree.

In this situation we will download the compromised dependency, v5.16.6' and ask QED for it, like we did before:

    $ ./membership.sh 6 rocksdb/v5.16.6p.json

And it will verify it.

To detect that we can user the snapshot store because it will have two entries for each tampered event on QED


#####  Tampering the QED snapshot store

