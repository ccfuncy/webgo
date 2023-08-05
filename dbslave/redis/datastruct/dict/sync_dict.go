package dict

import "sync"

type SyncDict struct {
	m sync.Map
}

func (dict *SyncDict) Get(key string) (any, bool) {
	return dict.m.Load(key)
}

func (dict *SyncDict) Len() int {
	var res int
	dict.m.Range(func(key, value any) bool {
		res++
		return true
	})
	return res
}

func (dict *SyncDict) Put(key string, value interface{}) int {
	_, ok := dict.m.Load(key)
	dict.m.Store(key, value)
	if ok {
		return 0
	}
	return 1
}

func (dict *SyncDict) PutIfAbsent(key string, value any) int {
	_, ok := dict.m.Load(key)
	if ok {
		return 0
	}
	dict.m.Store(key, value)
	return 1
}

func (dict *SyncDict) PutIfExist(key string, value any) int {
	_, ok := dict.m.Load(key)
	if !ok {
		return 0
	}
	dict.m.Store(key, value)
	return 1
}

func (dict *SyncDict) Remove(key string) int {
	_, ok := dict.m.Load(key)
	dict.m.Delete(key)
	if ok {
		return 1
	}
	return 0
}

func (dict *SyncDict) ForEach(consumer Consumer) {
	dict.m.Range(func(key, value any) bool {
		consumer(key.(string), value)
		return true
	})
}

func (dict *SyncDict) Keys() []string {
	res := make([]string, dict.Len())
	i := 0
	dict.m.Range(func(key, value any) bool {
		res[i] = key.(string)
		i++
		return true
	})
	return res
}

func (dict *SyncDict) RandomKeys(limit int) []string {
	res := make([]string, limit)
	for i := 0; i < limit; i++ {
		dict.m.Range(func(key, value any) bool {
			res[i] = key.(string)
			return false
		})
	}
	return res
}

func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
	res := make([]string, limit)
	i := 0
	dict.m.Range(func(key, value any) bool {
		res[i] = key.(string)
		i++
		if i >= limit {
			return false
		}
		return true
	})
	return res
}

func (dict *SyncDict) Clear() {
	*dict = *NewSyncDict()
}

func NewSyncDict() *SyncDict {
	return &SyncDict{}
}
