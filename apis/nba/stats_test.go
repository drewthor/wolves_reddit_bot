package nba

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func Test_parseRowSetValue(t *testing.T) {
	t.Run("string parsing", func(t *testing.T) {
		header := "key"
		value := "value"
		valueJSON, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("failed to marshal value to JSON with err: %v", err)
		}
		headersMap := map[string]int{header: 0}
		validRowSet := []json.RawMessage{valueJSON}
		invalidRowset := []json.RawMessage{[]byte("3.1")}

		type args struct {
			headersMap map[string]int
			rowSet     []json.RawMessage
			header     string
		}
		type testCase[T any] struct {
			name    string
			args    args
			want    T
			wantErr bool
		}
		tests := []testCase[string]{
			{
				name: "valid string succeeds",
				args: args{
					headersMap: headersMap,
					rowSet:     validRowSet,
					header:     header,
				},
				want:    value,
				wantErr: false,
			},
			{
				name: "invalid string fails",
				args: args{
					headersMap: headersMap,
					rowSet:     invalidRowset,
					header:     header,
				},
				want:    value,
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := parseRowSetValue[string](tt.args.headersMap, tt.args.rowSet, tt.args.header)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseRowSetValue() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err == nil && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseRowSetValue() got = %v, want %v", got, tt.want)
				}
			})
		}
	})
	t.Run("float64 parsing", func(t *testing.T) {
		header := "key"
		value := 4.2
		headersMap := map[string]int{header: 0}
		validRowSet := []json.RawMessage{[]byte(fmt.Sprintf("%g", value))}
		invalidRowset := []json.RawMessage{[]byte("value")}

		type args struct {
			headersMap map[string]int
			rowSet     []json.RawMessage
			header     string
		}
		type testCase[T any] struct {
			name    string
			args    args
			want    T
			wantErr bool
		}
		tests := []testCase[float64]{
			{
				name: "valid float64 succeeds",
				args: args{
					headersMap: headersMap,
					rowSet:     validRowSet,
					header:     header,
				},
				want:    value,
				wantErr: false,
			},
			{
				name: "invalid float64 fails",
				args: args{
					headersMap: headersMap,
					rowSet:     invalidRowset,
					header:     header,
				},
				want:    value,
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := parseRowSetValue[float64](tt.args.headersMap, tt.args.rowSet, tt.args.header)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseRowSetValue() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err == nil && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseRowSetValue() got = %v, want %v", got, tt.want)
				}
			})
		}
	})
	t.Run("string pointer parsing", func(t *testing.T) {
		header := "key"
		nullHeader := "null"
		value := "value"
		valueJSON, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("failed to marshal value to JSON with err: %v", err)
		}
		var nilValue *string
		headersMap := map[string]int{header: 0, nullHeader: 1}
		validRowSet := []json.RawMessage{valueJSON, []byte(`null`)}
		invalidRowset := []json.RawMessage{[]byte(fmt.Sprintf("%g", 3.2))}

		type args struct {
			headersMap map[string]int
			rowSet     []json.RawMessage
			header     string
		}
		type testCase[T any] struct {
			name    string
			args    args
			want    T
			wantErr bool
		}
		tests := []testCase[*string]{
			{
				name: "valid string succeeds",
				args: args{
					headersMap: headersMap,
					rowSet:     validRowSet,
					header:     header,
				},
				want:    &value,
				wantErr: false,
			},
			{
				name: "null string succeeds",
				args: args{
					headersMap: headersMap,
					rowSet:     validRowSet,
					header:     nullHeader,
				},
				want:    nilValue,
				wantErr: false,
			},
			{
				name: "invalid string fails",
				args: args{
					headersMap: headersMap,
					rowSet:     invalidRowset,
					header:     header,
				},
				want:    nil,
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := parseRowSetValue[*string](tt.args.headersMap, tt.args.rowSet, tt.args.header)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseRowSetValue() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err == nil && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseRowSetValue() got = %v, want %v", got, tt.want)
				}
			})
		}
	})
}
