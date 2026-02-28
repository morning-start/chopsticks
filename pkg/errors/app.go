package errors

import (
	"fmt"
)

var (
	ErrAppNotFound          = fmt.Errorf("app not found")
	ErrAppAlreadyInstalled  = fmt.Errorf("app already installed")
	ErrAppNotInstalled      = fmt.Errorf("app not installed")
	ErrVersionNotFound      = fmt.Errorf("version not found")
	ErrVersionAlreadyExists = fmt.Errorf("version already exists")
	ErrDependencyConflict   = fmt.Errorf("dependency conflict")
	ErrDependencyNotFound   = fmt.Errorf("dependency not found")
	ErrDownloadFailed       = fmt.Errorf("download failed")
	ErrChecksumMismatch     = fmt.Errorf("checksum mismatch")
	ErrInstallFailed        = fmt.Errorf("installation failed")
	ErrUninstallFailed      = fmt.Errorf("uninstallation failed")
	ErrUpdateFailed         = fmt.Errorf("update failed")
	ErrScriptFailed         = fmt.Errorf("script execution failed")
	ErrHookFailed           = fmt.Errorf("hook execution failed")
	ErrArchiveExtractFailed = fmt.Errorf("archive extraction failed")
)

func NewAppNotFound(name string) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s", ErrAppNotFound, name),
		"app lookup",
		KindNotFound,
	)
}

func NewAppAlreadyInstalled(name, version string) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s (version: %s)", ErrAppAlreadyInstalled, name, version),
		"install check",
		KindAlreadyExists,
	)
}

func NewAppNotInstalled(name string) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s", ErrAppNotInstalled, name),
		"app lookup",
		KindNotFound,
	)
}

func NewVersionNotFound(app, version string) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s@%s", ErrVersionNotFound, app, version),
		"version lookup",
		KindNotFound,
	)
}

func NewDownloadFailed(url string, err error) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s: %w", ErrDownloadFailed, url, err),
		"download",
		KindNetwork,
	)
}

func NewChecksumMismatch(expected, actual string) error {
	return WrapWithKind(
		fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, expected, actual),
		"checksum verify",
		KindInvalidInput,
	)
}

func NewInstallFailed(name string, err error) error {
	return WrapWithKind(
		fmt.Errorf("%s: %w", name, err),
		"install",
		KindExec,
	)
}

func NewUninstallFailed(name string, err error) error {
	return WrapWithKind(
		fmt.Errorf("%s: %w", name, err),
		"uninstall",
		KindExec,
	)
}

func NewUpdateFailed(name string, err error) error {
	return WrapWithKind(
		fmt.Errorf("%s: %w", name, err),
		"update",
		KindExec,
	)
}

func NewScriptFailed(name string, err error) error {
	return WrapWithKind(
		fmt.Errorf("%s: %w", name, err),
		"script",
		KindExec,
	)
}

func NewHookFailed(hookName string, err error) error {
	return WrapWithKind(
		fmt.Errorf("%s: %w", hookName, err),
		"hook",
		KindExec,
	)
}

func NewArchiveExtractFailed(path string, err error) error {
	return WrapWithKind(
		fmt.Errorf("%s: %w", path, err),
		"extract",
		KindIO,
	)
}

func NewDependencyConflict(name, reason string) error {
	return WrapWithKind(
		fmt.Errorf("%w: %s - %s", ErrDependencyConflict, name, reason),
		"dependency",
		KindInvalidInput,
	)
}
