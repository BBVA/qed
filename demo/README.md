# Usage

## Start service
```
./start_qed.sh
```

## Submit authorized list of dependencies for a project into QED.
```
./submit.sh
```
* Steps:
- [x] Download project
- [x] Analyze dependencies and run security check
- [x] Approves the dependencies
- [x] Generates QED event with the project metadata applying salt to the event msg


```
Event format:

{
        "commit_hash": "$commit.hash",
        "src_hash": "$source_code.hash (all source code files sha256sum)"
}

ex:

{
        "commit_hash": "b75d67cd51eb53c3c3a2fc406524c940021ffbda",
        "src_hash": "1ef7e1ef1f6656899e1b602921abfffef10362f9546cea7489f6c855724a0b8f"
}
```

## Execute build pipeline
```
./build.sh
```
* Steps:
- [x] Download project
- [x] Check dependencies list (go.mod) against QED (if check fails the build is canceled)
- [x] Get event metadata from the snapshot store
- [x] Verify event against QED (if check fails the build is canceled)
- [x] Test and build code
- [x] Generate a intermediate event for seak of the Incremental Proof 
- [x] Generate a QED event with the metadata of previously generated artifact
- [x] Upload the artifact to the artifact repository

## Execute deploy stage
```
./deploy.sh
```
* Steps:
- [x] Download the artifact
- [x] Check if artifact metadata matches an existing QED event (if check fails the deployment is canceled)
- [x] Get event metadata from the snapshot store
- [x] Verify artifact event against QED (if check fails the deployment is canceled)
- [x] Get required metadata from snapshot store to execute an Incremental Proof (start and end version)
- [x] Generate the Incremental Proof
- [x] Deploy the artifact
