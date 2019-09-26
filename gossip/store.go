package gossip

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/bbva/qed/protocol"

	"github.com/bbva/qed/util"
	"github.com/google/btree"
)

type SnapshotStore interface {
	PutBatch(b *protocol.BatchSnapshots) error
	PutSnapshot(version uint64, snapshot *protocol.SignedSnapshot) error
	GetRange(start, end uint64) ([]protocol.SignedSnapshot, error)
	GetSnapshot(version uint64) (*protocol.SignedSnapshot, error)
	DeleteRange(start, end uint64) error
	Count() (uint64, error)
}

// RestSnapshotStore implements access to a snapshot store
// in a http rest service.
// The process of sending the notifications is
// asynchronous, so a start and stop method is
type RestSnapshotStore struct {
	endpoint []string
	client   *http.Client
}

//RestSnapshotStore configuration object used to parse
//cli options and to build the SimpleNotifier instance
type RestSnapshotStoreConfig struct {
	Endpoint    []string      `desc:"REST snapshot store service endpoint list http://ip1:port1/path1,http://ip2:port2/path2... "`
	DialTimeout time.Duration `desc:"Timeout dialing the REST snapshot store service"`
	ReadTimeout time.Duration `desc:"Timeout reading the REST snapshot store service response"`
}

func NewRestSnapshotStoreFromConfig(c *RestSnapshotStoreConfig) *RestSnapshotStore {
	return NewRestSnapshotStore(c.Endpoint, c.DialTimeout, c.ReadTimeout)
}

func DefaultRestSnapshotStoreConfig() *RestSnapshotStoreConfig {
	return &RestSnapshotStoreConfig{
		DialTimeout: 200 * time.Millisecond,
		ReadTimeout: 200 * time.Millisecond,
	}

}

// Returns a new RestSnapshotStore client
func NewRestSnapshotStore(endpoint []string, dialTimeout, readTimeout time.Duration) *RestSnapshotStore {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				// timeout calling the server
				conn, err := net.DialTimeout(netw, addr, dialTimeout)
				if err != nil {
					return nil, err
				}
				// timeout reading from the connection
				_ = conn.SetDeadline(time.Now().Add(readTimeout))
				return conn, nil
			},
		}}

	return &RestSnapshotStore{
		endpoint: endpoint,
		client:   client,
	}
}

// Stores a batch int the store
func (r *RestSnapshotStore) PutBatch(b *protocol.BatchSnapshots) error {
	buf, err := b.Encode()
	if err != nil {
		return err
	}
	n := len(r.endpoint)
	if n == 0 {
		return fmt.Errorf("No endpoint configured for snapshot store!")
	}
	url := r.endpoint[0]
	if n > 1 {
		url = r.endpoint[rand.Intn(n)]
	}
	resp, err := r.client.Post(url+"/batch", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (r *RestSnapshotStore) PutSnapshot(version uint64, snapshot *protocol.SignedSnapshot) error {
	panic("not implemented")
}

func (r *RestSnapshotStore) GetRange(start uint64, end uint64) ([]protocol.SignedSnapshot, error) {
	panic("not implemented")
}

func (r *RestSnapshotStore) GetSnapshot(version uint64) (*protocol.SignedSnapshot, error) {
	n := len(r.endpoint)
	url := r.endpoint[0]
	if n > 1 {
		url = r.endpoint[rand.Intn(n)]
	}
	resp, err := r.client.Get(fmt.Sprintf("%s/snapshot?v=%d", url, version))
	if err != nil {
		return nil, fmt.Errorf("Error getting snapshot %d from store because %v", version, err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error getting snapshot from the store. Status: %d", resp.StatusCode)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var s protocol.SignedSnapshot
	err = s.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("Error decoding signed snapshot %d codec", s.Snapshot.Version)
	}
	return &s, nil
}

func (r *RestSnapshotStore) DeleteRange(start uint64, end uint64) error {
	panic("not implemented")
}

func (r *RestSnapshotStore) Count() (uint64, error) {
	n := len(r.endpoint)
	url := r.endpoint[0]
	if n > 1 {
		url = r.endpoint[rand.Intn(n)]
	}
	resp, err := r.client.Get(url + "/count")
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Error getting snapshot from the store. Status: %d", resp.StatusCode)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	count, err := strconv.ParseUint(string(buf), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Error parsin store response: got %d", resp.StatusCode)
	}

	return count, nil
}

type BPlusTreeStore struct {
	db *btree.BTree
}

type StoreItem struct {
	Key, Value []byte
}

func (p StoreItem) Less(b btree.Item) bool {
	return bytes.Compare(p.Key, b.(StoreItem).Key) < 0
}

func (s *BPlusTreeStore) Count() (uint64, error) {
	panic("not implemented")
}

func (s *BPlusTreeStore) PutSnapshot(version uint64, snapshot protocol.Snapshot) error {
	encoded, err := snapshot.Encode()
	if err != nil {
		return err
	}
	s.db.ReplaceOrInsert(StoreItem{util.Uint64AsBytes(version), encoded})
	return nil
}

func (s BPlusTreeStore) GetRange(start, end uint64) ([]protocol.Snapshot, error) {
	result := make([]protocol.Snapshot, 0)
	startKey := util.Uint64AsBytes(start)
	endKey := util.Uint64AsBytes(end)
	s.db.AscendGreaterOrEqual(StoreItem{startKey, nil}, func(i btree.Item) bool {
		key := i.(StoreItem).Key
		if bytes.Compare(key, endKey) > 0 {
			return false
		}
		var snapshot protocol.Snapshot
		if err := snapshot.Decode(i.(StoreItem).Value); err != nil {
			return false
		}
		result = append(result, snapshot)
		return true
	})
	return result, nil
}

func (s *BPlusTreeStore) DeleteRange(start, end uint64) error {
	startKey := util.Uint64AsBytes(start)
	endKey := util.Uint64AsBytes(end)
	s.db.AscendGreaterOrEqual(StoreItem{startKey, nil}, func(i btree.Item) bool {
		key := i.(StoreItem).Key
		if bytes.Compare(key, endKey) > 0 {
			return false
		}
		s.db.Delete(i)
		return true
	})
	return nil
}
