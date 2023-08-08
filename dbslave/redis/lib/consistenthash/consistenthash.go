package consistenthash

import (
	"hash/crc32"
	"redis/interface/utils"
	"sort"
)

//哈希环

// 哈希函数
type HashFunc func([]byte) uint32

// NodeMap 存储所有节点信息及一致性哈希的信息
type NodeMap struct {
	hashFunc    HashFunc
	nodeHashs   []int //哈希值 1,5,9 6
	nodehashMap map[int]string
}

func NewNodeMap(hashFunc HashFunc) *NodeMap {
	if hashFunc == nil {
		hashFunc = crc32.ChecksumIEEE
	}
	return &NodeMap{
		hashFunc:    hashFunc,
		nodehashMap: make(map[int]string),
	}
}

func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

func (m *NodeMap) AddNodes(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc(utils.StringToBytes(key)))
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodehashMap[hash] = key
	}
	sort.Ints(m.nodeHashs)
}

func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hashFunc(utils.StringToBytes(key)))
	index := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})
	return m.nodehashMap[m.nodeHashs[index%len(m.nodeHashs)]]
}
