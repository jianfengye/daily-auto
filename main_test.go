package main

import (
	"bytes"
	"html/template"
	"log"
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
			name: "正常测试知乎",
			args: args{
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
func Test_csdnSearcher(t *testing.T) {
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
			name: "正常测试知乎",
			args: args{
				"golang",
			},
			wantItems: nil,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotItems, err := csdnSearcher(tt.args.keyword)
			if (err != nil) != tt.wantErr {
				t.Errorf("csdnSearcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotItems) < 3 {
				t.Errorf("csdnSearcher() len = %v, want > %v", len(gotItems), 3)
			}
		})
	}
}

func Test_weichatSearcher(t *testing.T) {
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
			name: "正常测试微信",
			args: args{
				"golang",
			},
			wantItems: nil,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotItems, err := wechatSearcher(tt.args.keyword)
			if (err != nil) != tt.wantErr {
				t.Errorf("wechatSearcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotItems) < 3 {
				t.Errorf("wechatSearcher() len = %v, want > %v", len(gotItems), 3)
			}
		})
	}
}

func Test_UrlOutput(t *testing.T) {

	type Item struct {
		Link   string // 链接
		Title  string // 标题
		Source string // 来源
	}
	struc := struct {
		KeyWord string
		Items   []Item
		Author  string
	}{
		KeyWord: "你好",
		Items: []Item{
			{
				Link:   "http://baidu.com?a=1&b=2",
				Title:  "标题",
				Source: "",
			},
		},
		Author: "轩脉刃",
	}

	tmpl := `
今日话题： {{.KeyWord}}

{{range .Items}}
{{.Title}} {{.Link|noescape}}
{{end}}

编辑：{{.Author}}
汇总地址：http://www.huoding.com/#/
`

	t2, err := template.New("daily").Funcs(fn).Parse(tmpl)
	if err != nil {
		log.Panic(err)
	}
	var out bytes.Buffer
	err = t2.Execute(&out, struc)
	if err != nil {
		log.Panic(err)
	}
	panic(out.String())

}
