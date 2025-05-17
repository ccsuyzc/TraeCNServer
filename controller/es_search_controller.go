package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"TraeCNServer/model"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

//  ElasticSearch 客户端
var esClient *elastic.Client

// 初始化 ElasticSearch 客户端
func InitESClient(url string) error {
	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		return err
	}
	esClient = client
	return nil
}

// /es_search/users?q=xxx 查询用户
func ESSearchUsers(c *gin.Context) {
	query := c.Query("q")
	if esClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ElasticSearch client not initialized"})
		return
	}
	q := elastic.NewMultiMatchQuery(query, "username", "email", "personal_signature", "personal_intro")
	searchResult, err := esClient.Search().
		Index("users").
		Query(q).
		Do(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	users := make([]model.ESUser, 0)
	for _, hit := range searchResult.Hits.Hits {
		var user model.ESUser
		if err := json.Unmarshal(hit.Source, &user); err == nil {
			users = append(users, user)
		}
	}
	c.JSON(http.StatusOK, users)
}

// /es_search/articles?q=xxx 查询文章
func ESSearchArticles(c *gin.Context) {
	// 从请求参数中获取查询关键字
	query := c.Query("q")
	if esClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ElasticSearch client not initialized"})
		return
	}
	// 使用 ElasticSearch 进行全文搜索
	q := elastic.NewMultiMatchQuery(query, "title", "description", "content", "tags")
	searchResult, err := esClient.Search().
		Index("articles").
		Query(q).
		Do(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 处理搜索结果
	articles := make([]model.ESArticle, 0)
	for _, hit := range searchResult.Hits.Hits {
		var article model.ESArticle
		if err := json.Unmarshal(hit.Source, &article); err == nil {
			articles = append(articles, article)
		}
	}

	// 返回搜索结果
	c.JSON(http.StatusOK, articles)
}

// /es_search/tags?q=xxx 查询标签
func ESSearchTags(c *gin.Context) {
	query := c.Query("q")
	if esClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ElasticSearch client not initialized"})
		return
	}
	q := elastic.NewMultiMatchQuery(query, "name")
	searchResult, err := esClient.Search().
		Index("tags").
		Query(q).
		Do(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tags := make([]model.ESTag, 0)
	for _, hit := range searchResult.Hits.Hits {
		var tag model.ESTag
		if err := json.Unmarshal(hit.Source, &tag); err == nil {
			tags = append(tags, tag)
		}
	}
	c.JSON(http.StatusOK, tags)
}
