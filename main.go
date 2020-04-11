package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/jianfengye/collection"
	"github.com/spf13/cobra"
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
				Options: []string{"baidu", "zhihu", "wechat", "csdn", "cnblog", "custom"},
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
				log.Println(err)
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
				log.Println(err)
			}

			log.Println("获取了知乎数据条数：", len(ret))
			items = append(items, ret...)
			log.Println("结束获取知乎数据")
		}
		if sitesColl.Contains("csdn") {
			// 获取知乎数据
			log.Println("开始获取csdn数据")

			ret, err := csdnSearcher(keyword)
			if err != nil {
				log.Println(err)
			}
			// 只获取10条
			t := ret[0:10]

			log.Println("获取了csdn数据条数：", len(ret))
			items = append(items, t...)
			log.Println("结束获取csdn数据")
		}

		if sitesColl.Contains("cnblog") {
			// 获取知乎数据
			log.Println("开始获取cnblog数据")

			ret, err := cnblogSearcher(keyword)
			if err != nil {
				log.Println(err)
			}

			log.Println("获取了cnblog数据条数：", len(ret))
			items = append(items, ret...)
			log.Println("结束获取cnblog数据")
		}

		if sitesColl.Contains("wechat") {
			// 获取知乎数据
			log.Println("开始获取微信数据")

			ret, err := wechatSearcher(keyword)
			if err != nil {
				log.Println(err)
			}

			log.Println("获取了微信数据条数：", len(ret))
			items = append(items, ret...)
			log.Println("结束微信知乎数据")
		}

		if sitesColl.Contains("custom") {
			text := ""
			prompt := &survey.Multiline{
				Message: "你可以手动输入一些文章的标题和链接[标题;链接]，一行一个，中间用半角封号隔开",
			}
			survey.AskOne(prompt, &text)

			if text != "" {
				lines := strings.Split(text, "\n")
				for _, line := range lines {
					ss := strings.SplitN(line, ";", 2)
					if len(ss) == 2 {
						it := Item{
							Link:   ss[1][0:len(ss[1])-1],
							Title:  ss[0][1:],
							Source: "custom",
						}
						items = append(items, it)
					}
				}
			}
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
				PageSize: 100,
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
汇总小程序：搜索“全栈神盾局” 可以查看往期每日话题
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
		return items, err
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
	c.OnRequest(func(r *colly.Request) {
		//r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		//r.Headers.Set("accept-encoding", "gzip, deflate, br")
		//r.Headers.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7")
		//r.Headers.Set("cache-control", "no-cache")
		r.Headers.Set("Accept-Encoding", "identity")
		r.Headers.Set("Connection", "close")
	})
	c.OnResponse(func(r *colly.Response) {
		r.Save("/tmp/response")
	})
	err = c.Visit(fmt.Sprintf("https://so.csdn.net/so/search/s.do?q=%s", url.QueryEscape(keyword)))
	if err != nil {
		return items, err
	}
	return items, err
}

func cnblogSearcher(keyword string) (items []Item, err error) {
	c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299"
	c.OnHTML(".searchItemTitle", func(element *colly.HTMLElement) {
		link := element.ChildAttr("a", "href")
		title := element.ChildText("a")
		title = strings.ReplaceAll(title, "<strong>", "")
		title = strings.ReplaceAll(title, "</strong>", "")

		items = append(items, Item{
			Link:   link,
			Title:  title,
			Source: "cnblog",
		})
	})
	c.OnRequest(func(r *colly.Request) {
		//r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		//r.Headers.Set("accept-encoding", "gzip, deflate, br")
		//r.Headers.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7")
		//r.Headers.Set("cache-control", "no-cache")
		r.Headers.Set("cookie", "_ga=GA1.2.134585576.1579195134; __gads=ID=6edef42ac31b7603:T=1579195134:S=ALNI_MZBitAeuX6I9tY2scA2-Ezah2LDsQ; UM_distinctid=1700fd9cc07c83-03397bb01d665f-39617b0f-1aeaa0-1700fd9cc08351; Hm_lvt_ce2db1bd2b24b6516cb2451ef7ff1637=1581341268; _gid=GA1.2.991705243.1582625941; sc_is_visitor_unique=rx11857110.1582625941.80FA5686E37E4F228659E3750CC6093B.1.1.1.1.1.1.1.1.1; __utmc=59123430; __utma=59123430.134585576.1579195134.1582685583.1582694915.2; __utmz=59123430.1582694915.2.2.utmcsr=cnblogs.com|utmccn=(referral)|utmcmd=referral|utmcct=/; __utmt=1; DetectCookieSupport=OK; .AspNetCore.Session=CfDJ8Nf%2BZ6tqUPlNrwu2nvfTJEiEG3QSNH2MXBYABfC88qeJzmqaRbVcjoLB546iGtxkMdes7%2BxcSnxbLCojpUrZ5kz8nV1v2bdVW%2B%2BqfWYwfIyHR2sn8b2UJA0OMX4KN4WTtSNOzG1kFCcIXnMnzwMu%2Bu1qBQxFdQvLd%2B9K4Y3W6nB2; ShitNoRobotCookie=CfDJ8Nf-Z6tqUPlNrwu2nvfTJEiSfuovH4SldpdXZrkoNeUlBTosrljIbFzMnB2MWOCFHaos5lmK-SOaHcZRi0ZecIe-PD8Yd-sxJXSE76EyW3PQwlgzOdvtOSc_W960QWwZVw; __utmb=59123430.7.10.1582694915")
		r.Headers.Set("Accept-Encoding", "identity")
		r.Headers.Set("Connection", "close")
	})
	c.OnResponse(func(r *colly.Response) {
		r.Save("/tmp/response")
	})
	err = c.Visit(fmt.Sprintf("https://zzk.cnblogs.com/s/blogpost?w=%s", url.QueryEscape(keyword)))
	if err != nil {
		return items, err
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
					log.Println(err)
				}
			} else {
				log.Println(err)
			}
		}

		items = append(items, Item{
			Link:   realUrl.String(),
			Title:  baiduItem.Title,
			Source: "baidu",
		})
	})
	c.OnRequest(func(r *colly.Request) {
		//r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		//r.Headers.Set("accept-encoding", "gzip, deflate, br")
		//r.Headers.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7")
		//r.Headers.Set("cache-control", "no-cache")
		r.Headers.Set("Cookie", "BIDUPSID=561E3AA4D70CC5469689D5FC799ED8A4; PSTM=1579187292; BAIDUID=561E3AA4D70CC5469A789A8439E471FC:FG=1; BD_UPN=123253; MCITY=-315%3A; BDUSS=R3cHo1bkhYaldyZ1FFWjdaeElRdFJ0dHR0ZU1VaXlxWnM4S1lyaExXUkZaNUplRVFBQUFBJCQAAAAAAAAAAAEAAAAiFVUSeWVqaWFuZmVuZ25pY2sAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEXaal5F2mpeMm; BDORZ=B490B5EBF6F3CD402E515D22BCDA1598; delPer=0; BD_CK_SAM=1; PSINO=2; BD_HOME=1; H_PS_PSSID=30968_1432_21085_30794_30901_30996_31051_30823_31085; COOKIE_SESSION=7401_0_7_7_7_14_0_1_7_6_0_0_7400_0_6_0_1584347003_0_1584346997%7C9%23491372_21_1583481420%7C9; H_PS_645EC=8d02yQttoTNCqJMYhCikoYcIUfES5pPuFeyDtzzT2e48XrJ7BLJjYZj4JeM")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")

	})
	c.OnResponse(func(r *colly.Response) {
		r.Save("/tmp/response.baidu")
	})
	err = c.Visit(fmt.Sprintf("https://www.baidu.com/s?wd=%s", url.QueryEscape(keyword)))
	if err != nil {
		return items, err
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
			return
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
		return items, err
	}
	return items, err
}

func noescape(str string) template.HTML {
	return template.HTML(str)
}

var fn = template.FuncMap{
	"noescape": noescape,
}
