package nodes

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
)

type testNode struct {
	id    string
	forks []uint64
	seeds []int
	head  int // points to where we're currently at, in the chain
	mu    sync.Mutex
	// Counters
	queriedNumbers map[uint64]interface{}
	totalQueries   int
}

func hashFromSeed(seed int, number uint64) common.Hash {
	hash := make([]byte, 32)
	binary.BigEndian.PutUint64(hash, uint64(seed))
	binary.BigEndian.PutUint64(hash[8:], uint64(number))
	return crypto.Keccak256Hash(hash)
	//return common.BytesToHash(hash)
}

func (t *testNode) seedAt(number uint64) int {
	// Search uses binary search to find and return the smallest index i
	// in [0, n) at which f(i) is true
	seed := t.seeds[0]
	for i, fork := range t.forks {
		if fork <= number {
			seed = t.seeds[i]
		}
	}
	return seed
}

func (t *testNode) Status() int {
	return 0
}

func (t *testNode) SetStatus(int) {}

func (t *testNode) Version() (string, error) {
	return "TestNode/v0.1/darwin/go1.4.1", nil
}

func (t *testNode) Name() string {
	return fmt.Sprintf("TestNode(%v)", t.id)
}

func (t *testNode) UpdateLatest() error {
	return nil
}

func (t *testNode) BlockAt(num uint64, force bool) *blockInfo {
	t.mu.Lock()
	defer t.mu.Unlock()
	if num > uint64(t.head) {
		return nil
	}
	t.queriedNumbers[num] = num
	t.totalQueries++

	hash := hashFromSeed(t.seedAt(num), num)
	//log.Info("BlockAt", "node", t.id, "query", num, "hash", hash, "seed", t.seedAt(num))
	return &blockInfo{
		num:   num,
		hash:  hash,
		pHash: hashFromSeed(t.seedAt(num-1), num-1),
	}
}

func (t *testNode) HashAt(num uint64, force bool) common.Hash {
	if bl := t.BlockAt(num, force); bl != nil {
		return bl.hash
	}
	return common.Hash{}
}

func (t *testNode) HeadNum() uint64 {
	return uint64(t.head)
}

func (t *testNode) LastProgress() int64 {
	return 0
}

func (t *testNode) BadBlocks() []*eth.BadBlockArgs {
	return []*eth.BadBlockArgs{}
}

func newTestNode(id string, head int, forks []uint64, seeds []int) *testNode {
	return &testNode{
		id:             id,
		forks:          forks,
		seeds:          seeds,
		head:           head,
		queriedNumbers: make(map[uint64]interface{}),
	}
}

var testNodeId int64

func NewLiveTestNode(id string, head int, forks []uint64, seeds []int) *testNode {
	node := newTestNode(fmt.Sprintf("%v-%d", id, atomic.AddInt64(&testNodeId, 1)), head, forks, seeds)
	go func() {
		for {
			select {
			case <-time.After(time.Second*10 + time.Duration(rand.Int31n(4))*time.Second):
			}
			node.head++
		}
	}()
	return node
}
