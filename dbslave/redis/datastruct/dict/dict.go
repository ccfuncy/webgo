package dict

type Consumer func(key string, val any) bool

type Dict interface {
	Get(key string) (any, bool)
	Len() int
	Put(key string, value interface{}) int
	PutIfAbsent(key string, value any) int //如果没有才放入
	PutIfExist(key string, value any) int  //如果有才放
	Remove(key string) int
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string
	Clear()
}
