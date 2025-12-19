package testpilot

import (
	"reflect"
	"testing"
)

func TestAssertEqual(t *testing.T) {
	type args struct {
		response any
		body     []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test AssertEqual",
			args: args{
				response: map[string]interface{}{
					"key": "value",
				},
				body: []byte(`{"key":"value"}`),
			},
			wantErr: false,
		},
		{
			name: "Test Failure AssertEqual",
			args: args{
				response: map[string]interface{}{
					"key": "value",
				},
				body: []byte(`{"key":"value1"}`),
			},
			wantErr: true,
		},
		{
			name: "Test assert string",
			args: args{
				response: "value",
				body:     []byte(`value`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertEqual(tt.args.response)(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssertEqual() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertPath(t *testing.T) {
	type args[T comparable] struct {
		path   string
		assert func(val T) error
		body   []byte
	}
	type testCase[T comparable] struct {
		name    string
		args    args[T]
		wantErr bool
	}
	tests := []testCase[int]{
		{
			name: "Test AssertPath",
			args: args[int]{
				path:   "key",
				assert: Equal(1),
				body:   []byte(`{"key":1}`),
			},
			wantErr: false,
		},
		{
			name: "Test Failure AssertPath",
			args: args[int]{
				path:   "key",
				assert: Equal(1),
				body:   []byte(`{"key":2}`),
			},
			wantErr: true,
		},
		{
			name: "Test AssertPath with nested object",
			args: args[int]{
				path:   "key.subkey",
				assert: Equal(1),
				body:   []byte(`{"key":{"subkey":1}}`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertPath(tt.args.path, tt.args.assert)(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssertPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_convertToType_String(t *testing.T) {
	t.Run("Test convertToType string", func(t *testing.T) {
		got, err := convertToType[string]("value")
		if err != nil {
			t.Errorf("convertToType() error = %v", err)
		}
		if got != "value" {
			t.Errorf("convertToType() got = %v, want value", got)
		}
	})
	t.Run("Test convertToType struct", func(t *testing.T) {
		type testStruct struct {
			Key string
		}
		got, err := convertToType[testStruct](map[string]any{"Key": "value"})
		if err != nil {
			t.Errorf("convertToType() error = %v", err)
		}
		if !reflect.DeepEqual(got, testStruct{"value"}) {
			t.Errorf("convertToType() got = %v, want value", got)
		}
	})
}

func Test_AssertExists(t *testing.T) {
	t.Run("Test AssertExists", func(t *testing.T) {
		err := AssertExists("key")([]byte(`{"key":1}`))
		if err != nil {
			t.Errorf("AssertExists() error = %v", err)
		}
	})
	t.Run("Test Failure AssertExists", func(t *testing.T) {
		err := AssertExists("key")([]byte(`{"key1":1}`))
		if err == nil {
			t.Errorf("AssertExists() error = %v", err)
		}
	})
}

func Test_Exists(t *testing.T) {
	t.Run("Test Exists", func(t *testing.T) {
		err := Exists()("value")
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		}
	})
	t.Run("Test Failure Exists", func(t *testing.T) {
		err := Exists()("")
		if err == nil {
			t.Errorf("Exists() error = %v", err)
		}
	})
}
