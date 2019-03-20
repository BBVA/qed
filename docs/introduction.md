## Introduction

QED is an implementation of a forward-secure append-only persistent authenticated data structure. Each append operation emits a cryptographic structure (a signed snapshot) used to verify:
 *  if the data was appended or not
 * whether the appended data is consistent, in insertion order, to another entry

The snapshot, the piece of data inserted and a QED proof are needed to verify both statements. The snapshots are emited proactively by QED and must be stored elsewhere. The proofs are not verified or emited by QED proactively, its the user responsibility to know when and how the data must be verified depending on their needs.

The data is not stored in QED, only a representation of it produced by a collision-resistan hash function. A system to send the data or a representation of if to QED is also not provided by QED beyond a basic HTTP API.

The semantic of the appended data and the relation between each item appended is also a client responsibility.

In this document we are going to describe some use cases for the technology along with its limitations.

## Why 

The functionality described can be achieved using other tecnologies such as digitally signed data in a database, or other block chain's related structures. The advantages of the data structure implemented in QED are:
 * scalability to thousands of millions of entries
 * proof of membership or non-membership generation in logarithmic time
 * proofs of logarithmic size


## Use cases



### Artifactory

An artifactory is a repository of objects produced as a result of software construction and packaging. These artifacts are used, among other uses and users, by other software developers as a dependency of the software they are buildding.

This leads to the problem of whether a given artifact has been altered by a possible malicious third party. This problem is not only an integrity issue, as there are multiple versions of the same artifact used at the same time, the relation between these different versions is also important.

Modern development workflows include complex pipelines of highly automated processes which builds software dozens of times per day.

We can leverage QED to verify the history of an artifact ensuring our dependencies are correct automatically, failing the construction in the pipeline in case one of the dependencies have been modified.

#### workflow example

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
 
##### The first artifact
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

##### Lifecycle of a single dependency

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

Als we have 9 snapshots in a third party public store, so we can get anytime the snapshot of each version.

In this example we suppose we are using the version v5.16.6. In our pipeline we build our software depending on that version of rocksdb, but out DNS server gots corrupted and a fake github is presented to our build server, so it downloads a corrupted version of rocksdb containing malware.

Ir our pipeline, every time we download a dependency, we generate its QED entry, but this time, because the package was altered by a third party, we generate a different version of the entry:

    rocksdb/corrupted-v5.16.6.json
    
Before using it to build our software we ask QED for a membership proof, so we can verify that own download was already verified and inserted into the QED:

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

In this scenario we have authenticated an artifactory (a github release repository) which is the source of the events, then we have inserted into QED the information related to a set of  verified releases of rocksdb. Later, the artifactory was the target of an attack and a modified version of a dependency was downloaded automatically by the software construction pipeline. Our dependency check phase checked the information agains QED and discovered a tamper in the artifactory.

##### Tampering QED itself

Using the last scenario as a starting point, we have the situation that an administrator on QED inserted an old version of rocksdb again into QED, with the same version, but with other package digest, so a corrupted package can be downloaded and verified.