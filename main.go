package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/jianfengye/collection"
	"github.com/spf13/cobra"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	if err := flowCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type Item struct {
	Link   string // 链接
	Title  string // 标题
	Source string // 来源
}

var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "生成工具",
	Run: func(cmd *cobra.Command, args []string) {
		var keyword string
		var sites []string

		// 输入关键字
		{
			prompt := &survey.Input{
				Message: "请输入关键字：",
			}
			err := survey.AskOne(prompt, &keyword)
			if err != nil {
				panic(err)
			}
		}

		// 选择多个网站
		{
			prompt := &survey.MultiSelect{
				Message: "请选择搜索的网站：",
				Options: []string{"baidu", "zhihu", "wechat", "csdn"},
			}
			err := survey.AskOne(prompt, &sites)
			if err != nil {
				panic(err)
			}
		}

		items := make([]Item, 0)

		// 进行网站搜索
		sitesColl := collection.NewStrCollection(sites)
		if sitesColl.Contains("baidu") {
			// 获取百度数据
			log.Println("开始获取百度数据")

			ret, err := baiduSearcher(keyword)
			if err != nil {
				log.Panic(err)
			}

			log.Println("获取了百度数据条数：", len(ret))
			items = append(items, ret...)

			log.Println("结束获取百度数据")
		}

		if sitesColl.Contains("zhihu") {
			// 获取知乎数据
			log.Println("开始获取知乎数据")

			ret, err := zhihuSearcher(keyword)
			if err != nil {
				log.Panic(err)
			}

			log.Println("获取了知乎数据条数：", len(ret))
			items = append(items, ret...)
			log.Println("结束获取知乎数据")
		}
		if sitesColl.Contains("csdn") {
			// 获取知乎数据
			log.Println("开始获取知乎数据")

			ret, err := csdnSearcher(keyword)
			if err != nil {
				log.Panic(err)
			}

			log.Println("获取了csdn数据条数：", len(ret))
			items = append(items, ret...)
			log.Println("结束获取csdn数据")
		}

		if sitesColl.Contains("wechat") {
			// 获取知乎数据
			log.Println("开始获取微信数据")

			ret, err := wechatSearcher(keyword)
			if err != nil {
				log.Panic(err)
			}

			log.Println("获取了微信数据条数：", len(ret))
			items = append(items, ret...)
			log.Println("结束微信知乎数据")
		}

		// 网站搜索结束
		// 提示搜索成功
		log.Println("搜索成功，一共获得数据：", len(items))
		opts := []string{}
		for _, item := range items {
			fmt.Println(item.Title, "   ", item.Link)
			opts = append(opts, item.Title+" "+item.Link)
		}
	START_SELECT:
		// 提示进行选择
		selected := []string{}
		{
			prompt := &survey.MultiSelect{
				Message:  "请选择几条作为今日的知识点：",
				Options:  opts,
				PageSize: 20,
			}
			err := survey.AskOne(prompt, &selected)
			if err != nil {
				log.Panic(err)
			}
		}

		// 确定是否要生成知识点
		ok := false
		{
			prompt := &survey.Confirm{
				Message: "确认选择这些吗？No请重新选择",
			}
			survey.AskOne(prompt, &ok)
		}

		if !ok {
			goto START_SELECT
		}

		selectItems := make([]Item, 0)
		for _, s := range selected {
			i := strings.LastIndex(s, " ")
			a := Item{
				Link:  s[i:],
				Title: s[:i],
			}
			selectItems = append(selectItems, a)
		}

		ok = false
		{
			prompt := &survey.Confirm{
				Message: "是否要生成每日话题？",
			}
			survey.AskOne(prompt, &ok)
		}

		if ok {
			var author string
			{
				prompt := &survey.Input{
					Message: "请输入编辑者名称：",
				}
				err := survey.AskOne(prompt, &author)
				if err != nil {
					panic(err)
				}
			}
			outputDaily(keyword, selectItems, author)
		}

		// 生成插入mysql需要的content
		ok = false
		{
			prompt := &survey.Confirm{
				Message: "是否要生成每日话题的sql content？",
			}
			survey.AskOne(prompt, &ok)
		}

		if ok {
			outputSqlContent(selectItems)
		}
	},
}

func outputSqlContent(selectItems []Item) {
	type SqlItem struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Comment string `json:"comment"`
	}
	rets := make([]SqlItem, 0)
	for _, item := range selectItems {
		rets = append(rets, SqlItem{
			Title:   item.Title,
			Link:    item.Link,
			Comment: "",
		})
	}
	out, err := json.Marshal(rets)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(out))
}

// 生成每日话题
func outputDaily(keyword string, selectItems []Item, author string) {

	// 生成需要的结构
	struc := struct {
		KeyWord string
		Items   []Item
		Author  string
	}{
		KeyWord: keyword,
		Items:   selectItems,
		Author:  author,
	}
	tmpl := `
今日话题： {{.KeyWord}}

{{range .Items}}
{{.Title}} {{.Link|noescape}}
{{end}}

编辑：{{.Author}}
汇总地址：http://www.huoding.com/#/
			`
	t, err := template.New("daily").Funcs(fn).Parse(tmpl)
	if err != nil {
		log.Panic(err)
	}
	var out bytes.Buffer
	err = t.Execute(&out, struc)
	if err != nil {
		log.Panic(err)
	}
	fmt.Print(out.String())
}

func zhihuSearcher(keyword string) (items []Item, err error) {
	c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299"
	c.OnHTML(".ContentItem-title", func(element *colly.HTMLElement) {
		link := element.ChildAttr("a", "href")
		title := element.ChildText("span")
		title = strings.ReplaceAll(title, "<em>", "")
		title = strings.ReplaceAll(title, "</em>", "")

		if strings.HasPrefix(link, "//") {
			link = "https:" + link
		} else if strings.HasPrefix(link, "/") {
			link = "https://www.zhihu.com" + link
		}

		items = append(items, Item{
			Link:   link,
			Title:  title,
			Source: "zhihu",
		})
	})
	err = c.Visit(fmt.Sprintf("https://www.zhihu.com/search?type=content&q=%s", url.QueryEscape(keyword)))
	if err != nil {
		log.Panic(err)
	}
	return items, err
}

func csdnSearcher(keyword string) (items []Item, err error) {
	c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299"
	c.OnHTML(".limit_width", func(element *colly.HTMLElement) {
		link := element.ChildAttr("a", "href")
		title := element.ChildText("a")
		title = strings.ReplaceAll(title, "<em>", "")
		title = strings.ReplaceAll(title, "</em>", "")

		items = append(items, Item{
			Link:   link,
			Title:  title,
			Source: "csdn",
		})
	})
	err = c.Visit(fmt.Sprintf("https://so.csdn.net/so/search/s.do?q=%s", url.QueryEscape(keyword)))
	if err != nil {
		log.Panic(err)
	}
	return items, err
}

// 搜索百度, 获取第一页
func baiduSearcher(keyword string) (items []Item, err error) {
	c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299"
	type BaiduItem struct {
		Title string
		Url   string
	}
	c.OnHTML(".c-tools", func(element *colly.HTMLElement) {
		jsonData := element.Attr("data-tools")
		if jsonData == "" {
			log.Println("获取data-tool失败")
			return
		}

		var baiduItem BaiduItem
		err := json.Unmarshal([]byte(jsonData), &baiduItem)
		if err != nil {
			log.Println(err)
			return
		}
		// 转换baidu的短链接
		client := new(http.Client)
		req, err := http.NewRequest("GET", baiduItem.Url, nil)
		if err != nil {
			log.Println(err)
			return
		}

		realUrl := req.URL
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return errors.New("Redirect")
		}
		resp, err := client.Do(req)
		if err != nil {
			if resp.StatusCode == http.StatusFound ||
				resp.StatusCode == http.StatusMovedPermanently {
				realUrl, err = resp.Location()
				if err != nil {
					log.Panic(err)
				}
			} else {
				log.Panic(err)
			}
		}

		items = append(items, Item{
			Link:   realUrl.String(),
			Title:  baiduItem.Title,
			Source: "baidu",
		})
	})
	err = c.Visit(fmt.Sprintf("http://www.baidu.com/s?wd=%s", url.QueryEscape(keyword)))
	if err != nil {
		return nil, err
	}
	return items, err
}

// 搜索sougouweixin, 获取第一页
func wechatSearcher(keyword string) (items []Item, err error) {
	c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299"
	c1 := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))
	c1.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299"

	c.OnHTML("h3", func(element *colly.HTMLElement) {
		//link := element.ChildAttr("a", "href")
		title := element.ChildText("a")
		title = strings.ReplaceAll(title, "<em>", "")
		title = strings.ReplaceAll(title, "</em>", "")
		title = strings.ReplaceAll(title, "<!--red_beg-->", "")
		title = strings.ReplaceAll(title, "<!--red_end-->", "")

		// 再去sogou主搜索查询
		err := c1.Visit(fmt.Sprintf("https://www.sogou.com/web?query=%s", url.QueryEscape(title)))
		if err != nil {
			log.Panic(err)
		}
	})

	c1.OnHTML(".tit-ico", func(element *colly.HTMLElement) {
		link := element.Attr("href")
		title := element.DOM.Parent().Prev().Text()
		title = strings.ReplaceAll(title, "\n", "")
		title = strings.ReplaceAll(title, "<em>", "")
		title = strings.ReplaceAll(title, "</em>", "")
		title = strings.ReplaceAll(title, "<!--red_beg-->", "")
		title = strings.ReplaceAll(title, "<!--red_end-->", "")

		items = append(items, Item{
			Link:   link,
			Title:  title,
			Source: "wechat",
		})
	})

	err = c.Visit(fmt.Sprintf("https://weixin.sogou.com/weixin?query=%s&type=2&ie=utf8", url.QueryEscape(keyword)))
	if err != nil {
		log.Panic(err)
	}
	return items, err
}

func noescape(str string) template.HTML {
	return template.HTML(str)
}

var fn = template.FuncMap{
	"noescape": noescape,
}
