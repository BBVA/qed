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
        "msg": "$salt $go.mod.content ",
        "version": "$project_version",
        "hash": "$go.mod.hash"
}

ex:

{
        "msg": "b08b6378ec5c9ac6194c35e5b02aaecd7c11d957f55e3d23537b774c9b6703c2 module github.com/gin-gonic/gin go 1.12 require ( github.com/gin-contrib/sse v0.0.0-20190301062529-5545eab6dad3 github.com/golang/protobuf v1.3.0 github.com/json-iterator/go v1.1.5 github.com/mattn/go-isatty v0.0.6 github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect github.com/modern-go/reflect2 v1.0.1 // indirect github.com/stretchr/testify v1.3.0 github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43 golang.org/x/net v0.0.0-20190301231341-16b79f2e4e95 gopkg.in/go-playground/assert.v1 v1.2.1 // indirect gopkg.in/go-playground/validator.v8 v8.18.2 gopkg.in/yaml.v2 v2.2.2 )",
        "version": "v1.3.0",
        "hash": "917d7990693a5c4ec3010512717f73087fcecf235c3ec8808f53bb11c82d785e"
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

## Generate more traffic to QED server
```
./add_event1.sh
```

## Execute release stage
```
./release.sh
```
* Steps:
- [x] Generates a QED event with the metadata of previously generated artifact
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
