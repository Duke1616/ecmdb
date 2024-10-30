package cache

import (
	"fmt"
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"log"
	"time"
)

func Schema() {
	// 创建 RediSearch 索引
	index := redisearch.NewClient("localhost:6379", "idx:books")

	// 定义索引结构
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewTextField("body")).
		AddField(redisearch.NewTextFieldOptions("title", redisearch.TextFieldOptions{Weight: 5.0, Sortable: true})).
		AddField(redisearch.NewNumericField("date"))
	// Drop an existing index. If the index does not exist an error is returned
	err := index.Drop()
	if err != nil {
		return
	}

	// 创建索引（如果索引已存在，可以选择忽略此步骤）
	// Create the index with the given schema
	if err := index.CreateIndex(sc); err != nil {
		log.Fatal(err)
	}

	// 添加文档到索引
	doc := redisearch.NewDocument("doc1", 1.0)
	doc.Set("title", "Hello world").
		Set("body", "foo bar").
		Set("date", time.Now().Unix())

	if err := index.Index([]redisearch.Document{doc}...); err != nil {
		log.Fatal(err)
	}

	// 继续添加更多文档...

	// 执行搜索
	docs, total, err := index.Search(redisearch.NewQuery("hello world").
		Limit(0, 2).
		SetReturnFields("title"))

	// 获取总结果数
	totalResults := total
	fmt.Printf("Total Results: %d\n", totalResults)

	// 输出结果
	for _, doc := range docs {
		fmt.Printf("ID: %s, Title: %s, Author: %s\n", doc.Id, doc.Properties["title"], doc.Properties["author"])
	}
}
