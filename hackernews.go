package hackernews

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"

	"github.com/dustin/gojson"
)

// About HackerNews API : https://github.com/HackerNews/API

const (
	baseUrl = "https://hacker-news.firebaseio.com/v0/"
)

type StoryCategory int

const (
	TopStories StoryCategory = iota
	NewStories
	BestStories
)

func (c StoryCategory) URL() string {
	switch c {
	case TopStories:
		return baseUrl + "topstories.json"
	case NewStories:
		return baseUrl + "newstories.json"
	case BestStories:
		return baseUrl + "beststories.json"
	default:
		panic("unsupported category")
	}
}

func getRawData(url string, client *http.Client) ([]byte, error) {
	if client == nil {
		client = &http.Client{}
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

type StoryItem struct {
	ID    int    `json:"id"`
	By    string `json:"by"`
	Time  int64  `json:"time"`
	Title string `json:"title"`
	Score int    `json:"score"`
	Url   string `json:"url"`
}

func GetStoryItem(id int, client *http.Client) (StoryItem, error) {
	itemChan, errChan := GetStoryItemAsync(id, client)
	err := <-errChan
	if err != nil {
		return StoryItem{}, err
	} else {
		item := <-itemChan
		return item, nil
	}
}

func GetStoryItemAsync(id int, client *http.Client) (chan StoryItem, chan error) {
	itemChan := make(chan StoryItem, 1)
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			close(itemChan)
			close(errChan)
		}()
		var item StoryItem
		data, err := getRawData(baseUrl+"item/"+strconv.Itoa(id)+".json", client)
		if err != nil {
			errChan <- err
			return
		}
		if bytes.Compare(data, []byte{0x6E, 0x75, 0x6C, 0x6C}) == 0 { // null.
			errChan <- fmt.Errorf("ID %d does not exist.", id)
			return
		}
		err = json.Unmarshal(data, &item)
		if err != nil {
			errChan <- err
			return
		}
		itemChan <- item
		errChan <- nil
	}()
	return itemChan, errChan
}

func GetStoryItems(category StoryCategory, maxStoryCount int, client *http.Client) ([]StoryItem, int, error) {
	itemChan, errChan := GetStoryItemsAsync(category, maxStoryCount, 10, client) // TORIAEZU : chan サイズは適当な固定値。
	storyItem := make([]StoryItem, maxStoryCount)
	count := 0
	for {
		item, ok := <-itemChan
		if ok {
			storyItem[count] = item
			count++
		} else {
			err := <-errChan
			return storyItem, count, err
		}
	}
}

func GetStoryItemsAsync(category StoryCategory, maxStoryCount, storyItemChanSize int, client *http.Client) (chan StoryItem, chan error) {
	itemsChan := make(chan StoryItem, storyItemChanSize)
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			close(itemsChan)
			close(errChan)
		}()
		data, err := getRawData(category.URL(), client)
		if err != nil {
			errChan <- err
			return
		}
		var stories []int
		err = json.Unmarshal(data, &stories)
		if err != nil {
			errChan <- err
			return
		}
		for i, id := range stories {
			if i >= maxStoryCount {
				break
			}
			item, err := GetStoryItem(id, client)
			if err != nil {
				errChan <- err
				return
			}
			itemsChan <- item
		}
		errChan <- nil
	}()
	return itemsChan, errChan
}

func FilterByScore(items []StoryItem, threshold int) []StoryItem {
	var filteredItems = items[:0]
	for _, item := range items {
		if item.Score >= threshold {
			filteredItems = append(filteredItems, item)
		}
	}
	return filteredItems
}

type stories []StoryItem

func (s stories) Len() int {
	return len(s)
}

func (s stories) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s stories) Less(i, j int) bool {
	return s[i].Score < s[j].Score
}

func SortByScore(items []StoryItem) {
	var stories stories = items
	sort.Sort(stories)
}
