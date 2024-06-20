package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 定义了哈希函数类型
type Hash func(data []byte) uint32

// Map 包含了所有的哈希键
type Map struct {
	hash     Hash
	replicas int
	keys     []int          // Sorted
	hashMap  map[int]string // 虚拟节点映射
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// Add 添加一些节点到哈希环中
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}

	sort.Ints(m.keys)
}

// Get 获取与给定键最近的节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// 二分查找第一个大于等于hash的节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 使用取余数的方式处理哈希环的边界情况，即 idx = len(keys) 的情况
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
