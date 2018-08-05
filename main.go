package main

import (
	"fmt"
	"strings"
	"encoding/csv"
	"strconv"
	"sort"
	"os"
	"time"
	"io"
)

func main() {
	var language string
	var codes *[]Code
	language = "go"

	codes = fetchSearchCode(language)
	writeCodesCSV(fmt.Sprintf("./output/%s/codes.csv", language), codes)
}

func executeRepositories() {
	var language string

	language = "go"

	repositories := fetchRepositories(language)
	writeRepositoriesCSV(fmt.Sprintf("./output/%s/repositories.csv", language), repositories)
}

type Code struct {
	rawUrl string
	path   string
}

func fetchSearchCode(language string) *[]Code {

	var err error
	var fw *os.File
	var codeResponses *[]CodeResponse
	var codes []Code

	fw, err = NewFile(fmt.Sprintf("./output/%s/search_code_api_log.txt", language))
	if err != nil {
		fmt.Println(err)
	}
	logWriter := csv.NewWriter(fw)
	client := NewClient(logWriter)

	fr, err := os.Open(fmt.Sprintf("./output/%s/repositories.csv", language))
	if err != nil {
		fmt.Println(err)
	}
	reader := csv.NewReader(fr)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			break
		}
		repositoryName := record[0]
		defaultBranch := record[1]

		for page := 1; page <= 10; page++ {
			codeQuery := CodeQuery{
				q:              "",
				repositoryName: repositoryName,
				language:       language,
				extension:      pString("go"),
			}
			codeResponses, err = client.fetchCode(codeQuery, page)
			if err != nil {
				fmt.Println(err)
			} else {
				for _, codeResponse := range *codeResponses {
					codes = append(codes, Code{
						rawUrl: fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", codeQuery.repositoryName, defaultBranch, codeResponse.Path),
						path:   codeResponse.Path,
					})
				}
			}
		}
	}
	return &codes

}

func fetchRepositories(language string) *[]RepositoryResponse {

	var repositoryMap map[string]RepositoryResponse
	var repositories *[]RepositoryResponse
	var err error
	var fw *os.File

	const GITHUB_API_WAITING_TIME = 60
	const ALPHABET = 26

	repositoryMap = make(map[string]RepositoryResponse)

	fw, err = NewFile(fmt.Sprintf("./output/%s/api_log.txt", language))
	if err != nil {
		fmt.Println(err)
	}
	logWriter := csv.NewWriter(fw)
	client := NewClient(logWriter)

	for i := 1; i <= ALPHABET; i++ {
		initial := alphabet(i)

		for page := 1; page <= 10; page++ {
			repositories, err = client.fetchRepository(initial, language, page)
			if err != nil {
				fmt.Println(err)
			}
			if repositories != nil {
				for _, repo := range *repositories {
					repositoryMap[repo.FullName] = repo
				}
			}
		}
		time.Sleep(GITHUB_API_WAITING_TIME * time.Second)
	}

	sortedRepositories := sortMap(repositoryMap)

	defer fw.Close()
	return sortedRepositories
}

func writeCodesCSV(filename string, codes *[]Code) {

	var err error
	var fw *os.File
	fw, err = NewFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	writer := csv.NewWriter(fw)
	for i, code := range (*codes) {
		writer.Write([]string{strconv.Itoa(i), code.rawUrl, code.path})
	}
	writer.Flush()

	defer fw.Close()
}

func writeRepositoriesCSV(filename string, sortedRepositories *[]RepositoryResponse) {

	var err error
	var fw *os.File
	fw, err = NewFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	writer := csv.NewWriter(fw)
	for _, repository := range (*sortedRepositories) {
		writer.Write([]string{repository.FullName, repository.DefaultBranch, strconv.Itoa(repository.StargazersCount)})
	}
	writer.Flush()

	defer fw.Close()
}

func sortMap(repositoryMap map[string]RepositoryResponse) *[]RepositoryResponse {

	var repositories []RepositoryResponse
	for _, v := range repositoryMap {
		repositories = append(repositories, RepositoryResponse{
			FullName:        v.FullName,
			StargazersCount: v.StargazersCount,
			DefaultBranch:   v.DefaultBranch,
		})
	}
	sort.Slice(repositories, func(i, j int) bool {
		return repositories[i].StargazersCount > repositories[j].StargazersCount
	})

	return &repositories

}

func Keys(m map[string]RepositoryResponse) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func alphabet(i int) string {

	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	return strings.Split(alphabet, "")[i-1]
}
