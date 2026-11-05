package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"

	"tongcheng/conf"
)

var (
	es *elasticsearch.Client
)

func Init() error {
	config := conf.GetConfiguration().Elasticsearch

	cfg := elasticsearch.Config{
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
		CloudID:   config.CloudID,
		APIKey:    config.APIKey,
	}

	var err error
	es, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("创建Elasticsearch客户端失败: %v", err)
	}

	resp, err := es.Info()
	if err != nil {
		return fmt.Errorf("连接Elasticsearch失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("elasticsearch连接异常，状态码: %d", resp.StatusCode)
	}

	return nil
}

func GetES() *elasticsearch.Client {
	return es
}

func CreateIndex(ctx context.Context, indexName string, mapping string) error {
	req := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  strings.NewReader(mapping),
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("创建索引失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

func DeleteIndex(ctx context.Context, indexName string) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return fmt.Errorf("删除索引失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("删除索引失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

func IndexDocument(ctx context.Context, indexName string, docID string, doc any) error {
	var body string
	switch d := doc.(type) {
	case string:
		body = d
	case []byte:
		body = string(d)
	default:
		body = fmt.Sprintf("%v", d)
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       strings.NewReader(body),
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return fmt.Errorf("索引文档失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("索引文档失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

func GetDocument(ctx context.Context, indexName string, docID string) (*esapi.Response, error) {
	req := esapi.GetRequest{
		Index:      indexName,
		DocumentID: docID,
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return nil, fmt.Errorf("获取文档失败: %v", err)
	}

	if resp.StatusCode >= 400 {
		defer func() {
			_ = resp.Body.Close()
		}()
		return nil, fmt.Errorf("获取文档失败，状态码: %d", resp.StatusCode)
	}

	return resp, nil
}

func DeleteDocument(ctx context.Context, indexName string, docID string) error {
	req := esapi.DeleteRequest{
		Index:      indexName,
		DocumentID: docID,
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return fmt.Errorf("删除文档失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("删除文档失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

func SearchDocuments(ctx context.Context, indexName string, query string) (*esapi.Response, error) {
	req := esapi.SearchRequest{
		Index: []string{indexName},
		Body:  strings.NewReader(query),
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return nil, fmt.Errorf("搜索文档失败: %v", err)
	}

	if resp.StatusCode >= 400 {
		defer func() {
			_ = resp.Body.Close()
		}()
		return nil, fmt.Errorf("搜索文档失败，状态码: %d", resp.StatusCode)
	}

	return resp, nil
}

func BulkIndex(ctx context.Context, indexName string, docs []BulkDoc) error {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: es,
		Index:  indexName,
	})
	if err != nil {
		return fmt.Errorf("创建批量索引器失败: %v", err)
	}

	for _, doc := range docs {
		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: doc.ID,
			Body:       strings.NewReader(doc.Content),
		})
		if err != nil {
			return fmt.Errorf("添加文档到批量索引器失败: %v", err)
		}
	}

	if err = bi.Close(ctx); err != nil {
		return fmt.Errorf("执行批量索引失败: %v", err)
	}

	return nil
}

type BulkDoc struct {
	ID      string
	Content string
}

func UpdateDocument(ctx context.Context, indexName string, docID string, updateData string) error {
	req := esapi.UpdateRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       strings.NewReader(updateData),
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return fmt.Errorf("更新文档失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("更新文档失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

func ExistsDocument(ctx context.Context, indexName string, docID string) (bool, error) {
	req := esapi.ExistsRequest{
		Index:      indexName,
		DocumentID: docID,
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return false, fmt.Errorf("检查文档存在性失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	} else {
		return false, fmt.Errorf("检查文档存在性失败，状态码: %d", resp.StatusCode)
	}
}

func CountDocuments(ctx context.Context, indexName string, query string) (*esapi.Response, error) {
	req := esapi.CountRequest{
		Index: []string{indexName},
		Body:  strings.NewReader(query),
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return nil, fmt.Errorf("统计文档数量失败: %v", err)
	}

	if resp.StatusCode >= 400 {
		defer func() {
			_ = resp.Body.Close()
		}()
		return nil, fmt.Errorf("统计文档数量失败，状态码: %d", resp.StatusCode)
	}

	return resp, nil
}
