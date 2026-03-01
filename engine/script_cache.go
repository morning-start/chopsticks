package engine

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// 默认缓存配置常量
const (
	DefaultMaxCacheSize = 100 * 1024 * 1024 // 100MB
)

// 预定义错误变量
var (
	ErrComputeFileHash = errors.New("failed to compute file hash")
	ErrReadScript      = errors.New("failed to read script")
	ErrCompileScript   = errors.New("failed to compile script")
)

// ScriptCache 脚本编译缓存
// 字段按内存对齐优化排序
type ScriptCache struct {
	mu         sync.RWMutex
	entries    map[string]*cachedScript
	lruList    *list.List
	maxSize    int64 // 最大缓存大小 (字节)
	maxEntries int   // 最大条目数
	hits       int64
	misses     int64
}

// cachedScript 缓存的脚本
// 字段按内存对齐优化排序
type cachedScript struct {
	program  *goja.Program // 编译后的程序
	size     int64
	lastUsed time.Time
	element  *list.Element // LRU 列表中的元素
	path     string
	code     string
	hash     string
	useCount int
}

// NewScriptCache 创建脚本缓存
func NewScriptCache(maxSize int64, maxEntries int) *ScriptCache {
	if maxSize <= 0 {
		maxSize = DefaultMaxCacheSize
	}
	if maxEntries <= 0 {
		maxEntries = DefaultMaxCacheEntries
	}

	return &ScriptCache{
		maxSize:    maxSize,
		maxEntries: maxEntries,
		entries:    make(map[string]*cachedScript),
		lruList:    list.New(),
	}
}

// Get 获取缓存的脚本
func (c *ScriptCache) Get(path string) (*cachedScript, bool) {
	c.mu.RLock()
	entry, exists := c.entries[path]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// 验证文件是否被修改
	currentHash, err := c.computeFileHash(path)
	if err != nil || currentHash != entry.hash {
		// 文件已修改，删除缓存
		c.mu.Lock()
		c.removeEntry(entry)
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// 更新 LRU
	c.mu.Lock()
	c.lruList.MoveToFront(entry.element)
	entry.lastUsed = time.Now()
	entry.useCount++
	c.hits++
	c.mu.Unlock()

	return entry, true
}

// Set 缓存脚本
func (c *ScriptCache) Set(path string, code string, program *goja.Program) error {
	hash, err := c.computeFileHash(path)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrComputeFileHash, err)
	}

	size := int64(len(code))

	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已存在，更新
	if entry, exists := c.entries[path]; exists {
		entry.code = code
		entry.program = program
		entry.hash = hash
		entry.size = size
		entry.lastUsed = time.Now()
		entry.useCount++
		c.lruList.MoveToFront(entry.element)
		return nil
	}

	// 检查是否需要淘汰
	for (c.currentSize()+size > c.maxSize || len(c.entries) >= c.maxEntries) && c.lruList.Len() > 0 {
		c.evictLRU()
	}

	// 创建新条目
	entry := &cachedScript{
		path:     path,
		code:     code,
		program:  program,
		hash:     hash,
		size:     size,
		lastUsed: time.Now(),
		useCount: 1,
	}

	entry.element = c.lruList.PushFront(entry)
	c.entries[path] = entry

	return nil
}

// Remove 删除缓存
func (c *ScriptCache) Remove(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[path]; exists {
		c.removeEntry(entry)
	}
}

// Clear 清空缓存
func (c *ScriptCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cachedScript)
	c.lruList.Init()
}

// Stats 返回缓存统计
func (c *ScriptCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:    c.hits,
		Misses:  c.misses,
		HitRate: hitRate,
		Size:    c.currentSize(),
		MaxSize: c.maxSize,
		Entries: len(c.entries),
	}
}

// CacheStats 缓存统计
// int64 字段放在一起以优化内存布局
type CacheStats struct {
	Hits    int64
	Misses  int64
	Size    int64
	MaxSize int64
	HitRate float64
	Entries int
}

// computeFileHash 计算文件哈希
func (c *ScriptCache) computeFileHash(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:]), nil
}

// currentSize 计算当前缓存大小
func (c *ScriptCache) currentSize() int64 {
	var size int64
	for _, entry := range c.entries {
		size += entry.size
	}
	return size
}

// evictLRU 淘汰最久未使用的条目
func (c *ScriptCache) evictLRU() {
	elem := c.lruList.Back()
	if elem == nil {
		return
	}

	entry := elem.Value.(*cachedScript)
	c.removeEntry(entry)
}

// removeEntry 删除条目
func (c *ScriptCache) removeEntry(entry *cachedScript) {
	c.lruList.Remove(entry.element)
	delete(c.entries, entry.path)
}

// LoadAndCompile 加载并编译脚本（带缓存）
func (c *ScriptCache) LoadAndCompile(vm *goja.Runtime, path string) (*goja.Program, error) {
	// 尝试从缓存获取
	c.mu.RLock()
	entry, exists := c.entries[path]
	c.mu.RUnlock()

	if exists {
		// 验证文件是否被修改
		currentHash, err := c.computeFileHash(path)
		if err != nil || currentHash != entry.hash {
			// 文件已修改，删除缓存
			c.mu.Lock()
			c.removeEntry(entry)
			c.misses++
			c.mu.Unlock()
		} else {
			// 缓存命中
			c.mu.Lock()
			c.lruList.MoveToFront(entry.element)
			entry.lastUsed = time.Now()
			entry.useCount++
			c.hits++
			c.mu.Unlock()
			return entry.program, nil
		}
	} else {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
	}

	// 读取文件
	code, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrReadScript, path, err)
	}

	// 编译脚本
	program, err := goja.Compile(path, string(code), false)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrCompileScript, path, err)
	}

	// 缓存编译结果
	c.Set(path, string(code), program)

	return program, nil
}
