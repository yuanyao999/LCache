package store

import (
	"testing"
	"time"
)

type testString string

func (s testString) Len() int {
	return len(s)
}

func TestLRUCache_SetAndGet(t *testing.T) {
	cache := newLRUCache(Options{MaxBytes: 1024})
	defer cache.Close()

	cache.Set("a", testString("1"))
	val, ok := cache.Get("a")
	if !ok || string(val.(testString)) != "1" {
		t.Fatalf("expected value '1', got %v", val)
	}
}

func TestLRUCache_Expiration(t *testing.T) {
	cache := newLRUCache(Options{MaxBytes: 1024, CleanupInterval: 100 * time.Millisecond})
	defer cache.Close()

	cache.SetWithExpiration("a", testString("1"), 200*time.Millisecond)
	time.Sleep(300 * time.Millisecond)

	val, ok := cache.Get("a")
	if ok || val != nil {
		t.Fatal("expected key 'a' to be expired and removed")
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := newLRUCache(Options{MaxBytes: int64(len("a") + len("1") + len("b") + len("2"))})
	defer cache.Close()

	cache.Set("a", testString("1"))
	cache.Set("b", testString("2"))
	cache.Set("c", testString("3")) // This should trigger eviction

	if _, ok := cache.Get("a"); ok {
		t.Fatal("expected 'a' to be evicted")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	cache := newLRUCache(Options{MaxBytes: 1024})
	defer cache.Close()

	cache.Set("x", testString("value"))
	cache.Delete("x")

	if _, ok := cache.Get("x"); ok {
		t.Fatal("expected 'x' to be deleted")
	}
}

func TestLRUCache_UpdateExpiration(t *testing.T) {
	cache := newLRUCache(Options{MaxBytes: 1024, CleanupInterval: 200 * time.Millisecond})
	defer cache.Close()

	cache.SetWithExpiration("x", testString("val"), 300*time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	cache.UpdateExpiration("x", 500*time.Millisecond) // Extend TTL
	time.Sleep(400 * time.Millisecond)

	if _, ok := cache.Get("x"); !ok {
		t.Fatal("expected 'x' to be still alive after TTL extension")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := newLRUCache(Options{MaxBytes: 1024})
	defer cache.Close()

	cache.Set("a", testString("1"))
	cache.Set("b", testString("2"))
	cache.Clear()

	if cache.Len() != 0 {
		t.Fatalf("expected cache to be empty after Clear(), got %d items", cache.Len())
	}
}
