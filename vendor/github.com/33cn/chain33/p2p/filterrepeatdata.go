// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p2p

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/types"
	lru "github.com/hashicorp/golang-lru"
)

var Filter = NewFilter()

func NewFilter() *Filterdata {
	filter := new(Filterdata)
	filter.regRData, _ = lru.New(P2pCacheTxSize)
	return filter
}

type Filterdata struct {
	isclose    int32
	regRData   *lru.Cache
	atomicLock sync.Mutex
}

func (f *Filterdata) GetLock() {
	f.atomicLock.Lock()
}
func (f *Filterdata) ReleaseLock() {
	f.atomicLock.Unlock()
}
func (f *Filterdata) RegRecvData(key string) bool {
	f.regRData.Add(key, time.Duration(types.Now().Unix()))
	return true
}

func (f *Filterdata) QueryRecvData(key string) bool {
	ok := f.regRData.Contains(key)
	return ok

}

func (f *Filterdata) RemoveRecvData(key string) {
	f.regRData.Remove(key)
}

func (f *Filterdata) Close() {
	atomic.StoreInt32(&f.isclose, 1)
}

func (f *Filterdata) isClose() bool {
	return atomic.LoadInt32(&f.isclose) == 1
}

func (f *Filterdata) ManageRecvFilter() {
	ticker := time.NewTicker(time.Second * 30)
	var timeout int64 = 60
	defer ticker.Stop()
	for {
		<-ticker.C
		now := types.Now().Unix()
		for _, key := range f.regRData.Keys() {
			regtime, exist := f.regRData.Get(key)
			if !exist {
				log.Warn("Not found in regRData", "Key", key)
				continue
			}
			if now-int64(regtime.(time.Duration)) < timeout {
				break
			}
			f.regRData.Remove(key)
		}

		if f.isClose() {
			return
		}
	}
}
