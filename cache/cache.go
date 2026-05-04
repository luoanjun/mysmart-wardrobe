package cache

import (
	"sync"
	"time"
	"wardrobe/models"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type MemoryCache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

var instance *MemoryCache

func InitCache() {
	instance = &MemoryCache{
		items: make(map[string]*CacheItem),
	}
}

func GetCache() *MemoryCache {
	return instance
}

func (c *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items[key] = &CacheItem{
		Value:      value,
		Expiration: time.Now().Add(duration).UnixNano(),
	}
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, found := c.items[key]
	if !found {
		return nil, false
	}
	
	if time.Now().UnixNano() > item.Expiration {
		return nil, false
	}
	
	return item.Value, true
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *MemoryCache) DeleteByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.items, key)
		}
	}
}

func (c *MemoryCache) InvalidateClothesCache() {
	c.DeleteByPrefix("clothes:")
}

func (c *MemoryCache) GetClothesList(key string) ([]models.Cloth, bool) {
	val, found := c.Get(key)
	if !found {
		return nil, false
	}
	clothes, ok := val.([]models.Cloth)
	return clothes, ok
}

func (c *MemoryCache) SetClothesList(key string, clothes []models.Cloth) {
	c.Set(key, clothes, 5*time.Minute)
}
