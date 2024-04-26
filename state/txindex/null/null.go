package null

import (
	"context"
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/pubsub/query"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/state/txindex"
)

var _ txindex.TxIndexer = (*TxIndex)(nil)

// TxIndex acts as a /dev/null.
type TxIndex struct{}

func (txi *TxIndex) SetRetainHeight(_ int64) error {
	return nil
}

func (txi *TxIndex) GetRetainHeight() (int64, error) {
	return 0, nil
}

func (txi *TxIndex) Prune(_ int64) (int64, int64, error) {
	return 0, 0, nil
}

// Get on a TxIndex is disabled and panics when invoked.
func (txi *TxIndex) Get(hash []byte) (*abci.TxResult, error) {
	return nil, errors.New(`indexing is disabled (set 'tx_index = "kv"' in config)`)
}

// AddBatch is a noop and always returns nil.
func (txi *TxIndex) AddBatch(batch *txindex.Batch) error {
	return nil
}

// Index is a noop and always returns nil.
func (txi *TxIndex) Index(result *abci.TxResult) error {
	return nil
}

func (txi *TxIndex) Search(ctx context.Context, q *query.Query, pagSettings ctypes.Pagination) ([]*abci.TxResult, int, error) {
	return []*abci.TxResult{}, 0, nil
}

func (*TxIndex) SetLogger(log.Logger) {
}
