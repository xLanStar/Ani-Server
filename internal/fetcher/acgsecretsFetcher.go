package fetcher

import (
	Media "Ani-Server/internal/media"
	"fmt"
	"strconv"
	"strings"
	"time"

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

func InitacgsecretsFetcher() {
	acgsecretsUrlFormat = "https://acgsecrets.hk/bangumi/%4d%02d/"
	acgsecretsC = colly.NewCollector()

	acgsecretsC.OnHTML(".anime_content", func(e *colly.HTMLElement) {
		var title_jp string
		e.ForEach(".entity_original_name", func(_ int, el *colly.HTMLElement) { title_jp = el.Text })

		var title_tw string
		e.ForEach(".entity_localized_name", func(_ int, el *colly.HTMLElement) { title_tw = strings.Trim(el.Text, " \n\r") })

		var title_alt []string

		_title_alt := e.ChildText("i")

		if len(_title_alt) > 15 && _title_alt[:15] == "其他名稱：" {
			title_alt = strings.Split(_title_alt[15:], "、")
		}

		var mediaId uint32
		var media Media.IMedia

		// 查詢作品演算法
		for t := 0; t != 3 && mediaId == 0; t++ {
			var subs []string

			if t == 0 {
				subs = []string{title_jp}
			} else if t == 1 {
				subs = strings.Split(title_jp, " ")
				if len(subs) != 0 {
					subs = subs[:len(subs)-1]
				}
			} else if t == 2 {
				subs = title_alt[:]
			}

			for _, sub := range subs {
				time.Sleep(time.Second)

				media, _ = SearchMedia(sub, seasonYear, season)
				if media.GetId() != 0 {
					mediaId = media.GetId()
					break
				}
			}
		}

		description := strings.Trim(e.ChildText(".anime_story"), " \n\r")

		if mediaId == 0 {
			fmt.Println("找不到作品 標題:", title_tw, "即將建置一個")
			return
		} else {
			fmt.Println("找到作品 ID:", mediaId, "標題:", title_tw)
			media.SetTitle(title_tw)
			media.SetDescription(description)
		}

		var Id_if101 int = 0
		if101links := make([]If101Link, 0, 10)

		// 開始抓取 if101 網頁
		for _, title := range title_alt {
			t_if101links := SearchIf101Links(title)

			if len(t_if101links) == 1 {
				// 找到唯一
				Id_if101, _ = strconv.Atoi(t_if101links[0].Str_Id)
				media.(*Media.Anime).Id_if101 = uint32(Id_if101)
				fmt.Println("找到 IF101 ID:", Id_if101, "標題:", t_if101links[0].Name)
				break
			} else {
				if101links = append(if101links, t_if101links...)
			}
		}

		// 選擇 Id_if101
		if Id_if101 == 0 && len(if101links) != 0 {
			selection[mediaId] = if101links
		}

		result[mediaId] = media
	})
}

func FetchacgsecretsPage(SeasonYear int, Season int) (*map[uint32]Media.IMedia, *map[uint32][]If101Link) {
	if len(searchPoster.Url) == 0 {
		InitacgsecretsFetcher()
	}
	seasonYear = SeasonYear
	season = seasons[Season]
	result = make(map[uint32]Media.IMedia, 100)
	selection = make(map[uint32][]If101Link)

	acgsecretsC.Visit(fmt.Sprintf(acgsecretsUrlFormat, SeasonYear, 1+Season*3))

	return &result, &selection
}
