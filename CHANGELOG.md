# CHANGELOG

## v0.37.4-v25-osmo-6

* [#83](https://github.com/osmosis-labs/cometbft/pull/83)  perf(types): 3x speedup MakePartSet (#3117)
* [#85](https://github.com/osmosis-labs/cometbft/pull/85)  perf(flowrate): Speedup flowrate.Clock (#3016)
* [#86](https://github.com/osmosis-labs/cometbft/pull/86)  Comment out expensive debug logs
* [#91](https://github.com/osmosis-labs/cometbft/pull/91)  perf(consensus): Minor improvement by making add vote only do one peer set mutex call, not 3 (#3156)
* [#93](https://github.com/osmosis-labs/cometbft/pull/93)  perf(consensus): Make some consensus reactor messages take RLock's not WLock's (#3159)
* [#95](https://github.com/osmosis-labs/cometbft/pull/95)  perf(types) Make a new method `GetByAddressMut` for `ValSet`, which does not copy the returned validator. (#3129)
* [#97](https://github.com/osmosis-labs/cometbft/pull/97)   perf(p2p/connection): Lower wasted re-allocations in sendRoutine (#2986)
* [#99](https://github.com/osmosis-labs/cometbft/pull/99)   perf(consensus): Reuse an internal buffer for block building (#3162)
* [#97](https://github.com/osmosis-labs/cometbft/pull/101)  perf(consensus): Run broadcast routines out of process (speeds up consensus mutex) #3180


## v0.37.4-v25-osmo-5

* [#73](https://github.com/osmosis-labs/cometbft/pull/73) perf(consensus/blockexec): Add simplistic block validation cache
* [#74](https://github.com/osmosis-labs/cometbft/pull/74) perf(consensus): Minor speedup to mark late vote metrics
* [#75](https://github.com/osmosis-labs/cometbft/pull/75) perf(p2p): 4% speedup to readMsg by removing one allocation
* [#76](https://github.com/osmosis-labs/cometbft/pull/76) perf(consensus): Add LRU caches for blockstore operations used in gossip
* [#77](https://github.com/osmosis-labs/cometbft/pull/77) perf(consensus): Make every gossip thread use its own randomness instance, reducing mutex contention 

## v0.37.4-v25-osmo-4

* [#69](https://github.com/osmosis-labs/cometbft/pull/69) perf: Make mempool update async from block.Commit (#3008)
* [#67](https://github.com/osmosis-labs/cometbft/pull/67) fix: TimeoutTicker returns wrong value/timeout pair when timeouts are…

## v0.37.4-v25-osmo-3

* [#61](https://github.com/osmosis-labs/cometbft/pull/61) refactor(p2p/connection): Slight refactor to sendManyPackets that helps highlight performance improvements (backport #2953) (#2978)
* [#62](https://github.com/osmosis-labs/cometbft/pull/62) perf(consensus/blockstore): Remove validate basic call from LoadBlock
* [#59](https://github.com/osmosis-labs/cometbft/pull/59) `[blockstore]` Remove a redundant `Header.ValidateBasic` call in `LoadBlockMeta`, 75% reducing this time. ([\#2964](https://github.com/cometbft/cometbft/pull/2964))
* [#59](https://github.com/osmosis-labs/cometbft/pull/59) `[p2p/conn]` Speedup connection.WritePacketMsgTo, by reusing internal buffers rather than re-allocating. ([\#2986](https://github.com/cometbft/cometbft/pull/2986))


## v0.37.4-v25-osmo-2 & v0.37.4-v24-osmo-5

* [#35](https://github.com/osmosis-labs/cometbft/pull/35) Handle last element in PickRandom
* [#38](https://github.com/osmosis-labs/cometbft/pull/38) Remove expensive Logger debug call in PublishEventTx
* [#39](https://github.com/osmosis-labs/cometbft/pull/39) Change finalizeCommit to use applyVerifiedBlock
* [#40](https://github.com/osmosis-labs/cometbft/pull/40) Speedup NewDelimitedWriter
* [#41](https://github.com/osmosis-labs/cometbft/pull/41) Remove unnecessary atomic read
* [#42](https://github.com/osmosis-labs/cometbft/pull/42) Remove a minint call that was appearing in write packet delays
* [#43](https://github.com/osmosis-labs/cometbft/pull/43) Speedup extended commit.BitArray()

## v0.37.4-v24-osmo-4

* [#27](https://github.com/osmosis-labs/cometbft/pull/27) Lower allocation overhead of txIndex matchRange
* [#28](https://github.com/osmosis-labs/cometbft/pull/28) Significantly speedup bitArray.PickRandom
* [#29](https://github.com/osmosis-labs/cometbft/pull/29) Lower heap overhead of JSON encoding
* [#30](https://github.com/osmosis-labs/cometbft/pull/30) Cache the block hash

## v0.37.4-v24-osmo-3

* [#21](https://github.com/osmosis-labs/cometbft/pull/21) Move websocket logs to Debug
* [#22](https://github.com/osmosis-labs/cometbft/pull/22) Fix txSearch pagination performance issues
* [#25](https://github.com/osmosis-labs/cometbft/pull/25) Optimize merkle tree hashing

## v0.37.4-v24-osmo-2

* [#20](https://github.com/osmosis-labs/cometbft/pull/20) Fix the rollback command

## v0.37.4-v24-osmo-1

* [#5](https://github.com/osmosis-labs/cometbft/pull/5) Batch verification
* [#11](https://github.com/osmosis-labs/cometbft/pull/11) Skip verification of commit sigs
* [#13](https://github.com/osmosis-labs/cometbft/pull/13) Avoid double-saving ABCI responses

## v0.37.4-v23-osmo-3

* [#17](https://github.com/osmosis-labs/cometbft/pull/17) Set the max number of (concurrently) downloaded blocks to {peersCount * 20}
* [#18](https://github.com/osmosis-labs/cometbft/pull/18) Sort peers by download rate & multiple requests for closer blocks

## v0.37.4-v23-osmo-2

* [#3](https://github.com/osmosis-labs/cometbft/pull/3) Avoid double-calling types.BlockFromProto
* [#4](https://github.com/osmosis-labs/cometbft/pull/4) Do not validatorBlock twice

## v0.37.4-v23-osmo-1

## v0.37.4

*November 27, 2023*

This release provides the **nop** mempool for applications that want to build
their own mempool. Using this mempool effectively disables all mempool
functionality in CometBFT, including transaction dissemination and the
`broadcast_tx_*` endpoints.

Also fixes a small bug in the mempool for an experimental feature, and reverts
the change from v0.37.3 that bumped the minimum Go version to v1.21.

### BUG FIXES

- `[mempool]` Avoid infinite wait in transaction sending routine when
  using experimental parameters to limiting transaction gossiping to peers
  ([\#1654](https://github.com/cometbft/cometbft/pull/1654))

### FEATURES

- `[mempool]` Add `nop` mempool ([\#1643](https://github.com/cometbft/cometbft/pull/1643))

  If you want to use it, change mempool's `type` to `nop`:

  ```toml
  [mempool]

  # The type of mempool for this node to use.
  #
  # Possible types:
  # - "flood" : concurrent linked list mempool with flooding gossip protocol
  # (default)
  # - "nop"   : nop-mempool (short for no operation; the ABCI app is responsible
  # for storing, disseminating and proposing txs). "create_empty_blocks=false"
  # is not supported.
  type = "nop"
  ```

## v0.37.3

*November 17, 2023*

This release contains, among other things, an opt-in, experimental feature to
help reduce the bandwidth consumption associated with the mempool's transaction
gossip.

### BREAKING CHANGES

- `[p2p]` Remove unused UPnP functionality
  ([\#1113](https://github.com/cometbft/cometbft/issues/1113))

### BUG FIXES

- `[state/indexer]` Respect both height params while querying for events
   ([\#1529](https://github.com/cometbft/cometbft/pull/1529))

### FEATURES

- `[node/state]` Add Go API to bootstrap block store and state store to a height
  ([\#1057](https://github.com/tendermint/tendermint/pull/#1057)) (@yihuang)
- `[metrics]` Add metric for mempool size in bytes `SizeBytes`.
  ([\#1512](https://github.com/cometbft/cometbft/pull/1512))

### IMPROVEMENTS

- `[crypto/sr25519]` Upgrade to go-schnorrkel@v1.0.0 ([\#475](https://github.com/cometbft/cometbft/issues/475))
- `[node]` Make handshake cancelable ([cometbft/cometbft\#857](https://github.com/cometbft/cometbft/pull/857))
- `[node]` Close evidence.db OnStop ([cometbft/cometbft\#1210](https://github.com/cometbft/cometbft/pull/1210): @chillyvee)
- `[mempool]` Add experimental feature to limit the number of persistent peers and non-persistent
  peers to which the node gossip transactions (only for "v0" mempool).
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558))
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))
- `[config]` Add mempool parameters `experimental_max_gossip_connections_to_persistent_peers` and
  `experimental_max_gossip_connections_to_non_persistent_peers` for limiting the number of peers to
  which the node gossip transactions. 
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558))
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))

## v0.37.2

*June 14, 2023*

Provides several minor bug fixes, as well as fixes for several low-severity
security issues.

### BUG FIXES

- `[pubsub]` Pubsub queries are now able to parse big integers (larger than
  int64). Very big floats are also properly parsed into very big integers
  instead of being truncated to int64.
  ([\#771](https://github.com/cometbft/cometbft/pull/771))
- `[state/kvindex]` Querying event attributes that are bigger than int64 is now
  enabled. We are not supporting reading floats from the db into the indexer
  nor parsing them into BigFloats to not introduce breaking changes in minor
  releases. ([\#771](https://github.com/cometbft/cometbft/pull/771))

### IMPROVEMENTS

- `[rpc]` Remove response data from response failure logs in order
  to prevent large quantities of log data from being produced
  ([\#654](https://github.com/cometbft/cometbft/issues/654))

### SECURITY FIXES

- `[rpc/jsonrpc/client]` **Low severity** - Prevent RPC
  client credentials from being inadvertently dumped to logs
  ([\#787](https://github.com/cometbft/cometbft/pull/787))
- `[cmd/cometbft/commands/debug/kill]` **Low severity** - Fix unsafe int cast in
  `debug kill` command ([\#793](https://github.com/cometbft/cometbft/pull/793))
- `[consensus]` **Low severity** - Avoid recursive call after rename to
  `(*PeerState).MarshalJSON`
  ([\#863](https://github.com/cometbft/cometbft/pull/863))
- `[mempool/clist_mempool]` **Low severity** - Prevent a transaction from
  appearing twice in the mempool
  ([\#890](https://github.com/cometbft/cometbft/pull/890): @otrack)

## v0.37.1

*April 26, 2023*

This release fixes several bugs, and has had to introduce one small Go
API-breaking change in the `crypto/merkle` package in order to address what
could be a security issue for some users who directly and explicitly make use of
that code.

### BREAKING CHANGES

- `[crypto/merkle]` Do not allow verification of Merkle Proofs against empty trees (`nil` root). `Proof.ComputeRootHash` now panics when it encounters an error, but `Proof.Verify` does not panic
  ([\#558](https://github.com/cometbft/cometbft/issues/558))

### BUG FIXES

- `[consensus]` Unexpected error conditions in `ApplyBlock` are non-recoverable, so ignoring the error and carrying on is a bug. We replaced a `return` that disregarded the error by a `panic`.
  ([\#496](https://github.com/cometbft/cometbft/pull/496))
- `[consensus]` Rename `(*PeerState).ToJSON` to `MarshalJSON` to fix a logging data race
  ([\#524](https://github.com/cometbft/cometbft/pull/524))
- `[light]` Fixed an edge case where a light client would panic when attempting
  to query a node that (1) has started from a non-zero height and (2) does
  not yet have any data. The light client will now, correctly, not panic
  _and_ keep the node in its list of providers in the same way it would if
  it queried a node starting from height zero that does not yet have data
  ([\#575](https://github.com/cometbft/cometbft/issues/575))

### IMPROVEMENTS

- `[jsonrpc/client]` Improve the error message for client errors stemming from
  bad HTTP responses.
  ([cometbft/cometbft\#638](https://github.com/cometbft/cometbft/pull/638))

## v0.37.0

*March 6, 2023*

This is the first CometBFT release with ABCI 1.0, which introduces the
`PrepareProposal` and `ProcessProposal` methods, with the aim of expanding the
range of use cases that application developers can address. This is the first
change to ABCI towards ABCI++, and the full range of ABCI++ functionality will
only become available in the next major release with ABCI 2.0. See the
[specification](./spec/abci/) for more details.

In the v0.34.27 release, the CometBFT Go module is still
`github.com/tendermint/tendermint` to facilitate ease of upgrading for users,
but in this release we have changed this to `github.com/cometbft/cometbft`.

Please also see our [upgrading guidelines](./UPGRADING.md) for more details on
upgrading from the v0.34 release series.

Also see our [QA results](https://docs.cometbft.com/v0.37/qa/v037/cometbft) for
the v0.37 release.

We'd love your feedback on this release! Please reach out to us via one of our
communication channels, such as [GitHub
Discussions](https://github.com/cometbft/cometbft/discussions), with any of your
questions, comments and/or concerns.

See below for more details.

### BREAKING CHANGES

- The `TMHOME` environment variable was renamed to `CMTHOME`, and all environment variables starting with `TM_` are instead prefixed with `CMT_`
  ([\#211](https://github.com/cometbft/cometbft/issues/211))
- `[p2p]` Reactor `Send`, `TrySend` and `Receive` renamed to `SendEnvelope`,
  `TrySendEnvelope` and `ReceiveEnvelope` to allow metrics to be appended to
  messages and measure bytes sent/received.
  ([\#230](https://github.com/cometbft/cometbft/pull/230))
- Bump minimum Go version to 1.20
  ([\#385](https://github.com/cometbft/cometbft/issues/385))
- `[abci]` Make length delimiter encoding consistent
  (`uint64`) between ABCI and P2P wire-level protocols
  ([\#5783](https://github.com/tendermint/tendermint/pull/5783))
- `[abci]` Change the `key` and `value` fields from
  `[]byte` to `string` in the `EventAttribute` type.
  ([\#6403](https://github.com/tendermint/tendermint/pull/6403))
- `[abci/counter]` Delete counter example app
  ([\#6684](https://github.com/tendermint/tendermint/pull/6684))
- `[abci]` Renamed `EvidenceType` to `MisbehaviorType` and `Evidence`
  to `Misbehavior` as a more accurate label of their contents.
  ([\#8216](https://github.com/tendermint/tendermint/pull/8216))
- `[abci]` Added cli commands for `PrepareProposal` and `ProcessProposal`.
  ([\#8656](https://github.com/tendermint/tendermint/pull/8656))
- `[abci]` Added cli commands for `PrepareProposal` and `ProcessProposal`.
  ([\#8901](https://github.com/tendermint/tendermint/pull/8901))
- `[abci]` Renamed `LastCommitInfo` to `CommitInfo` in preparation for vote
  extensions. ([\#9122](https://github.com/tendermint/tendermint/pull/9122))
- Change spelling from British English to American. Rename
  `Subscription.Cancelled()` to `Subscription.Canceled()` in `libs/pubsub`
  ([\#9144](https://github.com/tendermint/tendermint/pull/9144))
- `[abci]` Removes unused Response/Request `SetOption` from ABCI
  ([\#9145](https://github.com/tendermint/tendermint/pull/9145))
- `[config]` Rename the fastsync section and the
  fast\_sync key blocksync and block\_sync respectively
  ([\#9259](https://github.com/tendermint/tendermint/pull/9259))
- `[types]` Reduce the use of protobuf types in core logic. `ConsensusParams`,
  `BlockParams`, `ValidatorParams`, `EvidenceParams`, `VersionParams` have
  become native types.  They still utilize protobuf when being sent over
  the wire or written to disk.  Moved `ValidateConsensusParams` inside
  (now native type) `ConsensusParams`, and renamed it to `ValidateBasic`.
  ([\#9287](https://github.com/tendermint/tendermint/pull/9287))
- `[abci/params]` Deduplicate `ConsensusParams` and `BlockParams` so
  only `types` proto definitions are use. Remove `TimeIotaMs` and use
  a hard-coded 1 millisecond value to ensure monotonically increasing
  block times. Rename `AppVersion` to `App` so as to not stutter.
  ([\#9287](https://github.com/tendermint/tendermint/pull/9287))
- `[abci]` New ABCI methods `PrepareProposal` and `ProcessProposal` which give
  the app control over transactions proposed and allows for verification of
  proposed blocks. ([\#9301](https://github.com/tendermint/tendermint/pull/9301))

### BUG FIXES

- `[consensus]` Fixed a busy loop that happened when sending of a block part failed by sleeping in case of error.
  ([\#4](https://github.com/informalsystems/tendermint/pull/4))
- `[state/kvindexer]` Fixed the default behaviour of the kvindexer to index and
  query attributes by events in which they occur. In 0.34.25 this was mitigated
  by a separated RPC flag. @jmalicevic
  ([\#77](https://github.com/cometbft/cometbft/pull/77))
- `[state/kvindexer]` Resolved crashes when event values contained slashes,
  introduced after adding event sequences in
  [\#77](https://github.com/cometbft/cometbft/pull/77). @jmalicevic
  ([\#382](https://github.com/cometbft/cometbft/pull/382))
- `[consensus]` ([\#386](https://github.com/cometbft/cometbft/pull/386)) Short-term fix for the case when `needProofBlock` cannot find previous block meta by defaulting to the creation of a new proof block. (@adizere)
  - Special thanks to the [Vega.xyz](https://vega.xyz/) team, and in particular to Zohar (@ze97286), for reporting the problem and working with us to get to a fix.
- `[docker]` enable cross platform build using docker buildx
  ([\#9073](https://github.com/tendermint/tendermint/pull/9073))
- `[consensus]` fix round number of `enterPropose`
  when handling `RoundStepNewRound` timeout.
  ([\#9229](https://github.com/tendermint/tendermint/pull/9229))
- `[docker]` ensure Docker image uses consistent version of Go
  ([\#9462](https://github.com/tendermint/tendermint/pull/9462))
- `[p2p]` prevent peers who have errored from being added to `peer_set`
  ([\#9500](https://github.com/tendermint/tendermint/pull/9500))
- `[blocksync]` handle the case when the sending
  queue is full: retry block request after a timeout
  ([\#9518](https://github.com/tendermint/tendermint/pull/9518))

### FEATURES

- `[abci]` New ABCI methods `PrepareProposal` and `ProcessProposal` which give
  the app control over transactions proposed and allows for verification of
  proposed blocks. ([\#9301](https://github.com/tendermint/tendermint/pull/9301))

### IMPROVEMENTS

- `[e2e]` Add functionality for uncoordinated (minor) upgrades
  ([\#56](https://github.com/tendermint/tendermint/pull/56))
- `[tools/tm-signer-harness]` Remove the folder as it is unused
  ([\#136](https://github.com/cometbft/cometbft/issues/136))
- `[p2p]` Reactor `Send`, `TrySend` and `Receive` renamed to `SendEnvelope`,
  `TrySendEnvelope` and `ReceiveEnvelope` to allow metrics to be appended to
  messages and measure bytes sent/received.
  ([\#230](https://github.com/cometbft/cometbft/pull/230))
- `[abci]` Added `AbciVersion` to `RequestInfo` allowing
  applications to check ABCI version when connecting to CometBFT.
  ([\#5706](https://github.com/tendermint/tendermint/pull/5706))
- `[cli]` add `--hard` flag to rollback command (and a boolean to the `RollbackState` method). This will rollback
   state and remove the last block. This command can be triggered multiple times. The application must also rollback
   state to the same height.
  ([\#9171](https://github.com/tendermint/tendermint/pull/9171))
- `[crypto]` Update to use btcec v2 and the latest btcutil.
  ([\#9250](https://github.com/tendermint/tendermint/pull/9250))
- `[rpc]` Added `header` and `header_by_hash` queries to the RPC client
  ([\#9276](https://github.com/tendermint/tendermint/pull/9276))
- `[proto]` Migrate from `gogo/protobuf` to `cosmos/gogoproto`
  ([\#9356](https://github.com/tendermint/tendermint/pull/9356))
- `[rpc]` Enable caching of RPC responses
  ([\#9650](https://github.com/tendermint/tendermint/pull/9650))
- `[consensus]` Save peer LastCommit correctly to achieve 50% reduction in gossiped precommits.
  ([\#9760](https://github.com/tendermint/tendermint/pull/9760))

---

CometBFT is a fork of [Tendermint Core](https://github.com/tendermint/tendermint) as of late December 2022.

## Bug bounty

Friendly reminder, we have a [bug bounty program](https://hackerone.com/cosmos).

## Previous changes

For changes released before the creation of CometBFT, please refer to the Tendermint Core [CHANGELOG.md](https://github.com/tendermint/tendermint/blob/a9feb1c023e172b542c972605311af83b777855b/CHANGELOG.md).

