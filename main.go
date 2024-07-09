package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type AnimeRate struct {
	Anime struct {
		ID   int    `json:"id"`
		Name string `json:"russian"`
	} `json:"anime"`
	Status string `json:"status"`
	Score  int    `json:"score"`
}

func getUserRatedAnime(userID string, accessToken string) (map[int]string, map[int]int, error) {
	ratedAnime := make(map[int]string)
	animeScore := make(map[int]int)
	page := 1

	client := &http.Client{}

	for {
		url := fmt.Sprintf("https://shikimori.one/api/users/%v/anime_rates?status=completed&page=%d&limit=50", userID, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, nil, err
		}

		// USER-AGENT
		req.Header.Set("User-Agent", "your_app")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		res, err := client.Do(req)
		if err != nil {
			return nil, nil, err
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusTooManyRequests {
			fmt.Println("Лимит скорости превышен. Ожидайте 1 минуту...")
			time.Sleep(1 * time.Minute)
			continue
		}

		if res.StatusCode != http.StatusOK {
			return nil, nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, nil, err
		}

		var animeRates []AnimeRate
		err = json.Unmarshal(body, &animeRates)
		if err != nil {
			return nil, nil, err
		}

		if len(animeRates) == 0 {
			break
		}

		for _, rate := range animeRates {
			if rate.Status == "completed" {
				ratedAnime[rate.Anime.ID] = rate.Anime.Name
				animeScore[rate.Anime.ID] = rate.Score
			}
		}
		page++
	}

	return ratedAnime, animeScore, nil
}

func main() {
	// ACCESS TOKEN
	accessToken := "your_token"

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Пользователь 1: ")
	userID1, _ := reader.ReadString('\n')
	userID1 = strings.TrimSuffix(userID1, "\n")

	reader = bufio.NewReader(os.Stdin)
	fmt.Println("Пользователь 2: ")
	userID2, _ := reader.ReadString('\n')
	userID2 = strings.TrimSuffix(userID2, "\n")

	user1RatedAnime, user1Score, err := getUserRatedAnime(userID1, accessToken)
	if err != nil {
		log.Fatal(err)
	}

	user2RatedAnime, user2Score, err := getUserRatedAnime(userID2, accessToken)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Аниме, которое смотрел только ", userID1)
	for animeID, animeName := range user1RatedAnime {
		if _, exists := user2RatedAnime[animeID]; !exists {
			fmt.Println(animeName, ", оценка: ", user1Score[animeID])
		}
	}

	fmt.Println("\nАниме, которое смотрел только ", userID2)
	for animeID, animeName := range user2RatedAnime {
		if _, exists := user1RatedAnime[animeID]; !exists {
			fmt.Println(animeName, ", оценка: ", user2Score[animeID])
		}
	}
}
