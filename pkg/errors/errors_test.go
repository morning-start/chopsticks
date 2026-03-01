package errors

import (
	"errors"
	"testing"
)

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrAlreadyExists", ErrAlreadyExists, "already exists"},
		{"ErrInvalidInput", ErrInvalidInput, "invalid input"},
		{"ErrPermissionDenied", ErrPermissionDenied, "permission denied"},
		{"ErrCancelled", ErrCancelled, "operation cancelled"},
		{"ErrTimeout", ErrTimeout, "operation timeout"},
		{"ErrNotSupported", ErrNotSupported, "not supported"},
		{"ErrInternal", ErrInternal, "internal error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("%s.Error() = %v, want %v", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		kind    ErrorKind
		msg     string
		wantErr bool
		wantMsg string
	}{
		{
			name:    "create not found error",
			kind:    KindNotFound,
			msg:     "file not found",
			wantErr: false,
			wantMsg: "file not found",
		},
		{
			name:    "create invalid input error",
			kind:    KindInvalidInput,
			msg:     "invalid parameter",
			wantErr: false,
			wantMsg: "invalid parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.kind, tt.msg)
			if err == nil {
				t.Fatal("New() returned nil")
			}

			e, ok := err.(*Error)
			if !ok {
				t.Fatal("New() did not return *Error type")
			}

			if e.Kind != tt.kind {
				t.Errorf("Kind = %v, want %v", e.Kind, tt.kind)
			}

			if e.Error() != tt.wantMsg {
				t.Errorf("Error() = %v, want %v", e.Error(), tt.wantMsg)
			}
		})
	}
}

func TestNewf(t *testing.T) {
	tests := []struct {
		name    string
		kind    ErrorKind
		format  string
		args    []interface{}
		wantMsg string
	}{
		{
			name:    "format with string",
			kind:    KindNotFound,
			format:  "file %s not found",
			args:    []interface{}{"test.txt"},
			wantMsg: "file test.txt not found",
		},
		{
			name:    "format with number",
			kind:    KindInvalidInput,
			format:  "expected %d, got %d",
			args:    []interface{}{1, 2},
			wantMsg: "expected 1, got 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Newf(tt.kind, tt.format, tt.args...)
			if err == nil {
				t.Fatal("Newf() returned nil")
			}

			e, ok := err.(*Error)
			if !ok {
				t.Fatal("Newf() did not return *Error type")
			}

			if e.Kind != tt.kind {
				t.Errorf("Kind = %v, want %v", e.Kind, tt.kind)
			}

			if e.Error() != tt.wantMsg {
				t.Errorf("Error() = %v, want %v", e.Error(), tt.wantMsg)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		op      string
		wantNil bool
		wantMsg string
	}{
		{
			name:    "wrap error with operation",
			err:     errors.New("original error"),
			op:      "read file",
			wantNil: false,
			wantMsg: "read file: original error",
		},
		{
			name:    "wrap nil error",
			err:     nil,
			op:      "read file",
			wantNil: true,
		},
		{
			name:    "wrap without operation",
			err:     errors.New("original error"),
			op:      "",
			wantNil: false,
			wantMsg: "original error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.err, tt.op)
			if tt.wantNil {
				if err != nil {
					t.Errorf("Wrap() = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatal("Wrap() returned nil")
			}

			e, ok := err.(*Error)
			if !ok {
				t.Fatal("Wrap() did not return *Error type")
			}

			if e.Op != tt.op {
				t.Errorf("Op = %v, want %v", e.Op, tt.op)
			}

			if e.Error() != tt.wantMsg {
				t.Errorf("Error() = %v, want %v", e.Error(), tt.wantMsg)
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		format  string
		args    []interface{}
		wantNil bool
		wantMsg string
	}{
		{
			name:    "wrapf with format",
			err:     errors.New("original error"),
			format:  "read %s",
			args:    []interface{}{"test.txt"},
			wantNil: false,
			wantMsg: "read test.txt: original error",
		},
		{
			name:    "wrapf nil error",
			err:     nil,
			format:  "read %s",
			args:    []interface{}{"test.txt"},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrapf(tt.err, tt.format, tt.args...)
			if tt.wantNil {
				if err != nil {
					t.Errorf("Wrapf() = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatal("Wrapf() returned nil")
			}

			if err.Error() != tt.wantMsg {
				t.Errorf("Error() = %v, want %v", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestWrapWithKind(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		op       string
		kind     ErrorKind
		wantNil  bool
		wantKind ErrorKind
	}{
		{
			name:     "wrap with kind",
			err:      errors.New("original error"),
			op:       "read file",
			kind:     KindNotFound,
			wantNil:  false,
			wantKind: KindNotFound,
		},
		{
			name:    "wrap nil with kind",
			err:     nil,
			op:      "read file",
			kind:    KindNotFound,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WrapWithKind(tt.err, tt.op, tt.kind)
			if tt.wantNil {
				if err != nil {
					t.Errorf("WrapWithKind() = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatal("WrapWithKind() returned nil")
			}

			e, ok := err.(*Error)
			if !ok {
				t.Fatal("WrapWithKind() did not return *Error type")
			}

			if e.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", e.Kind, tt.wantKind)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, "operation")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "unwrap wrapped error",
			err:  wrapped,
			want: original,
		},
		{
			name: "unwrap nil",
			err:  nil,
			want: nil,
		},
		{
			name: "unwrap standard error",
			err:  original,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unwrap(tt.err)
			if got != tt.want {
				t.Errorf("Unwrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "is same error",
			err:    ErrNotFound,
			target: ErrNotFound,
			want:   true,
		},
		{
			name:   "is different error",
			err:    ErrNotFound,
			target: ErrAlreadyExists,
			want:   false,
		},
		{
			name:   "is wrapped error",
			err:    Wrap(ErrNotFound, "operation"),
			target: ErrNotFound,
			want:   true,
		},
		{
			name:   "is nil",
			err:    nil,
			target: ErrNotFound,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Is(tt.err, tt.target)
			if got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAs(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "as Error type",
			err:  New(KindNotFound, "test"),
			want: true,
		},
		{
			name: "as wrapped Error",
			err:  Wrap(New(KindNotFound, "test"), "op"),
			want: true,
		},
		{
			name: "as different type",
			err:  errors.New("standard error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target *Error
			got := As(tt.err, &target)
			if got != tt.want {
				t.Errorf("As() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKind(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorKind
	}{
		{
			name: "get kind from Error",
			err:  New(KindNotFound, "test"),
			want: KindNotFound,
		},
		{
			name: "get kind from wrapped Error",
			err:  WrapWithKind(errors.New("test"), "op", KindInvalidInput),
			want: KindInvalidInput,
		},
		{
			name: "get kind from nil",
			err:  nil,
			want: KindUnknown,
		},
		{
			name: "get kind from standard error",
			err:  errors.New("standard error"),
			want: KindUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetKind(tt.err)
			if got != tt.want {
				t.Errorf("GetKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsKind(t *testing.T) {
	tests := []struct {
		name string
		err  error
		kind ErrorKind
		want bool
	}{
		{
			name: "is kind match",
			err:  New(KindNotFound, "test"),
			kind: KindNotFound,
			want: true,
		},
		{
			name: "is kind not match",
			err:  New(KindNotFound, "test"),
			kind: KindInvalidInput,
			want: false,
		},
		{
			name: "is kind from nil",
			err:  nil,
			kind: KindNotFound,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsKind(tt.err, tt.kind)
			if got != tt.want {
				t.Errorf("IsKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorKindConstants(t *testing.T) {
	// Test that ErrorKind constants have expected values
	if KindUnknown != 0 {
		t.Errorf("KindUnknown = %d, want 0", KindUnknown)
	}
	if KindNotFound != 1 {
		t.Errorf("KindNotFound = %d, want 1", KindNotFound)
	}
	if KindAlreadyExists != 2 {
		t.Errorf("KindAlreadyExists = %d, want 2", KindAlreadyExists)
	}
	if KindInvalidInput != 3 {
		t.Errorf("KindInvalidInput = %d, want 3", KindInvalidInput)
	}
	if KindPermission != 4 {
		t.Errorf("KindPermission = %d, want 4", KindPermission)
	}
	if KindNetwork != 5 {
		t.Errorf("KindNetwork = %d, want 5", KindNetwork)
	}
	if KindIO != 6 {
		t.Errorf("KindIO = %d, want 6", KindIO)
	}
	if KindExec != 7 {
		t.Errorf("KindExec = %d, want 7", KindExec)
	}
	if KindCancelled != 8 {
		t.Errorf("KindCancelled = %d, want 8", KindCancelled)
	}
	if KindTimeout != 9 {
		t.Errorf("KindTimeout = %d, want 9", KindTimeout)
	}
}
