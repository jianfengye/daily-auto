package main

import (
	"testing"
)

func Test_baiduSearcher(t *testing.T) {
	type args struct {
		keyword string
	}
	tests := []struct {
		name      string
		args      args
		wantItems []Item
		wantErr   bool
	}{
		{
			name: "正常测试百度",
			args: args{
				"golang",
			},
			wantItems: nil,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotItems, err := baiduSearcher(tt.args.keyword)
			if (err != nil) != tt.wantErr {
				t.Errorf("baiduSearcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotItems) > 3 {
				t.Errorf("baiduSearcher() len = %v, want > %v", len(gotItems), 3)
			}
		})
	}
}

func Test_zhihuSearcher(t *testing.T) {
	type args struct {
		keyword string
	}
	tests := []struct {
		name      string
		args      args
		wantItems []Item
		wantErr   bool
	}{
		{
			name:      "正常测试知乎",
			args:      args{
				"golang",
			},
			wantItems: nil,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotItems, err := zhihuSearcher(tt.args.keyword)
			if (err != nil) != tt.wantErr {
				t.Errorf("zhihuSearcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotItems) > 3 {
				t.Errorf("zhihuSearcher() len = %v, want > %v", len(gotItems), 3)
			}
		})
	}
}
