package errors

import (
	"fmt"
)

var (
	ErrBucketNotFound      = fmt.Errorf("bucket not found")
	ErrBucketAlreadyExists = fmt.Errorf("bucket already exists")
	ErrInvalidBucketURL    = fmt.Errorf("invalid bucket URL")
	ErrBucketLoadFailed    = fmt.Errorf("bucket load failed")
	ErrBucketUpdateFailed  = fmt.Errorf("bucket update failed")
	ErrAppManifestNotFound = fmt.Errorf("app manifest not found")
)

func NewBucketNotFound(name string) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s", ErrBucketNotFound, name),
		"bucket lookup",
		KindNotFound,
	)
}

func NewBucketAlreadyExists(name string) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s", ErrBucketAlreadyExists, name),
		"bucket add",
		KindAlreadyExists,
	)
}

func NewInvalidBucketURL(url string) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s", ErrInvalidBucketURL, url),
		"bucket validate",
		KindInvalidInput,
	)
}

func NewBucketLoadFailed(name string, err error) error {
	return WrapWithKind(
		fmt.Errorf("%s: %w", name, err),
		"bucket load",
		KindIO,
	)
}

func NewBucketUpdateFailed(name string, err error) error {
	return WrapWithKind(
		fmt.Errorf("%s: %w", name, err),
		"bucket update",
		KindNetwork,
	)
}

func NewAppManifestNotFound(bucket, app string) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s/%s", ErrAppManifestNotFound, bucket, app),
		"manifest lookup",
		KindNotFound,
	)
}
