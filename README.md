# Verifiable data

Forensics investigations can be flawed for many causes, such as that they can lack any real evidence of an incident. For that reasons, we have the demand for an immutable tamper-evident log of everything that happens in the Ether platform. 

The purpose of this research is to find a working technology to accomplish this objective. Or, if it is not currently available, design a suitable and efficient prototype capable of fulfilling the following requirements:

 * To allow the massive ingestion of heterogeneous logs or data
 * To have the capability to index data by different fields.
 * To enable an efficient and painless verification process.
 * To allow for a periodic check or snapshot to guarantee immutability against third-party agents or audit processes.

The initial study can be found here:

    https://docs.google.com/document/d/13acAtpW7PG0PVo5_as5g5mdpMsPZ6UNxhp_raMvgjMQ/edit#

# About this project

This project is an implementation of a system that can be used to verify large amounts of data for:

 * Prove inclusion of value
 * Prove non-inclusion of value
 * Retrieve provable value for key
 * Retrieve provable current value for key
 * Prove append-only
 * Enumerate all entries
 * Prove correct operation
 * Enable detection of split-view
 
 ## Environment
 
 We use the [Go](https://golang.org) programming language and environment as described in their  [documentation](https://golang.org/doc/code.html)
 
 
 ## Testing http api
 
 Document [here](http://blog.questionable.services/article/testing-http-handlers-go/)
 
 
 ## Guide
 
     $ godoc -http=:6060
     
     go test verifiabledata/util
     go test verifiabledata/store
     go test verifiabledata/tree
     go test verifiabledata/tree/history
     go test verifiabledata/store/memory
 
 