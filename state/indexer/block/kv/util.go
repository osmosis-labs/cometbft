package kv

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"

	"github.com/google/orderedcode"

	"github.com/cometbft/cometbft/libs/pubsub/query"
	"github.com/cometbft/cometbft/state/indexer"
	"github.com/cometbft/cometbft/types"
)

type HeightInfo struct {
	heightRange     indexer.QueryRange
	height          int64
	heightEqIdx     int
	onlyHeightRange bool
	onlyHeightEq    bool
}

func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

func int64FromBytes(bz []byte) int64 {
	v, _ := binary.Varint(bz)
	return v
}

func int64ToBytes(i int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, i)
	return buf[:n]
}

func heightKey(height int64) ([]byte, error) {
	return orderedcode.Append(
		nil,
		types.BlockHeightKey,
		height,
	)
}

// keyBelongsToHeightRange checks whether the passed key  belongs to the specified height range.
// That might not be the case for two reasons: the key belongs to the height out of the specified range,
// or the key doesn't belong to any height, meaning it's neither heightKey nor eventKey.
func keyBelongsToHeightRange(key []byte, left, right int64) bool {
	// left included, right excluded
	eventHeight, err := parseHeightFromEventKey(key)
	if err == nil {
		return left <= eventHeight && eventHeight < right
	}

	var (
		blockHeightKeyPrefix string
		possibleHeight       int64
	)
	remaining, err := orderedcode.Parse(string(key), &blockHeightKeyPrefix, &possibleHeight)
	if err != nil {
		return false
	}
	return len(remaining) == 0 &&
		blockHeightKeyPrefix == types.BlockHeightKey &&
		left <= possibleHeight &&
		possibleHeight < right
}

// getHeightFromKey requires its argument key to be the key with height (either heightKey or eventKey).
// Provided that, it extracts the height.
func getHeightFromKey(key []byte) int64 {
	// Must be called with either heightKey or eventKey
	eventHeight, err := parseHeightFromEventKey(key)
	if err == nil {
		return eventHeight
	}

	var (
		blockHeightKeyPrefix string
		possibleHeight       int64
	)
	remaining, err := orderedcode.Parse(string(key), &blockHeightKeyPrefix, &possibleHeight)
	if err != nil {
		panic(err)
	}
	if len(remaining) == 0 && blockHeightKeyPrefix == types.BlockHeightKey {
		return possibleHeight
	}
	panic(fmt.Errorf("key must be either heightKey or eventKey"))
}

func eventKey(compositeKey, typ, eventValue string, height int64, eventSeq int64) ([]byte, error) {
	return orderedcode.Append(
		nil,
		compositeKey,
		eventValue,
		height,
		typ,
		eventSeq,
	)
}

func parseValueFromPrimaryKey(key []byte) (string, error) {
	var (
		compositeKey string
		height       int64
	)

	remaining, err := orderedcode.Parse(string(key), &compositeKey, &height)
	if err != nil {
		return "", fmt.Errorf("failed to parse event key: %w", err)
	}

	if len(remaining) != 0 {
		return "", fmt.Errorf("unexpected remainder in key: %s", remaining)
	}

	return strconv.FormatInt(height, 10), nil
}

func parseValueFromEventKey(key []byte) (string, error) {
	var (
		compositeKey, typ, eventValue string
		height                        int64
	)

	_, err := orderedcode.Parse(string(key), &compositeKey, &eventValue, &height, &typ)
	if err != nil {
		return "", fmt.Errorf("failed to parse event key: %w", err)
	}

	return eventValue, nil
}

func parseHeightFromEventKey(key []byte) (int64, error) {
	var (
		compositeKey, typ, eventValue string
		height                        int64
	)

	_, err := orderedcode.Parse(string(key), &compositeKey, &eventValue, &height, &typ)
	if err != nil {
		return -1, fmt.Errorf("failed to parse event key: %w", err)
	}

	return height, nil
}

func parseEventSeqFromEventKey(key []byte) (int64, error) {
	var (
		compositeKey, typ, eventValue string
		height                        int64
		eventSeq                      int64
	)

	remaining, err := orderedcode.Parse(string(key), &compositeKey, &eventValue, &height, &typ)
	if err != nil {
		return 0, fmt.Errorf("failed to parse event key: %w", err)
	}

	// This is done to support previous versions that did not have event sequence in their key
	if len(remaining) != 0 {
		remaining, err = orderedcode.Parse(remaining, &eventSeq)
		if err != nil {
			return 0, fmt.Errorf("failed to parse event key: %w", err)
		}
		if len(remaining) != 0 {
			return 0, fmt.Errorf("unexpected remainder in key: %s", remaining)
		}
	}

	return eventSeq, nil
}

// Remove all occurrences of height equality queries except one. While we are traversing the conditions, check whether the only condition in
// addition to match events is the height equality or height range query. At the same time, if we do have a height range condition
// ignore the height equality condition. If a height equality exists, place the condition index in the query and the desired height
// into the heightInfo struct
func dedupHeight(conditions []query.Condition) (dedupConditions []query.Condition, heightInfo HeightInfo, found bool) {
	heightInfo.heightEqIdx = -1
	heightRangeExists := false
	var heightCondition []query.Condition
	heightInfo.onlyHeightEq = true
	heightInfo.onlyHeightRange = true
	for _, c := range conditions {
		if c.CompositeKey == types.BlockHeightKey {
			if c.Op == query.OpEqual {
				if found || heightRangeExists {
					continue
				} else {
					heightCondition = append(heightCondition, c)
					heightInfo.height = c.Operand.(*big.Int).Int64() // As height is assumed to always be int64
					found = true
				}
			} else {
				heightInfo.onlyHeightEq = false
				heightRangeExists = true
				dedupConditions = append(dedupConditions, c)
			}
		} else {
			heightInfo.onlyHeightRange = false
			heightInfo.onlyHeightEq = false
			dedupConditions = append(dedupConditions, c)
		}
	}
	if !heightRangeExists && len(heightCondition) != 0 {
		heightInfo.heightEqIdx = len(dedupConditions)
		heightInfo.onlyHeightRange = false
		dedupConditions = append(dedupConditions, heightCondition...)
	} else {
		// If we found a range make sure we set the hegiht idx to -1 as the height equality
		// will be removed
		heightInfo.heightEqIdx = -1
		heightInfo.height = 0
		heightInfo.onlyHeightEq = false
		found = false
	}
	return dedupConditions, heightInfo, found
}

func checkHeightConditions(heightInfo HeightInfo, keyHeight int64) bool {
	if heightInfo.heightRange.Key != "" {
		if !checkBounds(heightInfo.heightRange, big.NewInt(keyHeight)) {
			return false
		}
	} else {
		if heightInfo.height != 0 && keyHeight != heightInfo.height {
			return false
		}
	}
	return true
}

//nolint:unused,deadcode
func lookForHeight(conditions []query.Condition) (int64, bool, int) {
	for i, c := range conditions {
		if c.CompositeKey == types.BlockHeightKey && c.Op == query.OpEqual {
			return c.Operand.(int64), true, i
		}
	}

	return 0, false, -1
}
