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
	IdStr    string
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
	bufferSearchLinksIf101 []If101Link
	tempSearchLinksIf101   []If101Link

	// 更新 Update
	bufferUpdateLinksIf101 []If101Link
	tempUpdateLinksIf101   []If101Link

	// 資訊 Detail
	bufferVideoSourcesIf101 []videoSource
	tempVideoSourcesIf101   []videoSource
)

func initUpdateFetcher() {
	bufferUpdateLinksIf101 = make([]If101Link, 100)

	// if101 取得資訊
	updateC = colly.NewCollector()
	updateC.OnHTML(".stui-vodlist .clearfix", func(e *colly.HTMLElement) {
		sUpdateAt := e.ChildText("span.time")

		if len(sUpdateAt) < 16 {
			return
		}

		sIf101Id := e.ChildAttr("a[title]", "href")

		var updateAt uint32 = 0

		for _, c := range sUpdateAt[:13] {
			if '0' <= c && c <= '9' {
				updateAt = (updateAt << 3) + (updateAt << 1) + uint32(c-'0')
			}
		}

		var link If101Link = If101Link{
			IdStr:    sIf101Id[25 : len(sIf101Id)-5],
			Name:     e.ChildAttr("a[title]", "title"),
			UpdateAt: updateAt,
		}

		tempUpdateLinksIf101 = append(tempUpdateLinksIf101, link)
	})
}

func InitSearchFetcher() {
	bufferSearchLinksIf101 = make([]If101Link, 100)

	// if101 取得資訊
	searchC = colly.NewCollector()
	searchC.OnHTML(".stui-vodlist .clearfix", func(e *colly.HTMLElement) {
		sUpdateAt := e.ChildText("span.time")

		if len(sUpdateAt) < 16 {
			return
		}

		sIf101Id := e.ChildAttr("a[title]", "href")

		var updateAt uint32 = 0

		for _, c := range sUpdateAt[:13] {
			if '0' <= c && c <= '9' {
				updateAt = (updateAt << 3) + (updateAt << 1) + uint32(c-'0')
			}
		}

		var link If101Link = If101Link{
			IdStr:    sIf101Id[25 : len(sIf101Id)-5],
			Name:     e.ChildAttr("a[title]", "title"),
			UpdateAt: updateAt,
		}

		tempSearchLinksIf101 = append(tempSearchLinksIf101, link)
	})
}
func initDetailFetcher() {
	bufferVideoSourcesIf101 = make([]videoSource, 1200)

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

			tempVideoSourcesIf101 = append(tempVideoSourcesIf101, video)
		})
	})
}

// 搜尋指定的關鍵字，並回傳所有的連結
func SearchIf101Links(search string) []If101Link {
	if len(bufferSearchLinksIf101) == 0 {
		fmt.Println("if101 init")
		InitSearchFetcher()
	}

	tempSearchLinksIf101 = bufferUpdateLinksIf101[0:0]

	searchC.Visit(fmt.Sprintf(if101SearchUrlFormat, search))

	return tempSearchLinksIf101
}

// 以指定的時間為基準，將所有此時間點以後的 if101 id 傳回
func GetUpdatedIds(lastUpdateAt uint32) []uint32 {
	if len(bufferUpdateLinksIf101) == 0 {
		initUpdateFetcher()
	}

	var result []uint32 = make([]uint32, 0, 50)

	for page, breakpoint := 1, false; !breakpoint; page++ {
		tempUpdateLinksIf101 = bufferUpdateLinksIf101[0:0]

		updateC.Visit(fmt.Sprintf(if101UpdateUrlFormat, page))

		for _, link := range tempUpdateLinksIf101 {
			if link.UpdateAt < lastUpdateAt {
				breakpoint = true
				continue
			}

			idIf101, _ := strconv.Atoi(link.IdStr)

			result = append(result, uint32(idIf101))
		}
	}

	return result
}

// 基於以建置之 Media，填入 if101 資源
// 回傳 (是否有更新, 錯誤)
func FetchIf101Details(anime *media.Anime) bool {
	if anime.If101Id == 0 {
		return false
	}

	if bufferVideoSourcesIf101 == nil {
		initDetailFetcher()
	}

	tempVideoSourcesIf101 = bufferVideoSourcesIf101[0:0]

	detailC.Visit(fmt.Sprintf(if101DetailUrlFormat, anime.If101Id))

	var tempEpisodesIf101 = uint16(len(tempVideoSourcesIf101))

	if (anime.Episodes & 32767) == tempEpisodesIf101 {
		fmt.Printf("此作品 ID:%d 在 if101 ID:%d 的資源並沒有更新 集數:%d\n", anime.Id, anime.If101Id, tempEpisodesIf101)

		return false
	}

	anime.Episodes = tempEpisodesIf101

	if anime.Episodes != 0 {
		anime.Videos = make([]string, 0, anime.Episodes)

		sort.SliceStable(tempVideoSourcesIf101, func(i, j int) bool {
			if tempVideoSourcesIf101[i].isNumber && tempVideoSourcesIf101[j].isNumber {
				return tempVideoSourcesIf101[i].name.(float64) < tempVideoSourcesIf101[j].name.(float64)
			} else if tempVideoSourcesIf101[i].isNumber {
				return true
			} else if tempVideoSourcesIf101[j].isNumber {
				return false
			} else {
				return tempVideoSourcesIf101[i].name.(string) < tempVideoSourcesIf101[j].name.(string)
			}
		})

		for _, episode := range tempVideoSourcesIf101 {
			anime.Videos = append(anime.Videos, episode.video)
		}

		var lastEpisodeIf101 uint32 = 0

		counter := make([]uint32, 8)

		anime.ExEpisodes = make([]uint32, 0)

		for i, episode := range tempVideoSourcesIf101 {
			if episode.isNumber {
				fEpisode := episode.name.(float64)
				CEpisode := math.Trunc(fEpisode)
				IEpisode := uint32(CEpisode)

				if fEpisode != CEpisode {
					anime.ExEpisodes = append(anime.ExEpisodes, (media.HALF32<<29)+(IEpisode<<16)+uint32(i))
					continue
				} else if IEpisode-lastEpisodeIf101 > 1 {
					if IEpisode < lastEpisodeIf101 {
						log.Fatal(anime, tempVideoSourcesIf101, fEpisode, lastEpisodeIf101)
					}
					anime.ExEpisodes = append(anime.ExEpisodes, (media.OFFSET32<<29)+((IEpisode-lastEpisodeIf101)<<16)+uint32(i))
				} else if fEpisode == 1 && i == len(anime.ExEpisodes) {
					anime.Episodes |= 1 << 15
				}

				lastEpisodeIf101 = IEpisode

				continue
			}

			sEpisode := episode.name.(string)

			if len(sEpisode) >= 2 {
				if sEpisode[:2] == "SP" {
					ok := false

					if len(sEpisode) > 2 {
						if nEpisode, err := strconv.Atoi(sEpisode[2:]); err != nil {
							counter[media.SP32] = uint32(nEpisode)
							ok = true
						}
					}

					if !ok {
						counter[media.SP32]++
					}

					anime.ExEpisodes = append(anime.ExEpisodes, (media.SP32<<29)+(counter[media.SP]<<16)+uint32(i))
					continue
				} else if sEpisode[len(sEpisode)-2:] == ".5" {
					nEpisode, _ := strconv.Atoi(sEpisode[:len(sEpisode)-2])
					anime.ExEpisodes = append(anime.ExEpisodes, (media.HALF32<<29)+(uint32(nEpisode)<<16)+uint32(i))
					continue
				}
			}

			if len(sEpisode) >= 3 {
				if sEpisode[:3] == "OVA" {
					ok := false

					if len(sEpisode) > 3 {
						if nEpisode, err := strconv.Atoi(sEpisode[3:]); err != nil {
							counter[media.OVA] = uint32(nEpisode)
							ok = true
						}
					}

					if !ok {
						counter[media.OVA]++
					}

					anime.ExEpisodes = append(anime.ExEpisodes, (media.OVA32<<29)+(counter[media.OVA]<<16)+uint32(i))
					continue
				} else if sEpisode[:3] == "OAD" {
					ok := false

					if len(sEpisode) > 3 {
						if nEpisode, err := strconv.Atoi(sEpisode[3:]); err != nil {
							counter[media.OAD] = uint32(nEpisode)
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

			if len(sEpisode) >= 9 {
				if sEpisode[:9] == "劇場版" || sEpisode[:9] == "剧场版" {
					ok := false

					if len(sEpisode) > 9 {
						if nEpisode, err := strconv.Atoi(sEpisode[9:]); err != nil {
							counter[media.MOVIE] = uint32(nEpisode)
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
