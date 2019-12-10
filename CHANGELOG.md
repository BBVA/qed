## UNRELEASED

## 1.0.0-rc2 (December 10, 2019)

FEATURES

* server: Support TLS for Raft RPC and GRPC connections [[GH-180](https://github.com/BBVA/qed/pull/180)].

IMPROVEMENTS

* common: Bump Go version to 1.13.1 [[GH-179](https://github.com/BBVA/qed/pull/179)].
* server: Redirect from non-leader servers [[GH-176](https://github.com/BBVA/qed/pull/176)].
* server: Remove Raft's term checks when applying a new command [[GH-172](https://github.com/BBVA/qed/pull/172)].
* server: Upgrade Rocksdb to version v6.2.4 [[GH-162](https://github.com/BBVA/qed/pull/162)].
* server: Validate recovery snapshot before sending data to follower [[GH-171](https://github.com/BBVA/qed/pull/171)].
* server: Export Raft internal metrics [[GH-163](https://github.com/BBVA/qed/pull/163)].
* common: Integrate new logging system [[GH-164](https://github.com/BBVA/qed/pull/164)].
* deploy: Support Terraform v0.12.9 [[GH-167](https://github.com/BBVA/qed/pull/167)].
* common: Extend API to allow bulk insertions [[GH-107](https://github.com/BBVA/qed/pull/107)].

BUG FIXES

* server: Remove broken API-Key feature [[GH-177](https://github.com/BBVA/qed/pull/177)].
* server: Lock balloon operations to avoid race conditions [[GH-161](https://github.com/BBVA/qed/pull/161)].
* server: Improve cluster join to manage new servers with duplicat ID or address [[GH-165](https://github.com/BBVA/qed/pull/165)].
* tests: Fix timeouts problem with cluster and e2e tests [[GH-170](https://github.com/BBVA/qed/pull/170)].

## 1.0.0-rc1 (August 9, 2019)

FEATURES

* server: Enhanced cluster recovery with full backups and on-demand snapshots via new GRPC internal API [[GH-148](https://github.com/BBVA/qed/pull/148)].
* cli: Add version command to print build info [[GH-150](https://github.com/BBVA/qed/pull/150)].
* cli: Add *workload* command to stress QED servers [[GH-144](https://github.com/BBVA/qed/pull/144)].
* cli: Add *bug* command to open bug issues in Github [[GH-141](https://github.com/BBVA/qed/pull/141)].
* cli: Add *generate* command to create ed25519 keypairs [[GH-126](https://github.com/BBVA/qed/pull/126)].
* client: Add auto-verify option to membership and incremental queries [[GH-122](https://github.com/BBVA/qed/pull/122)].
* server: Support BLAKE2b hashing algorithm [[GH-123](https://github.com/BBVA/qed/pull/123)].

IMPROVEMENTS

* server: Avoid pruning unnecessary branches in history tree [[GH-149](https://github.com/BBVA/qed/pull/149)].
* cli: Improve *generate* command to create self-signed TLS cerfificates [[GH-138](https://github.com/BBVA/qed/pull/138)].
* cli: Validate correctness for address and URL parameters [[GH-137](https://github.com/BBVA/qed/pull/137)].
* server: Move event hashing out of balloon internals [[GH-136](https://github.com/BBVA/qed/pull/136)].
* server: Clean up membership API [[GH-130](https://github.com/BBVA/qed/pull/130)].
* workload: Add workload metrics [[GH-128](https://github.com/BBVA/qed/pull/128)].
* server: New cache recovery strategy [[GH-124](https://github.com/BBVA/qed/pull/124)].

BUG FIXES

* server: Replace hyper cache implementation to avoid losing events due to cache evictions [[GH-156](https://github.com/BBVA/qed/pull/156)].
* workload: Keep inserting data when QED leader changes [[GH-135](https://github.com/BBVA/qed/pull/135)].

## 0.2-alpha (April 22, 2019)

FEATURES

* agents: New agents platform [[GH-104](https://github.com/BBVA/qed/pull/104)].
* server: Expose Raft WAL metrics [[GH-102](https://github.com/BBVA/qed/pull/102)].
* server: Use RocksDB column families [[GH-87](https://github.com/BBVA/qed/pull/87)].
* server: Replace RocksDB as default storage engine [[GH-80](https://github.com/BBVA/qed/pull/80)].
* server: Use RocksDB as Raft log store [[GH-84](https://github.com/BBVA/qed/pull/84)].
* client: New advanced client with retries, healthcheks and topology discovery [[GH-82](https://github.com/BBVA/qed/pull/82)].
* server: Expose RocksDB metrics [[GH-85](https://github.com/BBVA/qed/pull/85)].
* server: New hyper tree implementation based on batches [[GH-71](https://github.com/BBVA/qed/pull/71)].
* server: Enable TLS in client API [[GH-66](https://github.com/BBVA/qed/pull/66)].

IMPROVEMENTS

* workload: Full refactor [[GH-81](https://github.com/BBVA/qed/pull/81)].
* server: Remove index table [[GH-97](https://github.com/BBVA/qed/pull/97)].
