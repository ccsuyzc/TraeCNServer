package pkg

import (
	"TraeCNServer/model"
	"fmt"
	"net/http"
	"strings"
	"time"
"encoding/json"
"regexp"
	"golang.org/x/net/html"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

type Crawler2 struct {
	client    *http.Client
	userAgent string
}

func NewCrawler() *Crawler2 {
	return &Crawler2{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}
}

func (c *Crawler2) SearchArticles(query string) ([]model.Article, error) {
	var results []model.Article

	// 腾讯云开发者社区
	if articles, err := c.searchTencentCloud(query); err == nil {
		results = append(results, articles...)
	}

	// 预留其他平台接口
	// c.searchCSDN(query)
	// c.searchJuejin(query)

	return results, nil
}

func (c *Crawler2) searchTencentCloud(query string) ([]model.Article, error) {
	url := fmt.Sprintf("https://cloud.tencent.com/developer/search/article?q=%s", query)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 防封禁策略
	time.Sleep(1 * time.Second)

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var articles []model.Article
	var parseNode func(*html.Node)
	parseNode = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, a := range n.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "article-list") {
					for child := n.FirstChild; child != nil; child = child.NextSibling {
						if child.Data == "a" {
							// link := getAttr(child, "href")
							// title := getText(child)
							// author := getAuthor(child)

							// articles = append(articles, model.Article{
							// 	Title:    title,
							// 	Content:  "",
							// 	Source:   "腾讯云开发者社区",
							// 	Link:     link,
							// 	Author:   author,
							// 	CreateAt: time.Now(),
							// })
						}
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			parseNode(child)
		}
	}
	parseNode(doc)

	return articles, nil
}

// 辅助函数
func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func getText(n *html.Node) string {
	var text strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(text.String())
}

func getAuthor(n *html.Node) string {
	// 实现作者解析逻辑
	return "腾讯云社区作者"
}

// 预留其他平台方法
func (c *Crawler2) searchCSDN(query string) ([]model.Article, error) {
	return nil, nil
}

func (c *Crawler2) searchJuejin(query string) ([]model.Article, error) {
	return nil, nil
}


// SearchResult 定义一个腾讯云开发者社区的搜索结果的结构体
type SearchResult struct {
	SearchData struct {
		List []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
			Desc  string `json:"desc"`
			ViewNum int    `json:"viewNum"`
			CurrentTime string `json:"currentTime"`
		} `json:"list"`
	} `json:"searchData"`
}

// 爬虫
type demo struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Url   string `json:"url"`
	Desc  string `json:"desc"`
	ViewNum int    `json:"viewNum"`
	CurrentTime string `json:"currentTime"`
}

func CrawlerTx(search string) ([]demo) {
	// search := g.Query("search")
	// if search == "" {
	// 	g.JSON(400, gin.H{
	// 		"code": 400,
	// 		"msg":  "搜索内容不能为空",
	// 	})
	// 	return
	// }

	c := colly.NewCollector(colly.Async(false))
	c.UserAgent = "xy"
	c.AllowURLRevisit = true
	extensions.RandomUserAgent(c) // 使用随机的UserAgent，最好能使用代理。这样就不容易被ban
	extensions.Referer(c)         // 在访问的时候带上Referrer，意思就是这一次点击是从哪个页面产生的
	data := make([]demo, 0, 10)
	c.OnError(func(_ *colly.Response, err error) {
		//log.Println("Something went wrong:", err)

	})

	c.OnResponse(func(r *colly.Response) {
		//fmt.Println("Visited", r.Request.URL)
	})

	c.OnHTML("script:nth-last-child(2)", func(e *colly.HTMLElement) {
		if e.Attr("class") == "" {
			scriptContent := e.Text

			// 使用正则表达式提取一下数组部分
			re := regexp.MustCompile(`\{.*\}`)
			matches := re.FindStringSubmatch(scriptContent)[0]

			// 解析 JSON 字符串
			var result SearchResult
			err := json.Unmarshal([]byte(matches), &result)
			if err != nil {
				// log.Fatal(err)
			}

			// 提取前三个结构体的 id 和 title
			for i, item := range result.SearchData.List {
				if i < 6 {
					fmt.Printf("ID: %d, Title: %s，Desc: %v，ViewNum: %v，Time:%v /n", item.ID, item.Title, item.Desc, item.ViewNum, item.CurrentTime)
					demo1 := demo{
						Id:    item.ID,
						Title: item.Title,
						Url:   fmt.Sprintf("https://cloud.tencent.com/developer/article/%d", item.ID),
						Desc:  item.Desc,
						ViewNum: item.ViewNum,
						CurrentTime: item.CurrentTime,
					}
					data = append(data, demo1)
				}
				if i > 6 {
					break
				}
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {

		//fmt.Println("Finished", r.Request.URL)
	})

	c.Visit("https://cloud.tencent.com/developer/search/article-" + search)
    
	c.Wait()
	fmt.Println(data) // 打印结果，你可以根据需要进行处理或返回给客户端
	return data
	// 爬取完成后发送响应
	// g.JSON(200, gin.H{
	// 	"code": 200,
	// 	"msg":  "爬取完成",
	// 	"data": data, // 假设 data 是全局变量或通过引用传递的
	// })

}