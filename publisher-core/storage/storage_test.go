package storage

import (
	"context"
	"testing"
)

func TestNewLocalStorage(t *testing.T) {
	storage, err := NewLocalStorage("./test-uploads", "")

	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}

	if storage == nil {
		t.Fatal("Storage should not be nil")
	}
}

func TestLocalStorageWriteAndRead(t *testing.T) {
	storage, _ := NewLocalStorage("./test-uploads", "")

	key := "test-key.txt"
	data := []byte("test data")

	err := storage.Write(context.Background(), key, data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	retrieved, err := storage.Read(context.Background(), key)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if string(retrieved) != string(data) {
		t.Errorf("Data mismatch, got '%s', want '%s'", string(retrieved), string(data))
	}
}

func TestLocalStorageDelete(t *testing.T) {
	storage, _ := NewLocalStorage("./test-uploads", "")

	key := "test-delete.txt"
	data := []byte("test data")

	_ = storage.Write(context.Background(), key, data)

	err := storage.Delete(context.Background(), key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = storage.Read(context.Background(), key)
	if err == nil {
		t.Error("Read should fail after Delete")
	}
}

func TestLocalStorageList(t *testing.T) {
	storage, _ := NewLocalStorage("./test-uploads", "")

	_ = storage.Write(context.Background(), "prefix/key1.txt", []byte("data1"))
	_ = storage.Write(context.Background(), "prefix/key2.txt", []byte("data2"))
	_ = storage.Write(context.Background(), "other/key3.txt", []byte("data3"))

	keys, err := storage.List(context.Background(), "prefix")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(keys) < 2 {
		t.Errorf("Expected at least 2 keys with prefix, got %d", len(keys))
	}
}

func TestLocalStorageExists(t *testing.T) {
	storage, _ := NewLocalStorage("./test-uploads", "")

	key := "test-exists.txt"

	exists, err := storage.Exists(context.Background(), key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist initially")
	}

	_ = storage.Write(context.Background(), key, []byte("data"))

	exists, err = storage.Exists(context.Background(), key)
	if err != nil {
		t.Fatalf("Exists failed after write: %v", err)
	}
	if !exists {
		t.Error("Key should exist after write")
	}
}

func TestNewBufferStorage(t *testing.T) {
	storage := NewBufferStorage()

	if storage == nil {
		t.Fatal("BufferStorage should not be nil")
	}
}

func TestBufferStorageWriteAndRead(t *testing.T) {
	storage := NewBufferStorage()

	key := "test-key"
	data := []byte("test data")

	err := storage.Write(context.Background(), key, data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	retrieved, err := storage.Read(context.Background(), key)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if string(retrieved) != string(data) {
		t.Errorf("Data mismatch, got '%s', want '%s'", string(retrieved), string(data))
	}
}

func TestBufferStorageDelete(t *testing.T) {
	storage := NewBufferStorage()

	key := "test-key"
	data := []byte("test data")

	_ = storage.Write(context.Background(), key, data)

	err := storage.Delete(context.Background(), key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = storage.Read(context.Background(), key)
	if err == nil {
		t.Error("Read should fail after Delete")
	}
}

func TestBufferStorageList(t *testing.T) {
	storage := NewBufferStorage()

	_ = storage.Write(context.Background(), "prefix/key1", []byte("data1"))
	_ = storage.Write(context.Background(), "prefix/key2", []byte("data2"))
	_ = storage.Write(context.Background(), "other/key3", []byte("data3"))

	keys, err := storage.List(context.Background(), "prefix")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys with prefix, got %d", len(keys))
	}
}

func TestBufferStorageExists(t *testing.T) {
	storage := NewBufferStorage()

	key := "test-key"

	exists, err := storage.Exists(context.Background(), key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist initially")
	}

	_ = storage.Write(context.Background(), key, []byte("data"))

	exists, err = storage.Exists(context.Background(), key)
	if err != nil {
		t.Fatalf("Exists failed after write: %v", err)
	}
	if !exists {
		t.Error("Key should exist after write")
	}
}
