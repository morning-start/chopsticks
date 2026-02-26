// Package storage 提供存储功能。
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/manifest"

	bolt "go.etcd.io/bbolt"
)

var (
	// installedBucket 是已安装应用的 bucket 名称。
	installedBucket = []byte("installed")
	// bucketsBucket 是软件源的 bucket 名称。
	bucketsBucket = []byte("buckets")
	// operationsBucket 是安装操作的 bucket 名称。
	operationsBucket = []byte("operations")
	// systemOperationsBucket 是系统操作的 bucket 名称。
	systemOperationsBucket = []byte("system_operations")
	// nameKey 是应用名称的 key（预转换避免重复转换）。
	nameKey = []byte("name")
)

// 一个安装操作记录 InstallOperation 表示。
type InstallOperation struct {
	ID           string
	InstalledID  string
	Operation    string
	TargetPath   string
	TargetValue  string
	CreatedAt    string
}

// SystemOperation 表示一个系统操作记录。
type SystemOperation struct {
	ID            string
	InstalledID   string
	Operation     string
	TargetType    string
	TargetPath    string
	TargetKey     string
	TargetValue   string
	OriginalValue string
	CreatedAt     string
}

// Storage 定义存储接口。
type Storage interface {
	// 已安装的应用
	SaveInstalledApp(ctx context.Context, a *manifest.InstalledApp) error
	GetInstalledApp(ctx context.Context, name string) (*manifest.InstalledApp, error)
	DeleteInstalledApp(ctx context.Context, name string) error
	ListInstalledApps(ctx context.Context) ([]*manifest.InstalledApp, error)
	IsInstalled(ctx context.Context, name string) (bool, error)

	// 软件源
	SaveBucket(ctx context.Context, b *manifest.BucketConfig) error
	GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error)
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) ([]*manifest.BucketConfig, error)

	// 安装操作追踪
	SaveInstallOperation(ctx context.Context, op *InstallOperation) error
	GetInstallOperations(ctx context.Context, installedID string) ([]*InstallOperation, error)
	DeleteInstallOperations(ctx context.Context, installedID string) error

	// 系统操作追踪
	SaveSystemOperation(ctx context.Context, op *SystemOperation) error
	GetSystemOperations(ctx context.Context, installedID string) ([]*SystemOperation, error)
	DeleteSystemOperations(ctx context.Context, installedID string) error

	Close() error
}

// boltStorage 是 Storage 的 BoltDB 实现。
type boltStorage struct {
	db   *bolt.DB
	path string
}

// 编译时接口检查。
var _ Storage = (*boltStorage)(nil)

// New 创建新的 Storage。
func New(path string) (Storage, error) {
	// 确保数据库目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录: %w", err)
	}

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("打开 bolt db: %w", err)
	}

	// 创建 buckets
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(installedBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bucketsBucket); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("创建 buckets: %w", err)
	}

	return &boltStorage{
		db:   db,
		path: path,
	}, nil
}

// SaveInstalledApp 保存已安装的应用。
func (s *boltStorage) SaveInstalledApp(ctx context.Context, a *manifest.InstalledApp) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(installedBucket)
		data, err := json.Marshal(a)
		if err != nil {
			return fmt.Errorf("序列化应用: %w", err)
		}
		return b.Put([]byte(a.Name), data)
	})
}

// GetInstalledApp 获取已安装的应用。
func (s *boltStorage) GetInstalledApp(ctx context.Context, name string) (*manifest.InstalledApp, error) {
	var app manifest.InstalledApp
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(installedBucket)
		data := b.Get([]byte(name))
		if data == nil {
			return fmt.Errorf("应用不存在: %s", name)
		}
		return json.Unmarshal(data, &app)
	})
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// DeleteInstalledApp 删除已安装的应用。
func (s *boltStorage) DeleteInstalledApp(ctx context.Context, name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(installedBucket)
		return b.Delete([]byte(name))
	})
}

// ListInstalledApps 列出所有已安装的应用。
func (s *boltStorage) ListInstalledApps(ctx context.Context) ([]*manifest.InstalledApp, error) {
	var apps []*manifest.InstalledApp
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(installedBucket)
		return b.ForEach(func(k, v []byte) error {
			var app manifest.InstalledApp
			if err := json.Unmarshal(v, &app); err != nil {
				return err
			}
			apps = append(apps, &app)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return apps, nil
}

// IsInstalled 检查应用是否已安装。
func (s *boltStorage) IsInstalled(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(installedBucket)
		data := b.Get([]byte(name))
		exists = data != nil
		return nil
	})
	return exists, err
}

// SaveBucket 保存软件源配置。
func (s *boltStorage) SaveBucket(ctx context.Context, b *manifest.BucketConfig) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketsBucket)
		data, err := json.Marshal(b)
		if err != nil {
			return fmt.Errorf("序列化软件源: %w", err)
		}
		return bucket.Put([]byte(b.ID), data)
	})
}

// GetBucket 获取软件源配置。
func (s *boltStorage) GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error) {
	var bucket manifest.BucketConfig
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketsBucket)
		data := b.Get([]byte(name))
		if data == nil {
			return fmt.Errorf("软件源不存在: %s", name)
		}
		return json.Unmarshal(data, &bucket)
	})
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

// DeleteBucket 删除软件源配置。
func (s *boltStorage) DeleteBucket(ctx context.Context, name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketsBucket)
		return b.Delete([]byte(name))
	})
}

// ListBuckets 列出所有软件源配置。
func (s *boltStorage) ListBuckets(ctx context.Context) ([]*manifest.BucketConfig, error) {
	var buckets []*manifest.BucketConfig
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketsBucket)
		return b.ForEach(func(k, v []byte) error {
			var bucket manifest.BucketConfig
			if err := json.Unmarshal(v, &bucket); err != nil {
				return err
			}
			buckets = append(buckets, &bucket)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return buckets, nil
}

// Close 关闭数据库连接。
func (s *boltStorage) Close() error {
	return s.db.Close()
}

func (s *boltStorage) SaveInstallOperation(ctx context.Context, op *InstallOperation) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(operationsBucket)
		data, err := json.Marshal(op)
		if err != nil {
			return fmt.Errorf("序列化操作: %w", err)
		}
		return b.Put([]byte(op.ID), data)
	})
}

func (s *boltStorage) GetInstallOperations(ctx context.Context, installedID string) ([]*InstallOperation, error) {
	var ops []*InstallOperation
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(operationsBucket)
		return b.ForEach(func(k, v []byte) error {
			var op InstallOperation
			if err := json.Unmarshal(v, &op); err != nil {
				return err
			}
			if op.InstalledID == installedID {
				ops = append(ops, &op)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return ops, nil
}

func (s *boltStorage) DeleteInstallOperations(ctx context.Context, installedID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(operationsBucket)
		return b.ForEach(func(k, v []byte) error {
			var op InstallOperation
			if err := json.Unmarshal(v, &op); err != nil {
				return err
			}
			if op.InstalledID == installedID {
				return b.Delete(k)
			}
			return nil
		})
	})
}

func (s *boltStorage) SaveSystemOperation(ctx context.Context, op *SystemOperation) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(systemOperationsBucket)
		data, err := json.Marshal(op)
		if err != nil {
			return fmt.Errorf("序列化系统操作: %w", err)
		}
		return b.Put([]byte(op.ID), data)
	})
}

func (s *boltStorage) GetSystemOperations(ctx context.Context, installedID string) ([]*SystemOperation, error) {
	var ops []*SystemOperation
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(systemOperationsBucket)
		return b.ForEach(func(k, v []byte) error {
			var op SystemOperation
			if err := json.Unmarshal(v, &op); err != nil {
				return err
			}
			if op.InstalledID == installedID {
				ops = append(ops, &op)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return ops, nil
}

func (s *boltStorage) DeleteSystemOperations(ctx context.Context, installedID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(systemOperationsBucket)
		return b.ForEach(func(k, v []byte) error {
			var op SystemOperation
			if err := json.Unmarshal(v, &op); err != nil {
				return err
			}
			if op.InstalledID == installedID {
				return b.Delete(k)
			}
			return nil
		})
	})
}
