package fetcher

import (
	"Ani-Server/internal/media"
	"encoding/json"
	"fmt"
	"strings"
)

// SECTION_TYPE: RAWMEDIA
type RawMedia struct {
	Id          int
	Title       map[string]string
	Description string
	Type        string
	Format      string
	UpdatedAt   int
}

func (rawMedia *RawMedia) ToMedia() media.IMedia {
	if rawMedia.Type == "ANIME" {
		return &media.Anime{Id: uint32(rawMedia.Id), Title: rawMedia.Title["chinese"], Description: rawMedia.Description, Episodes: 0, If101Id: 0}
	} else if rawMedia.Format == "NOVEL" {
		return &media.Novel{Id: uint32(rawMedia.Id), Title: rawMedia.Title["chinese"], Description: rawMedia.Description}
	} else {
		return &media.Manga{Id: uint32(rawMedia.Id), Title: rawMedia.Title["chinese"], Description: rawMedia.Description, CartoonmadId: 0}
	}
}

var (
	fetchPoster                   Poster
	searchPoster                  Poster
	fetchMediasSortedIdPoster     Poster
	fetchMediasSortedUpdatePoster Poster
)

// SECTION_TYPE: 初始化POSTER
func initSearchPoster() {
	// 輸入 $search, $seasonYear, $season
	var searchQuery string = strings.Trim(`
	query ($search: String, $seasonYear: Int, $season: MediaSeason) {
		Media(search: $search, type: ANIME, seasonYear: $seasonYear, season : $season) {
			id
			title {
				native
			}
			type
			format
		}
	}`, "\n\r ")
	searchPoster = Poster{Url: "https://trace.moe/anilist/", Data: Data{Query: searchQuery, Variables: make(map[string]interface{})}}
}

func initFetchPoster() {
	// 輸入 $id, $seasonYear, $season
	var fetchQuery string = strings.Trim(`
	query ($id: Int) {
		Media(id: $id) {
			id
			title {
				native
			}
			type
			format
		}
	}`, "\n\r ")
	fetchPoster = Poster{Url: "https://trace.moe/anilist/", Data: Data{Query: fetchQuery, Variables: make(map[string]interface{})}}
}

func initFetchMediasSortedIdPoster() {
	// 輸入 $page, 以 id 排序
	var fetchIdQuery string = strings.Trim(`
	query ($page: Int) {
		Page (page: $page, perPage: 50) {
			media (sort: ID) {
				id
				title {
					native
				}
				type
				format
			}
		}
	}`, "\n\r ")
	fetchMediasSortedIdPoster = Poster{Url: "https://trace.moe/anilist/", Data: Data{Query: fetchIdQuery, Variables: make(map[string]interface{})}}
}

func initFetchMediasSortedUpdatePoster() {
	// 輸入 $page, 以 updateAt 排序
	var fetchUpdateQuery string = strings.Trim(`
	query ($page: Int, $perPage: Int) {
		Page (page: $page, perPage: $perPage) {
			media (sort: UPDATED_AT_DESC) {
				id
				title {
					native
				}
				type
				format
				updatedAt
			}
		}
	}`, "\n\r ")
	fetchMediasSortedUpdatePoster = Poster{Url: "https://trace.moe/anilist/", Data: Data{Query: fetchUpdateQuery, Variables: make(map[string]interface{})}}
}

// SECTION_TYPE: SearchMedia
func SearchMedia(title string, seasonYear int, season string) (media.IMedia, error) {
	if searchPoster.Url == "" {
		fmt.Println("Init Search")
		initSearchPoster()
	}
	searchPoster.Data.Variables["search"] = title
	searchPoster.Data.Variables["season"] = season
	searchPoster.Data.Variables["seasonYear"] = seasonYear

	var res struct {
		Data struct {
			Media RawMedia
		} `json:"data"`
	}

	bytes, err := searchPoster.Post()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	json.Unmarshal(bytes, &res)

	return res.Data.Media.ToMedia(), nil
}

// SECTION_TYPE: FetchMediaById
func FetchMediaById(mediaId uint32) (media.IMedia, error) {
	if fetchPoster.Url == "" {
		initFetchPoster()
	}
	fetchPoster.Data.Variables["id"] = mediaId

	var res struct {
		Data struct {
			Media RawMedia
		} `json:"data"`
	}

	bytes, err := fetchPoster.Post()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	err = json.Unmarshal(bytes, &res)

	if err != nil {
		return nil, err
	}

	return res.Data.Media.ToMedia(), nil
}

// SECTION_TYPE: FetchMediasSortedById
func FetchMediasSortedById(page int) ([]media.IMedia, error) {
	if fetchMediasSortedIdPoster.Url == "" {
		initFetchMediasSortedIdPoster()
	}
	fetchMediasSortedIdPoster.Data.Variables["page"] = page

	var res struct {
		Data struct {
			Page struct {
				Media []RawMedia `json:"media"`
			}
		} `json:"data"`
	}

	bytes, err := fetchMediasSortedIdPoster.Post()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	err = json.Unmarshal(bytes, &res)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	mRes := make([]media.IMedia, len(res.Data.Page.Media))

	for i := 0; i != len(mRes); i++ {
		mRes[i] = res.Data.Page.Media[i].ToMedia()
	}

	return mRes, nil
}

// SECTION_TYPE: FetchMediasSortedByUpdate
func FetchMediasSortedByUpdate(page int) ([]media.IMedia, error) {
	if fetchMediasSortedUpdatePoster.Url == "" {
		initFetchMediasSortedUpdatePoster()
	}
	fetchMediasSortedUpdatePoster.Data.Variables["page"] = page

	var res struct {
		Data struct {
			Page struct {
				Media []RawMedia `json:"media"`
			}
		} `json:"data"`
	}

	bytes, err := fetchMediasSortedUpdatePoster.Post()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	json.Unmarshal(bytes, &res)

	return nil, nil
}

// SECTION_TYPE: 釋放記憶體空間
func DisposeAnilistPoster() {
	searchPoster = Poster{}
	fetchMediasSortedIdPoster = Poster{}
	fetchMediasSortedUpdatePoster = Poster{}
	fetchPoster = Poster{}
}
