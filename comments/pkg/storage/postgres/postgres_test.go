package postgres

import (
	"reflect"
	"testing"
	"GoNews/comments/pkg/storage"
)

func TestStorage_Comments(t *testing.T) {
	type args struct {
		n int64
	}
	tests := []struct {
		name    string
		s       *Storage
		args    args
		want    []comStorage.Comment
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Comments(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Comments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.Comments() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_AddComments(t *testing.T) {
	type args struct {
		comments []comStorage.Comment
	}
	tests := []struct {
		name    string
		s       *Storage
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.AddComments(tt.args.comments, ""); (err != nil) != tt.wantErr {
				t.Errorf("Storage.AddComments() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
