package fetcher

import (
	Media "Ani-Server/internal/media"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

var (
	acgsecretsUrlFormat string
	acgsecretsC         *colly.Collector
	seasonYear          int
	season              string
	// if101 搜尋
	result    map[uint32]Media.IMedia
	selection map[uint32][]If101Link
	seasons   = []string{"WINTER", "SPRING", "SUMMER", "FALL"}
)

func initACGSecretFetcher() {
	acgsecretsUrlFormat = "https://acgsecrets.hk/bangumi/%4d%02d/"
	acgsecretsC = colly.NewCollector()

	acgsecretsC.OnHTML(".anime_content", func(e *colly.HTMLElement) {
		var titleJP string
		e.ForEach(".entity_original_name", func(_ int, el *colly.HTMLElement) { titleJP = el.Text })

		var titleTW string
		e.ForEach(".entity_localized_name", func(_ int, el *colly.HTMLElement) { titleTW = strings.Trim(el.Text, " \n\r") })

		var titleALT []string

		_titleALT := e.ChildText("i")

		if len(_titleALT) > 15 && _titleALT[:15] == "其他名稱：" {
			titleALT = strings.Split(_titleALT[15:], "、")
		}

		media := AdvancedSearchMedia(titleJP, titleALT)

		description := strings.Trim(e.ChildText(".anime_story"), " \n\r")

		if media.GetId() == 0 {
			fmt.Println("找不到作品 標題:", titleTW, "即將建置一個")
			return
		} else {
			fmt.Println("找到作品 ID:", media.GetId(), "標題:", titleTW)
			media.SetTitle(titleTW)
			media.SetDescription(description)
		}

		var if101ID int = 0
		if101links := make([]If101Link, 0, 10)

		// 開始抓取 if101 網頁
		for _, title := range titleALT {
			tIf101Links := SearchIf101Links(title)

			if len(tIf101Links) == 1 {
				// 找到唯一
				if101ID, _ = strconv.Atoi(tIf101Links[0].IdStr)
				media.(*Media.Anime).If101Id = uint32(if101ID)
				fmt.Println("找到 IF101 ID:", if101ID, "標題:", tIf101Links[0].Name)
				break
			} else {
				if101links = append(if101links, tIf101Links...)
			}
		}

		// 選擇 if101ID
		if if101ID == 0 && len(if101links) != 0 {
			selection[media.GetId()] = if101links
		}

		result[media.GetId()] = media
	})
}

func FetchacgsecretsPage(SeasonYear int, Season int) (*map[uint32]Media.IMedia, *map[uint32][]If101Link) {
	if len(searchPoster.Url) == 0 {
		initACGSecretFetcher()
	}
	seasonYear = SeasonYear
	season = seasons[Season]
	result = make(map[uint32]Media.IMedia, 100)
	selection = make(map[uint32][]If101Link)

	acgsecretsC.Visit(fmt.Sprintf(acgsecretsUrlFormat, SeasonYear, 1+Season*3))

	return &result, &selection
}
