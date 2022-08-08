package fetcher

import (
	"Ani-Server/internal/media"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type If101Link struct {
	Name     string
	Str_Id   string
	UpdateAt uint32
}
type videoSource struct {
	isNumber bool
	name     interface{}
	video    string
}

const (
	if101UpdateUrlFormat string = "https://demo.if101.tv/index.php/vod/type/id/136/page/%d.html" // 更新 URL
	if101SearchUrlFormat string = "https://demo.if101.tv/index.php/vod/type/id/136.html?wd=%s"   // 搜尋 URL
	if101DetailUrlFormat string = "https://demo.if101.tv/index.php/vod/detail/id/%d.html"        // 作品資訊 URL
)

var (
	updateC *colly.Collector
	searchC *colly.Collector
	detailC *colly.Collector

	// 搜尋 Search
	buffer_Search_links_if101 []If101Link
	temp_Search_links_if101   []If101Link

	// 更新 Update
	buffer_Update_links_if101 []If101Link
	temp_Update_links_if101   []If101Link

	// 資訊 Detail
	buffer_videoSources_if101 []videoSource
	temp_videoSources_if101   []videoSource
)

func initUpdateFetcher() {
	buffer_Update_links_if101 = make([]If101Link, 100)

	// if101 取得資訊
	updateC = colly.NewCollector()
	updateC.OnHTML(".stui-vodlist .clearfix", func(e *colly.HTMLElement) {
		s_UpdateAt := e.ChildText("span.time")

		if len(s_UpdateAt) < 16 {
			return
		}

		s_id_if101 := e.ChildAttr("a[title]", "href")

		var updateAt uint32 = 0

		for _, c := range s_UpdateAt[:13] {
			if '0' <= c && c <= '9' {
				updateAt = (updateAt << 3) + (updateAt << 1) + uint32(c-'0')
			}
		}

		var link If101Link = If101Link{
			Str_Id:   s_id_if101[25 : len(s_id_if101)-5],
			Name:     e.ChildAttr("a[title]", "title"),
			UpdateAt: updateAt,
		}

		temp_Update_links_if101 = append(temp_Update_links_if101, link)
	})
}

func InitSearchFetcher() {
	buffer_Search_links_if101 = make([]If101Link, 100)

	// if101 取得資訊
	searchC = colly.NewCollector()
	searchC.OnHTML(".stui-vodlist .clearfix", func(e *colly.HTMLElement) {
		s_UpdateAt := e.ChildText("span.time")

		if len(s_UpdateAt) < 16 {
			return
		}

		s_id_if101 := e.ChildAttr("a[title]", "href")

		var updateAt uint32 = 0

		for _, c := range s_UpdateAt[:13] {
			if '0' <= c && c <= '9' {
				updateAt = (updateAt << 3) + (updateAt << 1) + uint32(c-'0')
			}
		}

		var link If101Link = If101Link{
			Str_Id:   s_id_if101[25 : len(s_id_if101)-5],
			Name:     e.ChildAttr("a[title]", "title"),
			UpdateAt: updateAt,
		}

		temp_Search_links_if101 = append(temp_Search_links_if101, link)
	})
}
func initDetailFetcher() {
	buffer_videoSources_if101 = make([]videoSource, 1200)

	// if101 取得資訊
	detailC = colly.NewCollector()
	detailC.OnHTML(".stui-content__playlist", func(e *colly.HTMLElement) {
		e.ForEach(".copy_text", func(i int, e *colly.HTMLElement) {
			m := strings.Index(e.Text, "$")
			name := strings.ToUpper(strings.Trim(e.Text[:m], " \r\n"))
			video := videoSource{
				video: e.Text[m+28 : len(e.Text)-21],
			}

			f, err := strconv.ParseFloat(name, 64)

			if err == nil {
				video.isNumber = true
				video.name = f
			} else {
				video.name = name
			}

			temp_videoSources_if101 = append(temp_videoSources_if101, video)
		})
	})
}

// 搜尋指定的關鍵字，並回傳所有的連結
func SearchIf101Links(search string) []If101Link {
	if len(buffer_Search_links_if101) == 0 {
		fmt.Println("if101 init")
		InitSearchFetcher()
	}

	temp_Search_links_if101 = buffer_Update_links_if101[0:0]

	searchC.Visit(fmt.Sprintf(if101SearchUrlFormat, search))

	// result := make([]If101Link, len(temp_Search_links_if101))

	// for i := 0; i != len(temp_Search_links_if101); i++ {
	// 	result[i] = temp_Search_links_if101[i]
	// }

	return temp_Search_links_if101
}

// 以指定的時間為基準，將所有此時間點以後的 if101 id 傳回
func GetUpdatedIds(lastUpdateAt uint32) []uint32 {
	if len(buffer_Update_links_if101) == 0 {
		initUpdateFetcher()
	}

	var result []uint32 = make([]uint32, 0, 50)

	for page, breakpoint := 1, false; !breakpoint; page++ {
		temp_Update_links_if101 = buffer_Update_links_if101[0:0]

		updateC.Visit(fmt.Sprintf(if101UpdateUrlFormat, page))

		for _, link := range temp_Update_links_if101 {
			if link.UpdateAt < lastUpdateAt {
				breakpoint = true
				continue
			}

			id_if101, _ := strconv.Atoi(link.Str_Id)

			result = append(result, uint32(id_if101))
		}
	}

	return result
}

// 基於以建置之 Media，填入 if101 資源
// 回傳 (是否有更新, 錯誤)
func FetchIf101Details(anime *media.Anime) bool {
	if anime.Id_if101 == 0 {
		return false
	}

	if buffer_videoSources_if101 == nil {
		initDetailFetcher()
	}

	temp_videoSources_if101 = buffer_videoSources_if101[0:0]

	detailC.Visit(fmt.Sprintf(if101DetailUrlFormat, anime.Id_if101))

	var temp_episodes_if101 = uint16(len(temp_videoSources_if101))

	if (anime.Episodes & 32767) == temp_episodes_if101 {
		fmt.Printf("此作品 ID:%d 在 if101 ID:%d 的資源並沒有更新 集數:%d\n", anime.Id, anime.Id_if101, temp_episodes_if101)

		return false
	}

	anime.Episodes = temp_episodes_if101

	if anime.Episodes != 0 {
		anime.Videos = make([]string, 0, anime.Episodes)

		sort.SliceStable(temp_videoSources_if101, func(i, j int) bool {
			if temp_videoSources_if101[i].isNumber && temp_videoSources_if101[j].isNumber {
				return temp_videoSources_if101[i].name.(float64) < temp_videoSources_if101[j].name.(float64)
			} else if temp_videoSources_if101[i].isNumber {
				return true
			} else if temp_videoSources_if101[j].isNumber {
				return false
			} else {
				return temp_videoSources_if101[i].name.(string) < temp_videoSources_if101[j].name.(string)
			}
		})

		for _, episode := range temp_videoSources_if101 {
			anime.Videos = append(anime.Videos, episode.video)
		}

		var last_Episode_if101 uint32 = 0

		counter := make([]uint32, 8)

		anime.ExEpisodes = make([]uint32, 0)

		for i, episode := range temp_videoSources_if101 {
			if episode.isNumber {
				f_episode := episode.name.(float64)
				C_episode := math.Trunc(f_episode)
				I_episode := uint32(C_episode)

				if f_episode != C_episode {
					anime.ExEpisodes = append(anime.ExEpisodes, (media.HALF32<<29)+(I_episode<<16)+uint32(i))
					continue
				} else if I_episode-last_Episode_if101 > 1 {
					if I_episode < last_Episode_if101 {
						log.Fatal(anime, temp_videoSources_if101, f_episode, last_Episode_if101)
					}
					anime.ExEpisodes = append(anime.ExEpisodes, (media.OFFSET32<<29)+((I_episode-last_Episode_if101)<<16)+uint32(i))
				} else if f_episode == 1 && i == len(anime.ExEpisodes) {
					anime.Episodes |= 1 << 15
				}

				last_Episode_if101 = I_episode

				continue
			}

			s_episode := episode.name.(string)

			if len(s_episode) >= 2 {
				if s_episode[:2] == "SP" {
					ok := false

					if len(s_episode) > 2 {
						if n_episode, err := strconv.Atoi(s_episode[2:]); err != nil {
							counter[media.SP32] = uint32(n_episode)
							ok = true
						}
					}

					if !ok {
						counter[media.SP32]++
					}

					anime.ExEpisodes = append(anime.ExEpisodes, (media.SP32<<29)+(counter[media.SP]<<16)+uint32(i))
					continue
				} else if s_episode[len(s_episode)-2:] == ".5" {
					n_episode, _ := strconv.Atoi(s_episode[:len(s_episode)-2])
					anime.ExEpisodes = append(anime.ExEpisodes, (media.HALF32<<29)+(uint32(n_episode)<<16)+uint32(i))
					continue
				}
			}

			if len(s_episode) >= 3 {
				if s_episode[:3] == "OVA" {
					ok := false

					if len(s_episode) > 3 {
						if n_episode, err := strconv.Atoi(s_episode[3:]); err != nil {
							counter[media.OVA] = uint32(n_episode)
							ok = true
						}
					}

					if !ok {
						counter[media.OVA]++
					}

					anime.ExEpisodes = append(anime.ExEpisodes, (media.OVA32<<29)+(counter[media.OVA]<<16)+uint32(i))
					continue
				} else if s_episode[:3] == "OAD" {
					ok := false

					if len(s_episode) > 3 {
						if n_episode, err := strconv.Atoi(s_episode[3:]); err != nil {
							counter[media.OAD] = uint32(n_episode)
							ok = true
						}
					}

					if !ok {
						counter[media.OAD]++
					}

					anime.ExEpisodes = append(anime.ExEpisodes, (media.OAD32<<29)+(counter[media.OAD]<<16)+uint32(i))
					continue
				}
			}

			if len(s_episode) >= 9 {
				if s_episode[:9] == "劇場版" || s_episode[:9] == "剧场版" {
					ok := false

					if len(s_episode) > 9 {
						if n_episode, err := strconv.Atoi(s_episode[9:]); err != nil {
							counter[media.MOVIE] = uint32(n_episode)
							ok = true
						}
					}

					if !ok {
						counter[media.MOVIE]++
					}

					anime.ExEpisodes = append(anime.ExEpisodes, (media.MOVIE32<<29)+(counter[media.MOVIE]<<16)+uint32(i))
					continue
				}
			}

			counter[media.OTHER]++
			anime.ExEpisodes = append(anime.ExEpisodes, (media.OTHER32<<29)+(counter[media.OTHER]<<16)+uint32(i))
			continue
		}

		fmt.Printf("作品ID:%d 更新Episodes:%v\n", anime.Id, anime.Episodes&32767)
	}

	return true
}
