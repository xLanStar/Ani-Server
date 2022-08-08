package mediaManager

import (
	"Ani-Server/internal/fetcher"
	Media "Ani-Server/internal/media"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	fastio "github.com/xLanStar/go-fast-io"
)

var (
	idMap map[uint32]Media.IMedia = make(map[uint32]Media.IMedia, 7000)

	change map[uint32]bool = make(map[uint32]bool, 10)

	mediaFolder string
)

func Load(_mediaFolder string) {
	mediaFolder = _mediaFolder

	var fileReader fastio.FileReader

	fileReader.Init()

	files, _ := ioutil.ReadDir(mediaFolder)

	for _, f := range files {
		fileReader.OpenFile(mediaFolder+f.Name(), os.O_RDONLY, 0666)

		mediaId, _ := strconv.Atoi(f.Name())

		media := Media.ReadMedia(&fileReader)

		media.SetId(uint32(mediaId))

		idMap[uint32(mediaId)] = media

		fileReader.Close()
	}

	fmt.Printf("[MediaManager] 共有 %d 個作品\n", len(idMap))
}

func Save() {
	fmt.Println("[MediaManager] 存檔")

	var fileWriter fastio.FileWriter
	fileWriter.Init()

	for id, changed := range change {

		if !changed {
			continue
		}

		media := idMap[id]

		fmt.Printf("[MediaManager] 保存作品ID: %d%s\n", id, media)

		fileWriter.OpenFile(mediaFolder+strconv.Itoa(int(id)), os.O_CREATE|os.O_WRONLY, 0666)

		media.Write(&fileWriter)

		fileWriter.Close()
	}

	change = make(map[uint32]bool)
}

// 從 anilist 抓取
func FetchMediaById(mediaId uint32) (Media.IMedia, error) {
	media, err := fetcher.FetchMediaById(mediaId)

	if err != nil {
		fmt.Println("從 anilist 抓取失敗...")
		return nil, err
	}

	change[mediaId] = true

	// read media
	idMap[mediaId] = media

	return media, nil
}

func UpdateIf101(lastUpdateAt uint32) {
	if lastUpdateAt == 0 {
		log.Println("尚未指定最新資料時間點，將不進行更新")
		return
	}

	fmt.Println("[MediaManager] 即將進行更新if101...")

	var if101Map map[uint32]uint32 = make(map[uint32]uint32, len(idMap))

	for mediaId, media := range idMap {
		if media.GetType() == Media.ANIME {
			if101Map[media.(*Media.Anime).Id_if101] = mediaId
		}
	}

	var updatedIds_if101 []uint32 = fetcher.GetUpdatedIds(lastUpdateAt)

	for _, id_if101 := range updatedIds_if101 {
		if mediaId, ok := if101Map[id_if101]; ok {
			changed := fetcher.FetchIf101Details(idMap[mediaId].(*Media.Anime))

			if changed {
				change[mediaId] = true
			}
		}
	}

	fmt.Println("[MediaManager] 更新完畢!")
}

func GetSimpleMediaInfo(mediaIds []uint32) (map[uint32]string, []uint32) {
	if len(mediaIds) == 0 {
		return nil, nil
	}

	titles := make(map[uint32]string)
	hasResources := make([]uint32, 0, 40)

	for _, mediaId := range mediaIds {
		if media, ok := idMap[mediaId]; ok {
			titles[mediaId] = media.GetTitle()
			if media.GetType() == Media.ANIME {
				if media.(*Media.Anime).Episodes != 0 {
					hasResources = append(hasResources, mediaId)
				}
			} else if media.GetType() == Media.MANGA {
				if len(media.(*Media.Manga).Volumes) != 0 {
					hasResources = append(hasResources, mediaId)
				}
			}
		}
	}

	return titles, hasResources
}

func HasMediaId(mediaId uint32) bool {
	_, ok := idMap[mediaId]
	return ok
}

func GetMediaById(mediaId uint32) Media.IMedia {
	return idMap[mediaId]
}

func EditMedia(mediaId uint32, mediaType Media.MediaType, data map[string]interface{}) {
	fmt.Printf("[MediaManager] 編輯作品ID:%6d  類型:%s\n", mediaId, mediaType)

	// 建置 / 取得作品
	if _, ok := idMap[mediaId]; !ok {
		fmt.Printf("    作品不存在，即將建置一個作品類型:%v\n", mediaType)

		// 嘗試用 anilist
		_, err := FetchMediaById(mediaId)

		if err != nil {
			// 系統手動建置資料
			if mediaType == Media.ANIME {
				idMap[mediaId] = &Media.Anime{Id: mediaId}
			} else if mediaType == Media.MANGA {
				idMap[mediaId] = &Media.Manga{Id: mediaId}
			} else if mediaType == Media.NOVEL {
				idMap[mediaId] = &Media.Novel{Id: mediaId}
			}
		}
	}

	if idMap[mediaId].GetType() != mediaType {
		fmt.Printf("    發現作品類型錯誤，即將修正作品類型:%v\n", idMap[mediaId].GetType())
		if mediaType == Media.ANIME {
			idMap[mediaId] = &Media.Anime{Id: mediaId, Title: idMap[mediaId].GetTitle(), Description: idMap[mediaId].GetDescription()}
		} else if mediaType == Media.MANGA {
			idMap[mediaId] = &Media.Manga{Id: mediaId, Title: idMap[mediaId].GetTitle(), Description: idMap[mediaId].GetDescription()}
		} else if mediaType == Media.NOVEL {
			idMap[mediaId] = &Media.Novel{Id: mediaId, Title: idMap[mediaId].GetTitle(), Description: idMap[mediaId].GetDescription()}
		}
	}

	// 編輯作品
	var media Media.IMedia = idMap[mediaId]

	fmt.Println(media)

	if title, ok := data["title"]; ok {
		fmt.Printf("    [標題]\n    %s\n  ->%s\n", media.GetTitle(), title.(string))
		media.SetTitle(strings.Trim(title.(string), "\r\n "))
	}

	if description, ok := data["description"]; ok {
		fmt.Printf("    [簡介]\n    %s\n  ->%s\n", media.GetDescription(), description.(string))
		media.SetDescription(strings.Trim(strings.Replace(description.(string), "\r\n", "\n", -1), "\r\n "))
	}

	if media.GetType() == Media.ANIME {

		anime := media.(*Media.Anime)

		if id_if101, ok := data["id_if101"]; ok {
			fmt.Printf("    [ID_if101]\n    %6d\n  ->%6d\n", anime.Id_if101, uint32(id_if101.(float64)))

			fmt.Println("原本:", anime)

			anime.Id_if101 = uint32(id_if101.(float64))
			anime.Videos = make([]string, 0)
			anime.Episodes = 0
			anime.ExEpisodes = make([]uint32, 0)

			if anime.Id_if101 != 0 {
				changed := fetcher.FetchIf101Details(anime)

				if !changed {
					fmt.Println("作品並沒有更新if101detail")
				} else {
					fmt.Println("更新後:", anime)
				}
			}
		}
	} else if media.GetType() == Media.NOVEL {
		// TODO: NOVEL INFORMATIONS
	} else {

		manga := media.(*Media.Manga)

		if id_cartoonmad, ok := data["id_cartoonmad"]; ok {
			fmt.Printf("    [ID_cartoonmad]\n    %6d\n  ->%6d\n", manga.Id_cartoonmad, uint32(id_cartoonmad.(float64)))
			manga.Id_cartoonmad = uint32(id_cartoonmad.(float64))
			manga.Volumes = make([]uint32, 0)

			if manga.Id_cartoonmad != 0 {
				fetcher.FetchCartoonmadDetail(manga)
			}
		}
	}

	change[mediaId] = true
}

// func EditMediaDescription(id uint32, mediaType Media.MediaType, description string) {
// 	fmt.Printf("[MediaManager] 修改作品ID:%7d 簡介:%s\n", id, description)
// 	//TODO:LOG

// 	if media, ok := idMap[id]; ok {
// 		media.SetDescription(description)
// 	} else {
// 		if mediaType == Media.ANIME {
// 			anime := &Media.Anime{Media: Media.Media{Id: id, Type: mediaType, Description: description}}
// 			idMap[id] = anime
// 		} else if mediaType == Media.MANGA {
// 			manga := &Media.Anime{Media: Media.Media{Id: id, Type: mediaType, Description: description}}
// 			idMap[id] = manga
// 		}
// 	}

// 	change[id] = true
// }
