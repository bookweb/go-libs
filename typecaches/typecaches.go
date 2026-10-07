package typecaches

import (
	"reflect"
	"sync"
)

type CachedField struct {
	Name  string
	Type  reflect.Type
	Index []int
}

type StructInfo struct {
	Name   string
	Fields []CachedField
}

type TypeCache struct {
	cache sync.Map
}

func NewTypeCache() *TypeCache {
	return &TypeCache{}
}

func (c *TypeCache) GetInfo(v interface{}) *StructInfo {
	t := reflect.TypeOf(v)

	// Unify pointers to their base type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Fast Path: Lock-free atomic read
	if val, ok := c.cache.Load(t); ok {
		return val.(*StructInfo)
	}

	// Slow Path: Compute metadata if not cached
	info := &StructInfo{
		Name: t.Name(),
	}

	if t.Kind() == reflect.Struct {
		fields := reflect.VisibleFields(t)
		info.Fields = make([]CachedField, 0, len(fields))

		for _, f := range fields {
			if !f.IsExported() {
				continue
			}

			info.Fields = append(info.Fields, CachedField{
				Name:  f.Name,
				Type:  f.Type,
				Index: f.Index,
			})
		}
	}

	// Store atomically (or return existing if written concurrently)
	actual, _ := c.cache.LoadOrStore(t, info)
	return actual.(*StructInfo)
}
