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
    
    Adding key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.13.3",
    	"pkg_digest": "2a37af01c9a7a0fda64ccb57aa350d0dfd495beaba6649977cc4c8f31ba2e8db"
    } ]
    
    Received snapshot with values:
    
     EventDigest: e11e5995de7d52fa22d6111592fb6ab4377702bdb8265e72655f0a02f2803c6a
     HyperDigest: 0d9c4318fd8c0c9c79b8b2d594d77245d9a00fa8807bfdf8e1a37b0a7c00c64f
     HistoryDigest: d7655612204fe6b7d12522b83b7588c4ec6d9314e620a94e712013a39daea707
     Version: 0
    
    
    Adding key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.12.5",
    	"pkg_digest": "adc6660befe44bd4030a3bfadaccbb7e6a41a1c2f7526849686c80dc7a5c5a5d"
    } ]
    
    Received snapshot with values:
    
     EventDigest: 431c542196c7a6992a779930349cc25730074f2742144867185f651dcdcbb11b
     HyperDigest: 8da93187d3ea448c3ebb3054e08d4039f877d55179a6cc1b8ac9c3461188518a
     HistoryDigest: a07836a792ddc9b2f2010e994f788df82ed13529f5dfa786808a614ece447c63
     Version: 1
    
    
    Adding key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.13.4",
    	"pkg_digest": "d4df72df4faf9fddf942ba1eb0946138c218a572301b8b7f604754189ea16ce5"
    } ]
    
    Received snapshot with values:
    
     EventDigest: 4c19f3212b9d2b5ca2439ece3d6b3740354102025dbed159631c70c2af9240bc
     HyperDigest: 8189a13e4d7bd4987dca88a6a2b9efd10b294f17f48be6183f3e28a8ba75f315
     HistoryDigest: 7d2e6c4306af2be0fcddfbc893aadbdd22f0bf4df4a3409606b70422cca6e6ed
     Version: 2
    
    
    Adding key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.14.2",
    	"pkg_digest": "15bb12b9492fc2a20c0c4dbd8873703a1b6b620c32708aaaf558b0a4f58feeda"
    } ]
    
    Received snapshot with values:
    
     EventDigest: c4edaeafda5a0540d49f9f2e1d28e87fe881c7772c33058ae7bd6b11c96826c4
     HyperDigest: 141f1c5a50140940c7253dcd59a0ffb37c95e2a411cca3a772e9fce6da496f6a
     HistoryDigest: e4c182324960bec2a2ada1c9a06a3f429e94e44be852d8ead4e0352f77a91920
     Version: 3
    
    
    Adding key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.14.3",
    	"pkg_digest": "e927efa48b01100bfe7aa43cd0f18c1a3c37afdcdf7337d89cd9ab7541d4f07a"
    } ]
    
    Received snapshot with values:
    
     EventDigest: 04962bc63cbee1fbdcd91ed385f7aba15b9b654a384797edd82acb43da066b7e
     HyperDigest: 830fc0a7e6ab78d84327ba8dd493f8a1be4f7b8ccb0bedd708ba6538cce4ed66
     HistoryDigest: 961a857ad441a2d3d9326264383ef1e3b777d0aea0b9b5cc1425814ebd933e92
     Version: 4
    
    
    Adding key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.15.10",
    	"pkg_digest": "16356771775376b50e5cd4e7a185e84f398493183d375ff14cd6d396cdae6ea0"
    } ]
    
    Received snapshot with values:
    
     EventDigest: 2714b2d8a1eb6fb888ff6deb428861bebf8b05feba10d121e90567e427fdcae3
     HyperDigest: 58961f73238e55b50739ab5786351a8580a46af3ee4e07f534a2537c600962ff
     HistoryDigest: bdb080bd3d92f7363306cfcce22b0e8d6f4b4a6b4ff7cda94f19f9b5009a04d2
     Version: 5
    
    
    Adding key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.16.6",
    	"pkg_digest": "34c6575304c905418d85aa6b6e84b2854286bdc6083c0e6e2df756d0cf74663f"
    } ]
    
    Received snapshot with values:
    
     EventDigest: f7370d22bb6172aa1880c422133dc6c682a6fe7dc020676d2b762de4ea2799b6
     HyperDigest: 8300b4a08d023c5569858d71593560742758e2e5878ff201db134ef7fb8a36a8
     HistoryDigest: fa50ad9f6644ac21fecd6ebd3648e31c5521f5e0bba1350f65e00cdccfbb866f
     Version: 6
    
    
    Adding key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.17.2",
    	"pkg_digest": "269c266c1fc12d1e73682ed1a05296588e8482d188e6d56408a29de447ce87d7"
    } ]
    
    Received snapshot with values:
    
     EventDigest: 1594149fd2713cd25c2d8cf742cf94a3340652025b8f8b6b31c30ef739f2d566
     HyperDigest: 2d724a6262ed4b9339f6a11c2ff797407399b3b79cd8695e54994fca846caaaf
     HistoryDigest: 85c91d11b7034cae1e589ce1684ec17a8df3d88af14d3b8dbe8653aa337de699
     Version: 7
    
    
    Adding key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.18.3",
    	"pkg_digest": "4d86973cd9f034b95f7b03fe513feee5cf089ebbe34d4a83f59fdbb7c59f3ae3"
    } ]

    Received snapshot with values:
    
     EventDigest: 41eeff6cb10f1bd92c6141be0cc5e6313a28e97114ec58ca4cc5a326e6c6e55f
     HyperDigest: 33e90cdcc6eff1cecbde0ae5b0d4edbfdbaf64b2bccc470c7ee4ba935982c454
     HistoryDigest: 113b14afa62f8187f37a9e3692c6edfe3c70437213a9c4a0e3a51faf69be960b
     Version: 8

Now we have a QED with 9 rocksdb versions ordered by release date from older to newer.

Also we have 9 snapshots in a third party public store, so we can get anytime the snapshot of each version.

In this example we suppose we are using the version v5.16.6. In our pipeline we build our software depending on that version of rocksdb. The software distributor or repository has been compromised, a new package has been uploaded to the official repository, with an old version but new contents and new download hashes.

Ir our pipeline, every time we download a dependency, we generate its QED entry, but this time, because the package was altered by a third party, we generate a different version of the entry:

    rocksdb/corrupted-v5.16.6.json
    
Before using it to build our software we ask QED for a membership proof, so we can verify that our download was already verified and inserted into the QED:

    $ ./membership.sh 6 rocksdb/corrupted-v5.16.6.json

To verify the event we need the correct snapshot information. Because we're using the interactive client, it ask us the hyperDigest for the current version of the QED. We go to the snapshot store to get it:

    $ ./getsnapshot.sh 8
    HyperDigest: 33e90cdcc6eff1cecbde0ae5b0d4edbfdbaf64b2bccc470c7ee4ba935982c454
    HistoryDigest: 113b14afa62f8187f37a9e3692c6edfe3c70437213a9c4a0e3a51faf69be960b
    
We use the hyperdigest presented here, and the client tries to verify the information:

    $ ./membership.sh 6 rocksdb/corrupted-v5.16.6.json
    
    Querying key [ {
    	"pkg_name": "rocksdb",
    	"pkg_version": "v5.16.6",
    	"pkg_digest": "34c6575304c905418d85aa6b6e84b2854286bdc6083c0e6e2df756d0cf74663l"
    } ] with version [ 6 ]
    
    Received membership proof:
    
     Exists: false
     Hyper audit path: <TRUNCATED>
     History audit path: <TRUNCATED>
     CurrentVersion: 8
     QueryVersion: 6
     ActualVersion: 6
     KeyDigest: e217828ce68032cc4ab2840e5303311aad19f0d139f21f440ca455d6691f8f65
    
    Please, provide the hyperDigest for current version [ 8 ]: 33e90cdcc6eff1cecbde0ae5b0d4edbfdbaf64b2bccc470c7ee4ba935982c454
    
    Verifying with Snapshot: 
    
     EventDigest:e217828ce68032cc4ab2840e5303311aad19f0d139f21f440ca455d6691f8f65
     HyperDigest: 33e90cdcc6eff1cecbde0ae5b0d4edbfdbaf64b2bccc470c7ee4ba935982c454
     HistoryDigest: 
     Version: 6
    
    Verify: KO
    
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



This fork is possible if we insert three new events into QED, and then tamper the hyper tree using the following values:

 * event v5.16.6 and v5.16.6', both,  point to history tree version 6
 * event v5.17.2 and v5.17.2' to version 7
 * event v5.18.3 and v5.18.3' to version 8

We insert into QED three new events, v5.16.6' v5.17.2' v5.18.3':

    $ for i in v5.16.6p v5.17.2p v5.18.3p; do ./append.sh rocksdb/$i.json; done
    Adding key [ {
            "pkg_name": "rocksdb",
            "pkg_version": "v5.16.6",
            "pkg_digest": "aaaa5304c905418d85aa6b6e84b2854286bdc6083c0e6e2df756d0cf74663f"
    } ]
    
    Received snapshot with values:
    
     EventDigest: 71f7fc4137f752b9128f57b903b0216ff949dd3b90a85c1918ace4f7608c7a7c
     HyperDigest: 6e5bf0d4b9351463ba600cb70e042ad0e4cbb5d0d84b8f54b5cc098eab631b4d
     HistoryDigest: 460572d1444d21135b83c36d0e4defd0f66b568fbd1359a2de2a0ec358e64bf5
     Version: 9
    
    
    Adding key [ {
            "pkg_name": "rocksdb",
            "pkg_version": "v5.17.2",
            "pkg_digest": "aaaa266c1fc12d1e73682ed1a05296588e8482d188e6d56408a29de447ce87d7"
    } ]
    
    Received snapshot with values:
    
     EventDigest: dae174b4f11ae7fca95bdc1f95ab9b002ecb97f36f54a7b9cecc0b0a0c597894
     HyperDigest: bdc713608fe0a158e62e66865e9e3d3098ae1f4e698f9b01d6047035e18d7673
     HistoryDigest: c2f309be529381a12fcb41168676ef0f1c8511174e1c96aaae29cd97e94295fb
     Version: 10
    
    
    Adding key [ {
            "pkg_name": "rocksdb",
            "pkg_version": "v5.18.3",
            "pkg_digest": "aaaa973cd9f034b95f7b03fe513feee5cf089ebbe34d4a83f59fdbb7c59f3ae3"
    } ]
    
    Received snapshot with values:
    
     EventDigest: b460c0fba33f7d14f1d2f454dc5f381e82dc6eb1d5d77dacc0fd90f6c3076c14
     HyperDigest: 16e7c19c00998d1478f37612544a9ca7b27ea6d4087883e8e4d45f28a027d70f
     HistoryDigest: 2df4d8ac43f559a5954cab3a095cd7be9e17dd191602fe166fabc04956f6dc69
     Version: 11

After this, our history tree contains:

    v5.13.3   ... v5.15.10 v5.16.6  v5.17.2  v5.18.3  v5.16.6' v5.17.2' v5.18.3'   
    .________ ... .________.________.________.________.________.________.


And the events v5.16.6' v5.17.2' v5.18.3' in hyper tree points to versions 9, 10 and 11 respectively.

Also, every time an event is inserted into QED, a new snapshot is generated and publised, so or snapshot store contains also this three extra snapshots.

Now we tamper the last three events to point to the prior history versions:

	$ ./tamperhyper.sh 6e5bf0d4b9351463ba600cb70e042ad0e4cbb5d0d84b8f54b5cc098eab631b4d 6
	$ ./tamperhyper.sh bdc713608fe0a158e62e66865e9e3d3098ae1f4e698f9b01d6047035e18d7673 7
    $ ./tamperhyper.sh 16e7c19c00998d1478f37612544a9ca7b27ea6d4087883e8e4d45f28a027d70f 8

Our forked history tree now looks like:

    v5.13.3  v5.12.5  v5.13.4  v5.14.2  v5.14.3  v5.15.10 v5.16.6  v5.17.2  v5.18.3    real
    .________.________.________.________.________.________.________.________.

                                                          |
                                                          .________.________.
                                                          v5.16.6' v5.17.2' v5.18.3'    fork

if we use the versions stored in the hyper tree.

In this situation we will download the compromised dependency, v5.16.6' and ask QED for it, like we did before:

    $ ./membership.sh 6 rocksdb/v5.16.6p.json

    
 ./membership.sh 6 rocksdb/v5.16.6p.json
    
    Querying key [ key6 ] with version [ 6 ]
    
    Received membership proof:
    
     Exists: false
     Hyper audit path: <TRUNCATED>
     History audit path: <TRUNCATED>
     CurrentVersion: 11
     QueryVersion: 6
     ActualVersion: 6
     KeyDigest: 8dfad052fee5c62957d3ebe1752219a02f45634b2c32a6ac408b26ffcedfb7da
    
    Please, provide the hyperDigest for current version [ 11 ]: 16e7c19c00998d1478f37612544a9ca7b27ea6d4087883e8e4d45f28a027d70f
    
    Verifying with Snapshot: 
    
     EventDigest:8dfad052fee5c62957d3ebe1752219a02f45634b2c32a6ac408b26ffcedfb7da
     HyperDigest: 16e7c19c00998d1478f37612544a9ca7b27ea6d4087883e8e4d45f28a027d70f
     HistoryDigest: 
     Version: 6
    
    Verify: KO

    

The first we notice is that the entry for the version v5.16.6 is in the version 6, when we inserted it in the version 9. This means the tampering was succesfull.

But, because we use the version as part of the digest process, the root digest of the tree is different, as the original had a version of 9, and the verification fails.

This means simple tampering in the database will not work. In order to tamper we need to build a special version of QED which will do a valid insert, and will publish an snapshot, but inserting a custom version in the hyper tree instead of the one corresponding to the history version.


#####  Tampering the QED snapshot store