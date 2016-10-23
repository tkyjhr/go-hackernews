package hackernews

import (
	"fmt"
	"testing"
)

func TestGetStoryItems(t *testing.T) {
	maxStoryCount := 10
	_, count, err := GetStoryItems(TopStories, maxStoryCount, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != maxStoryCount {
		t.Errorf("len(stories) != maxStoryCount(%v)", maxStoryCount)
	}
}

func TestGetStoryItemsMany(t *testing.T) {
	maxStoryCount := 1000
	stories, count, err := GetStoryItems(NewStories, maxStoryCount, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 500 {
		// 500 = Hacker News API から取得できる最大数。
		t.Errorf("len(stories)(%v) != 500", len(stories))
	}
}

func TestGetStoryItem(t *testing.T) {
	{
		// item が取得されることを期待する。
		_, err := GetStoryItem(8863, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		// 存在しないため取得できずエラーが返ってくることを期待する。
		item, err := GetStoryItem(0, nil)
		if err == nil {
			t.Fatal(item)
		}
	}
}

func TestGetStoryItemAsync(t *testing.T) {
	{
		// item が取得されることを期待する。
		itemChan, errChan := GetStoryItemAsync(8863, nil)
		select {
		case item := <-itemChan:
			fmt.Println(item)
			break
		case err := <-errChan:
			if err != nil {
				t.Fatal(err)
				break
			}
		}
	}

	{
		// 存在しないため取得できずエラーが返ってくることを期待する。
		itemChan, errChan := GetStoryItemAsync(0, nil)
		fmt.Println("Waiting...")
		select {
		case item := <-itemChan:
			t.Fatal(item)
			break
		case err := <-errChan:
			if err != nil {
				fmt.Println(err)
				break
			}
		}
	}
}

func TestGetStoryItemsAsync(t *testing.T) {
	maxStoryCount := 10
	itemChan, _ := GetStoryItemsAsync(NewStories, maxStoryCount, 1, nil)
	for i := 0; i < maxStoryCount; i++ {
		item := <-itemChan
		fmt.Println(item)
	}
}

func TestSortByScore(t *testing.T) {
	maxStoryCount := 10
	stories, _, err := GetStoryItems(TopStories, maxStoryCount, nil)
	if err != nil {
		t.Fatal(err)
	}
	SortByScore(stories)
	pre := 0
	for _, s := range stories {
		if pre > s.Score {
			t.Fatal()
		}
	}
}
