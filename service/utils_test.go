package service

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mapFromSlice(t *testing.T) {
	res := mapFromSlice([]string{"abc", "123", "a1b2c3"})
	_, ok := res["123"]
	assert.True(t, ok)
	_, ok = res["abc"]
	assert.True(t, ok)
	_, ok = res["a1b2c3"]
	assert.True(t, ok)
	_, ok = res["zxc"]
	assert.False(t, ok)
}

func Test_isOwner(t *testing.T) {
	type args struct {
		user   string
		owners []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "owner",
			args: args{user: "userA", owners: []string{"userB", "userC", "userA", "userD"}},
			want: true,
		},
		{
			name: "not_owner",
			args: args{user: "userA", owners: []string{"userB", "userC", "userD"}},
			want: false,
		},
		{
			name: "empty owners",
			args: args{user: "userA", owners: nil},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOwner(tt.args.user, tt.args.owners); got != tt.want {
				t.Errorf("isOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getOwnersFromAnnotations(t *testing.T) {
	type args struct {
		annotations map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "exists",
			args: args{
				annotations: map[string]string{
					ownersKey: "userA,userB,userC",
				},
			},
			want: []string{"userA", "userB", "userC"},
		},
		{
			name: "exists and trim",
			args: args{
				annotations: map[string]string{
					ownersKey: "userA, userB, userC",
				},
			},
			want: []string{"userA", "userB", "userC"},
		},
		{
			name: "exists and trim one item",
			args: args{
				annotations: map[string]string{
					ownersKey: "userA",
				},
			},
			want: []string{"userA"},
		},
		{
			name: "empty annotations",
			args: args{
				annotations: map[string]string{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOwnersFromAnnotations(tt.args.annotations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOwnersFromAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeRepeatedItems(t *testing.T) {
	type args struct {
		items []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "abc",
			args: struct{ items []string }{
				items: []string{"a", "b", "c"},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "a",
			args: struct{ items []string }{
				items: []string{"a"},
			},
			want: []string{"a"},
		},
		{
			name: "abcaccb",
			args: struct{ items []string }{
				items: []string{"a", "b", "c", "a", "c", "c", "b"},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "empty",
			args: struct{ items []string }{
				items: []string{},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeRepeatedItems(tt.args.items); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeRepeatedItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeItemFromItems(t *testing.T) {
	type args struct {
		item  string
		items []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "a_abacda",
			args: args{
				item:  "a",
				items: []string{"a", "b", "a", "c", "d", "a"},
			},
			want: []string{"b", "c", "d"},
		},
		{
			name: "a_aaa",
			args: args{
				item:  "a",
				items: []string{"a", "a", "a"},
			},
			want: []string{},
		},
		{
			name: "empty",
			args: args{
				item:  "a",
				items: nil,
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeItemFromItems(tt.args.item, tt.args.items); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeItemFromItems() = %v, want %v", got, tt.want)
			}
		})
	}
}
