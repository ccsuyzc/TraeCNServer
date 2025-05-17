package service

import (
	"TraeCNServer/model"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"gorm.io/gorm"
)

type RecommendService struct {
	DB *gorm.DB
}

// 获取用户行为数据
func (s *RecommendService) GetUserBehavior(ctx context.Context, userID uint) ([]model.Favorite, []model.SearchHistory, []model.ReadingHistory) {
	type result struct {
		favorites []model.Favorite
		searches  []model.SearchHistory
		readings  []model.ReadingHistory
		err       error
	}

	resChan := make(chan result, 1)

	go func() {
		defer close(resChan)
		res := result{}

		// 并行查询三个数据源
		var wg sync.WaitGroup
		wg.Add(3)

		// 查询收藏记录
		go func() {
			defer wg.Done()
			err := s.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&res.favorites).Error
			if err != nil {
				res.err = fmt.Errorf("查询收藏记录失败: %w", err)
			}
		}()

		// 查询搜索记录
		go func() {
			defer wg.Done()
			err := s.DB.WithContext(ctx).Where("user_id = ?", userID).
				Order("timestamp desc").
				Limit(50).
				Find(&res.searches).Error
			if err != nil {
				res.err = fmt.Errorf("查询搜索记录失败: %w", err)
			}
		}()

		// 查询阅读记录
		go func() {
			defer wg.Done()
			err := s.DB.WithContext(ctx).Where("user_id = ?", userID).
				Order("timestamp desc").
				Limit(100).
				Find(&res.readings).Error
			if err != nil {
				res.err = fmt.Errorf("查询阅读记录失败: %w", err)
			}
		}()

		wg.Wait()
		resChan <- res
	}()

	select {
	case <-ctx.Done():
		return nil, nil, nil
	case res := <-resChan:
		if res.err != nil {
			return nil, nil, nil
		}
		return res.favorites, res.searches, res.readings
	}
}

// 生成推荐列表
func (s *RecommendService) GenerateRecommendations(ctx context.Context, userID uint, limit int) []model.Article {
	// 获取行为数据
	favorites, searches, readings := s.GetUserBehavior(ctx, userID)

	// 计算分类权重
	categoryWeights := make(map[uint]float64)
	// 收藏权重（最高）
	for _, f := range favorites {
		var article model.Article
		s.DB.WithContext(ctx).First(&article, f.ArticleID)
		categoryWeights[article.CategoryID] += 2.0
	}

	// 搜索关键词匹配（次高）
	searchKeywords := extractKeywords(searches)
	var articles []model.Article
	s.DB.WithContext(ctx).Find(&articles)
	for _, a := range articles {
		if containsAnyKeyword(a.Title, searchKeywords) || containsAnyKeyword(a.Content, searchKeywords) {
			categoryWeights[a.CategoryID] += 1.5
		}
	}

	// 阅读历史偏好（基础）
	for _, r := range readings {
		var article model.Article
		s.DB.WithContext(ctx).First(&article, r.ArticleID)
		categoryWeights[article.CategoryID] += 1.0
	}

	// 获取推荐分类
	recommendedCategories := getTopCategories(categoryWeights, 3)

	// 获取推荐文章
	var recommendations []model.Article
	s.DB.WithContext(ctx).Where("category_id IN (?) AND status = 'published'", recommendedCategories).
		Preload("Category").Preload("Tags").
		Order("created_at desc").
		Limit(limit).
		Find(&recommendations)

	return recommendations
}

// 辅助函数：提取搜索关键词
func extractKeywords(searches []model.SearchHistory) []string {
	keywords := make(map[string]bool)
	for _, s := range searches {
		// 简单分词逻辑（实际应使用更复杂的分词器）
		words := strings.Fields(s.SearchContent)
		for _, w := range words {
			if len(w) > 2 { // 过滤短词
				keywords[strings.ToLower(w)] = true
			}
		}
	}
	result := make([]string, 0, len(keywords))
	for k := range keywords {
		result = append(result, k)
	}
	return result
}

// 辅助函数：获取权重最高的分类
func getTopCategories(weights map[uint]float64, n int) []uint {
	type categoryWeight struct {
		CategoryID uint
		Weight     float64
	}

	var cwList []categoryWeight
	for id, w := range weights {
		cwList = append(cwList, categoryWeight{id, w})
	}

	sort.Slice(cwList, func(i, j int) bool {
		return cwList[i].Weight > cwList[j].Weight
	})

	result := make([]uint, 0, n)
	for i := 0; i < len(cwList) && i < n; i++ {
		result = append(result, cwList[i].CategoryID)
	}
	return result
}

// 简单关键词匹配
func containsAnyKeyword(text string, keywords []string) bool {
	lowerText := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lowerText, kw) {
			return true
		}
	}
	return false
}
