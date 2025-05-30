package model

import (
	"BBingyan/internal/log"
	"bytes"
	"encoding/json"
	es "github.com/elastic/go-elasticsearch/v8"
	"strconv"
	"time"
)

type SimplePost struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Tag       string    `json:"tag"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created-at"`
}

var ES *es.Client

type IndexSettings struct {
	Aliases  map[string]map[string]interface{} `json:"aliases,omitempty"`
	Mappings struct {
		Properties map[string]map[string]interface{} `json:"properties"`
	} `json:"mappings"`
}

func newElasticsearch() {
	client, err := es.NewClient(es.Config{
		Addresses: []string{"http://localhost:9200"},
		Username:  "test",
		Password:  "123456",
	})
	if err != nil {
		log.Fatalf("Fail to init es,err:%v", err)
	}
	ES = client

	res, err := ES.Indices.Exists([]string{"post"})
	if err != nil {
		log.Fatalf("Fail to check index,err:%v", err)
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		log.Infof("初始化索引")
		CreatePostIndex()
	} else if res.StatusCode != 200 {
		log.Fatalf("不确定的index状态,code:%d", res.StatusCode)
	}
}

func CreatePostIndex() {
	req := IndexSettings{}
	req.Mappings.Properties = map[string]map[string]interface{}{
		"id": {
			"type": "keyword",
		},
		"title": {
			"type":     "text",
			"analyzer": "ik_max_word",
		},
		"tag": {
			"type": "keyword",
		},
		"author": {
			"type": "keyword",
		},
		"content": {
			"type":     "text",
			"analyzer": "ik_max_word",
		},
		"created-at": {
			"type": "date",
		},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		log.Fatalf("Error encoding request: %s", err)
	}
	res, err := ES.Indices.Create("post", ES.Indices.Create.WithBody(&buf))
	if err != nil {
		log.Fatalf("Fail to init index,err:%s", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		log.Fatalf("Index creation failed: %s", res.String())
	}
}

func AddPostToES(post *Post) error {
	simple := SimplePost{
		ID:        post.ID,
		Title:     post.Title,
		Tag:       post.Tag,
		Author:    post.Author,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
	}
	id := strconv.Itoa(int(simple.ID))
	simpleJson, _ := json.Marshal(simple)

	_, err := ES.Index(
		"post",
		bytes.NewReader(simpleJson),
		ES.Index.WithDocumentID(id),
		ES.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}

	return nil
}

// SearchPost 如果未指定tag,传入""
func SearchPost(tag string, query string, desc bool, page int, pageSize int) ([]SimplePost, error) {
	var order string
	if desc {
		order = "desc"
	} else {
		order = "asc"
	}

	var DSL map[string]interface{}

	if tag == "" {
		DSL = map[string]interface{}{
			"from": page,
			"size": pageSize,
			"sort": []map[string]interface{}{
				{"created-at": map[string]string{"order": order}},
			},
			"query": map[string]interface{}{
				"match": map[string]string{"content": query},
			},
			"highlight": map[string]interface{}{
				"fields": map[string]interface{}{
					"content": map[string]interface{}{},
				},
			},
		}
	} else {
		DSL = map[string]interface{}{
			"from": page,
			"size": pageSize,
			"sort": []map[string]interface{}{
				{"created-at": map[string]string{"order": order}},
			},
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []interface{}{
						map[string]interface{}{"match": map[string]string{"content": query}},
						map[string]interface{}{"match": map[string]string{"tag": tag}},
					},
				},
			},
			"highlight": map[string]interface{}{
				"fields": map[string]interface{}{
					"content": map[string]interface{}{},
				},
			},
		}
	}
	DSLJson, _ := json.Marshal(DSL)

	res, err := ES.Search(
		ES.Search.WithIndex("post"),
		ES.Search.WithBody(bytes.NewReader(DSLJson)),
		ES.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	log.Infof("%v", result)

	data := result["hits"].(map[string]interface{})
	hits := data["hits"].([]interface{})

	posts := make([]SimplePost, 0)

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		sourceJson, _ := json.Marshal(source)
		var simple SimplePost
		if err := json.Unmarshal(sourceJson, &simple); err != nil {
			log.Warnf("Fail to unmarshal data,err:%s", err)
		}

		//用第一处高亮文段替换一下content内容
		highlights := hit.(map[string]interface{})["highlight"].(map[string]interface{})
		highlight := highlights["content"].([]interface{})[0].(string)
		simple.Content = highlight

		posts = append(posts, simple)
	}

	return posts, nil
}

func DeletePostInES(id int) error {
	Id := strconv.Itoa(id)
	_, err := ES.Delete("post", Id)
	return err
}
