package main

import (
	"github.com/google/go-github/github"
	"context"
	"fmt"
	"strings"
	"encoding/csv"
	"time"
)

type Client struct {
	client       *github.Client
	clientID     string
	clientSecret string
	logWriter    *csv.Writer
	api          map[API]string
}

type API int

const (
	SEARCH_REPOSITORIES = iota
	SEARCH_CODE
)

func NewClient(csvWriter *csv.Writer) *Client {

	var api map[API]string

	api = map[API]string{
		SEARCH_CODE:         "https://api.github.com/search/code",
		SEARCH_REPOSITORIES: "https://api.github.com/search/repositories",
	}

	return &Client{
		client:       github.NewClient(nil),
		clientID:     "e6bcf720314a21989e9b",
		clientSecret: "9ed627c12d5bf597459c0b86154f7f7e52a37138",
		logWriter:    csvWriter,
		api:          api,
	}
}

type CodeQuery struct {
	q              string
	repositoryName string
	language       string
	extension      *string
}

func (*Client) createCodeQuery(query CodeQuery) string {

	var outQuery []string

	outQuery = append(outQuery, query.q)
	outQuery = append(outQuery, fmt.Sprintf("in:file"))
	outQuery = append(outQuery, fmt.Sprintf("language:%s", query.language))
	outQuery = append(outQuery, fmt.Sprintf("repo:%s", query.repositoryName))

	if query.extension != nil {
		outQuery = append(outQuery, fmt.Sprintf("extension:%s", *query.extension))
	}
	return strings.Join(outQuery, " ")
}

func (c *Client) fetchCode(query CodeQuery, page int) (*[]CodeResponse, error) {

	var codeResponses []CodeResponse

	q := c.createCodeQuery(query)
	opts := &github.SearchOptions{Sort: "indexed", Order: "desc", ListOptions: github.ListOptions{PerPage: 100, Page: page}}

	auth,res,err:=c.client.Authorizations.Create(context.Background(), &github.AuthorizationRequest{ClientID: &c.clientID, ClientSecret: &c.clientSecret})
	if err != nil{
		fmt.Println(err)
	}else{
		fmt.Println(*auth,*res)
	}


	url := fmt.Sprintf("%s?q=%s+language:%s+repo:%s&sort=indexed&order=desc&per_page=100&page=%d&client_id=%s&client_secret=%s", c.api[SEARCH_CODE], query.q, query.language, query.repositoryName, page, c.clientID, c.clientSecret)
	c.logWriter.Write([]string{time.Now().String(), url})
	c.logWriter.Flush()
	debugLog(strings.Join([]string{time.Now().String(), url}, ","))

	result, _, err := c.client.Search.Code(context.Background(), q, opts)
	if err != nil {
		return nil, err
	}

	for _, item := range result.CodeResults {
		var codeResponse CodeResponse
		codeResponse = CodeResponse{
			Path: item.GetPath(),
		}
		codeResponses = append(codeResponses, codeResponse)
	}
	return &codeResponses, nil
}

type CodeResponse struct {
	Path string
}

type RepositoryResponse struct {
	FullName        string
	StargazersCount int
	DefaultBranch   string
}

func debugLog(log string) {
	fmt.Println(log)
}

func (c *Client) fetchRepository(word string, language string, page int) (*[]RepositoryResponse, error) {

	var repositories []RepositoryResponse

	q := fmt.Sprintf("%s language:%s", word, language)
	opts := &github.SearchOptions{Sort: "stars", Order: "desc", ListOptions: github.ListOptions{PerPage: 100, Page: page},}

	url := fmt.Sprintf("%s?q=%s+language:%s&sort=stars&order=desc&per_page=100&page=%d", c.api[SEARCH_REPOSITORIES], word, language, page)
	c.logWriter.Write([]string{time.Now().String(), url})
	c.logWriter.Flush()
	debugLog(strings.Join([]string{time.Now().String(), url}, ","))

	result, _, err := c.client.Search.Repositories(context.Background(), q, opts)
	if err != nil {
		return nil, err
	}

	for _, item := range result.Repositories {

		var repository RepositoryResponse

		repository = RepositoryResponse{
			FullName:        item.GetFullName(),
			StargazersCount: item.GetStargazersCount(),
			DefaultBranch:   item.GetDefaultBranch(),
		}
		repositories = append(repositories, repository)
	}

	return &repositories, nil
}
