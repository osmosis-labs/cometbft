- `[consensus]` Make the consensus reactor no longer have packets on receive take the consensus lock.
The reactor's view is now updated via synchronous events produced by consensus state upon the relevant updates.
  ([\#3211](https://github.com/cometbft/cometbft/pull/3211))
