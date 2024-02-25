package blocksync

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	flow "github.com/cometbft/cometbft/libs/flowrate"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
)

/*
eg, L = latency = 0.1s
	P = num peers = 10
	FN = num full nodes
	BS = 1kB block size
	CB = 1 Mbit/s = 128 kB/s
	CB/P = 12.8 kB
	B/S = CB/P/BS = 12.8 blocks/s

	12.8 * 0.1 = 1.28 blocks on conn
*/

const (
	requestIntervalMS         = 2
	maxTotalRequesters        = 1000
	maxPendingRequests        = maxTotalRequesters
	maxPendingRequestsPerPeer = 200
	requestRetrySeconds       = 30

	// Minimum recv rate to ensure we're receiving blocks from a peer fast
	// enough. If a peer is not sending us data at at least that rate, we
	// consider them to have timedout and we disconnect.
	//
	// Assuming a DSL connection (not a good choice) 128 Kbps (upload) ~ 15 KB/s,
	// sending data across atlantic ~ 7.5 KB/s.
	minRecvRate = 76800 // 75 KB/s

	// Maximum difference between current and new block's height.
	maxDiffBetweenCurrentAndReceivedBlockHeight = 1000
	maxBlocksAhead                              = 200
)

var peerTimeout = 2 * time.Second // not const so we can override with tests

/*
	Peers self report their heights when we join the block pool.
	Starting from our latest pool.height, we request blocks
	in sequence from peers that reported higher heights than ours.
	Every so often we ask peers what height they're on so we can keep going.

	Requests are continuously made for blocks of higher heights until
	the limit is reached. If most of the requests have no available peers, and we
	are not at peer limits, we can probably switch to consensus reactor
*/

// BlockPool keeps track of the block sync peers, block requests and block responses.
type BlockPool struct {
	service.BaseService
	startTime time.Time

	mtx cmtsync.Mutex
	// block requests
	requesters map[int64]*bpRequester
	height     int64 // the lowest key in requesters.
	// peers
	peers         map[p2p.ID]*bpPeer
	maxPeerHeight int64 // the biggest reported height

	// atomic
	numPending int32 // number of requests pending assignment or block response

	requestsCh chan<- BlockRequest
	errorsCh   chan<- peerError
}

// NewBlockPool returns a new BlockPool with the height equal to start. Block
// requests and errors will be sent to requestsCh and errorsCh accordingly.
func NewBlockPool(start int64, requestsCh chan<- BlockRequest, errorsCh chan<- peerError) *BlockPool {
	bp := &BlockPool{
		peers: make(map[p2p.ID]*bpPeer),

		requesters: make(map[int64]*bpRequester),
		height:     start,
		numPending: 0,

		requestsCh: requestsCh,
		errorsCh:   errorsCh,
	}
	bp.BaseService = *service.NewBaseService(nil, "BlockPool", bp)
	return bp
}

// OnStart implements service.Service by spawning requesters routine and recording
// pool's start time.
func (pool *BlockPool) OnStart() error {
	go pool.makeRequestersRoutine()
	pool.startTime = time.Now()
	return nil
}

// spawns requesters as needed
func (pool *BlockPool) makeRequestersRoutine() {
	for {
		if !pool.IsRunning() {
			break
		}

		_, numPending, lenRequesters := pool.GetStatus()
		if numPending >= maxPendingRequests || lenRequesters >= maxTotalRequesters {
			pool.sleepAndRemoveTimedoutPeers()
		} else {
			// request for more blocks.
			pool.makeNextRequester()
		}
	}
}

func (pool *BlockPool) removeTimedoutPeers() {
	pool.mtx.Lock()
	peersCopy := make(map[p2p.ID]*bpPeer, len(pool.peers))
	for id, peer := range pool.peers {
		peersCopy[id] = peer
	}
	pool.mtx.Unlock()

	removeCh := make(chan p2p.ID, len(peersCopy)) // Channel to collect peers to remove

	var wg sync.WaitGroup
	for peerID, peer := range peersCopy {
		wg.Add(1)
		go func(peerID p2p.ID, peer *bpPeer) {
			defer wg.Done()
			if peer.didTimeout || (peer.numPending > 0 && peer.recvMonitor.Status().CurRate < minRecvRate) {
				pool.Logger.Error("Removing timed out peer", "peer", peerID)
				removeCh <- peerID // Send peerID to be removed
			}
		}(peerID, peer)
	}

	go func() {
		wg.Wait()
		close(removeCh) // Close channel once all goroutines are done
	}()

	// Remove peers based on IDs received from removeCh
	for id := range removeCh {
		pool.mtx.Lock()
		pool.removePeer(id) // Perform removal within the lock
		pool.mtx.Unlock()
	}
}

// GetStatus returns pool's height, numPending requests and the number of
// requesters.
func (pool *BlockPool) GetStatus() (height int64, numPending int32, lenRequesters int) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	return pool.height, atomic.LoadInt32(&pool.numPending), len(pool.requesters)
}

// IsCaughtUp returns true if this node is caught up, false - otherwise.
// TODO: relax conditions, prevent abuse.
func (pool *BlockPool) IsCaughtUp() bool {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	// Need at least 1 peer to be considered caught up.
	if len(pool.peers) == 0 {
		pool.Logger.Debug("Blockpool has no peers")
		return false
	}

	// Some conditions to determine if we're caught up.
	// Ensures we've either received a block or waited some amount of time,
	// and that we're synced to the highest known height.
	// Note we use maxPeerHeight - 1 because to sync block H requires block H+1
	// to verify the LastCommit.
	receivedBlockOrTimedOut := pool.height > 0 || time.Since(pool.startTime) > 5*time.Second
	ourChainIsLongestAmongPeers := pool.maxPeerHeight == 0 || pool.height >= (pool.maxPeerHeight-1)
	isCaughtUp := receivedBlockOrTimedOut && ourChainIsLongestAmongPeers
	return isCaughtUp
}

// PeekTwoBlocks returns blocks at pool.height and pool.height+1.
// We need to see the second block's Commit to validate the first block.
// So we peek two blocks at a time.
// The caller will verify the commit.
func (pool *BlockPool) PeekTwoBlocks() (first *types.Block, second *types.Block) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	if r := pool.requesters[pool.height]; r != nil {
		first = r.getBlock()
	}
	if r := pool.requesters[pool.height+1]; r != nil {
		second = r.getBlock()
	}
	return
}

// PopRequest pops the first block at pool.height.
// It must have been validated by 'second'.Commit from PeekTwoBlocks().
func (pool *BlockPool) PopRequest() {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	if r := pool.requesters[pool.height]; r != nil {
		/*  The block can disappear at any time, due to removePeer().
		if r := pool.requesters[pool.height]; r == nil || r.block == nil {
			PanicSanity("PopRequest() requires a valid block")
		}
		*/
		if err := r.Stop(); err != nil {
			pool.Logger.Error("Error stopping requester", "err", err)
		}
		delete(pool.requesters, pool.height)
		pool.height++
	} else {
		panic(fmt.Sprintf("Expected requester to pop, got nothing at height %v", pool.height))
	}
}

// RedoRequest invalidates the block at pool.height,
// Remove the peer and redo request from others.
// Returns the ID of the removed peer.
func (pool *BlockPool) RedoRequest(height int64) p2p.ID {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	request := pool.requesters[height]
	peerID := request.getPeerID()
	if peerID != p2p.ID("") {
		// RemovePeer will redo all requesters associated with this peer.
		pool.removePeer(peerID)
	}
	return peerID
}

// AddBlock validates that the block comes from the peer it was expected from and calls the requester to store it.
// TODO: ensure that blocks come in order for each peer.
func (pool *BlockPool) AddBlock(peerID p2p.ID, block *types.Block, blockSize int) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	requester := pool.requesters[block.Height]
	if requester == nil {
		pool.Logger.Info(
			"peer sent us a block we didn't expect",
			"peer",
			peerID,
			"curHeight",
			pool.height,
			"blockHeight",
			block.Height)
		diff := pool.height - block.Height
		if diff < 0 {
			diff *= -1
		}
		if diff > maxDiffBetweenCurrentAndReceivedBlockHeight {
			pool.sendError(errors.New("peer sent us a block we didn't expect with a height too far ahead/behind"), peerID)
		}
		return
	}

	if requester.setBlock(block, peerID) {
		atomic.AddInt32(&pool.numPending, -1)
		peer := pool.peers[peerID]
		if peer != nil {
			peer.decrPending(blockSize)
		}
	} else {
		pool.Logger.Info("invalid peer", "peer", peerID, "blockHeight", block.Height)
		pool.sendError(errors.New("invalid peer"), peerID)
	}
}

// MaxPeerHeight returns the highest reported height.
func (pool *BlockPool) MaxPeerHeight() int64 {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()
	return pool.maxPeerHeight
}

// SetPeerRange sets the peer's alleged blockchain base and height.
func (pool *BlockPool) SetPeerRange(peerID p2p.ID, base int64, height int64) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	peer := pool.peers[peerID]
	if peer != nil {
		peer.base = base
		peer.height = height
	} else {
		peer = newBPPeer(pool, peerID, base, height)
		peer.setLogger(pool.Logger.With("peer", peerID))
		pool.peers[peerID] = peer
	}

	if height > pool.maxPeerHeight {
		pool.maxPeerHeight = height
	}
}

// RemovePeer removes the peer with peerID from the pool. If there's no peer
// with peerID, function is a no-op.
func (pool *BlockPool) RemovePeer(peerID p2p.ID) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	pool.removePeer(peerID)
}

func (pool *BlockPool) removePeer(peerID p2p.ID) {
	for _, requester := range pool.requesters {
		if requester.getPeerID() == peerID {
			requester.redo(peerID)
		}
	}

	peer, ok := pool.peers[peerID]
	if ok {
		if peer.timeout != nil {
			peer.timeout.Stop()
		}

		delete(pool.peers, peerID)

		// Find a new peer with the biggest height and update maxPeerHeight if the
		// peer's height was the biggest.
		if peer.height == pool.maxPeerHeight {
			pool.updateMaxPeerHeight()
		}
	}
}

// If no peers are left, maxPeerHeight is set to 0.
func (pool *BlockPool) updateMaxPeerHeight() {
	var max int64
	for _, peer := range pool.peers {
		if peer.height > max {
			max = peer.height
		}
	}
	pool.maxPeerHeight = max
}

func (pool *BlockPool) makeNextRequester() {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	nextHeight := pool.height + pool.requestersLen()
	if nextHeight > pool.maxPeerHeight || nextHeight > pool.height+maxBlocksAhead {
		return
	}

	request := newBPRequester(pool, nextHeight)

	pool.requesters[nextHeight] = request
	atomic.AddInt32(&pool.numPending, 1)

	err := request.Start()
	if err != nil {
		request.Logger.Error("Error starting request", "err", err)
	}
}

func (pool *BlockPool) requestersLen() int64 {
	return int64(len(pool.requesters))
}

func (pool *BlockPool) sendRequest(height int64, peerID p2p.ID) {
	if !pool.IsRunning() {
		return
	}
	pool.requestsCh <- BlockRequest{height, peerID}
}

func (pool *BlockPool) sendError(err error, peerID p2p.ID) {
	if !pool.IsRunning() {
		return
	}
	pool.errorsCh <- peerError{err, peerID}
}

// for debugging purposes
//
//nolint:unused
func (pool *BlockPool) debug() string {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	str := ""
	nextHeight := pool.height + pool.requestersLen()
	for h := pool.height; h < nextHeight; h++ {
		if pool.requesters[h] == nil {
			str += fmt.Sprintf("H(%v):X ", h)
		} else {
			str += fmt.Sprintf("H(%v):", h)
			str += fmt.Sprintf("B?(%v) ", pool.requesters[h].block != nil)
		}
	}
	return str
}

//-------------------------------------

type bpPeer struct {
	didTimeout  bool
	numPending  int32
	height      int64
	base        int64
	pool        *BlockPool
	id          p2p.ID
	recvMonitor *flow.Monitor

	timeout *time.Timer

	logger log.Logger
}

func newBPPeer(pool *BlockPool, peerID p2p.ID, base int64, height int64) *bpPeer {
	peer := &bpPeer{
		pool:       pool,
		id:         peerID,
		base:       base,
		height:     height,
		numPending: 0,
		logger:     log.NewNopLogger(),
	}
	return peer
}

func (peer *bpPeer) setLogger(l log.Logger) {
	peer.logger = l
}

func (peer *bpPeer) resetTimeout() {
	if peer.timeout == nil {
		peer.timeout = time.AfterFunc(peerTimeout, peer.onTimeout)
	} else {
		peer.timeout.Reset(peerTimeout)
	}
}

func (peer *bpPeer) decrPending(recvSize int) {
	peer.numPending--
	if peer.numPending == 0 {
		peer.timeout.Stop()
	} else {
		peer.recvMonitor.Update(recvSize)
		peer.resetTimeout()
	}
}

func (peer *bpPeer) onTimeout() {
	peer.pool.mtx.Lock()
	defer peer.pool.mtx.Unlock()

	err := errors.New("peer did not send us anything")
	peer.pool.sendError(err, peer.id)
	peer.logger.Error("SendTimeout", "reason", err, "timeout", peerTimeout)
	peer.didTimeout = true
}

//-------------------------------------

type bpRequester struct {
	service.BaseService
	pool       *BlockPool
	height     int64
	gotBlockCh chan struct{}
	redoCh     chan p2p.ID // redo may send multitime, add peerId to identify repeat

	mtx    cmtsync.Mutex
	peerID p2p.ID
	block  *types.Block
}

func newBPRequester(pool *BlockPool, height int64) *bpRequester {
	bpr := &bpRequester{
		pool:       pool,
		height:     height,
		gotBlockCh: make(chan struct{}, 1),
		redoCh:     make(chan p2p.ID, 1),

		peerID: "",
		block:  nil,
	}
	bpr.BaseService = *service.NewBaseService(nil, "bpRequester", bpr)
	return bpr
}

func (bpr *bpRequester) OnStart() error {
	go bpr.requestRoutine()
	return nil
}

// Returns true if the peer matches and block doesn't already exist.
func (bpr *bpRequester) setBlock(block *types.Block, peerID p2p.ID) bool {
	bpr.mtx.Lock()
	if bpr.block != nil || bpr.peerID != peerID {
		bpr.mtx.Unlock()
		return false
	}
	bpr.block = block
	bpr.mtx.Unlock()

	select {
	case bpr.gotBlockCh <- struct{}{}:
	default:
	}
	return true
}

func (bpr *bpRequester) getBlock() *types.Block {
	bpr.mtx.Lock()
	defer bpr.mtx.Unlock()
	return bpr.block
}

func (bpr *bpRequester) getPeerID() p2p.ID {
	bpr.mtx.Lock()
	defer bpr.mtx.Unlock()
	return bpr.peerID
}

// Tells bpRequester to pick another peer and try again.
// NOTE: Nonblocking, and does nothing if another redo
// was already requested.
func (bpr *bpRequester) redo(peerID p2p.ID) {
	select {
	case bpr.redoCh <- peerID:
	default:
	}
}

// Responsible for making more requests as necessary
// Returns only when a block is found (e.g. AddBlock() is called)
func (bpr *bpRequester) requestRoutine() {
	for {
		if !bpr.IsRunning() || !bpr.pool.IsRunning() {
			return
		}

		// Pick multiple peers and send requests concurrently.
		peers := bpr.pool.pickNPeersForHeight(bpr.height, 3) // pick 3 peers
		if len(peers) == 0 {
			bpr.Logger.Debug("No peers currently available; will retry shortly", "height", bpr.height)
			time.Sleep(requestIntervalMS * time.Millisecond)
			continue
		}

		// Create a channel to signal the arrival of a block.
		blockArrived := make(chan struct{}, 1)
		var wg sync.WaitGroup

		for _, peer := range peers {
			wg.Add(1)
			go func(peer *bpPeer) {
				defer wg.Done()
				bpr.pool.sendRequest(bpr.height, peer.id)
				select {
				case <-blockArrived:
					// If a block has arrived from another peer, stop this goroutine.
					return
				case <-time.After(requestRetrySeconds * time.Second):
					// Timeout for this peer; consider sending a new request or logging an error.
					bpr.Logger.Debug("Request to peer timed out", "peer", peer.id, "height", bpr.height)
				}
			}(peer)
		}

		select {
		case <-bpr.pool.Quit():
			close(blockArrived)
			wg.Wait() // Wait for all request goroutines to finish.
			return
		case <-bpr.Quit():
			close(blockArrived)
			wg.Wait() // Wait for all request goroutines to finish.
			return
		case <-bpr.gotBlockCh:
			// We got a block! Signal all goroutines to stop.
			close(blockArrived)
			wg.Wait() // Wait for all request goroutines to finish.
			return
		}
	}
}

func (pool *BlockPool) pickNPeersForHeight(height int64, n int) []*bpPeer {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	var peers []*bpPeer
	for _, peer := range pool.peers {
		if peer.didTimeout || peer.numPending >= maxPendingRequestsPerPeer {
			continue
		}
		if height >= peer.base && height <= peer.height {
			peers = append(peers, peer)
			if len(peers) == n {
				break
			}
		}
	}
	return peers
}

// BlockRequest stores a block request identified by the block Height and the PeerID responsible for
// delivering the block
type BlockRequest struct {
	Height int64
	PeerID p2p.ID
}

// sleepAndRemoveTimedoutPeers encapsulates the common pattern of sleeping
// for a specified interval and then removing timed-out peers.
func (pool *BlockPool) sleepAndRemoveTimedoutPeers() {
	time.Sleep(requestIntervalMS * time.Millisecond)
	pool.removeTimedoutPeers()
}
