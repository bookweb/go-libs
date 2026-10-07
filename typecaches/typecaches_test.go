package typecaches

import (
	"fmt"
	"testing"
)

type AuditInfo struct {
	CreatedAt string
}

type Employee struct {
	AuditInfo
	ID   int
	Name string
}

func TestTypeCacheSuccess(t *testing.T) {
	cache := NewTypeCache()
	emp := Employee{ID: 99, Name: "Bob"}

	info := cache.GetInfo(&emp)

	fmt.Printf("Struct: %s\n", info.Name)
	fmt.Println("Visible Fields:")
	for _, f := range info.Fields {
		fmt.Printf(" - Name: %-10s | Type: %-10s | Index: %v\n", f.Name, f.Type, f.Index)
	}
}
