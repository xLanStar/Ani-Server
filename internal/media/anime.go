package media

import (
	"fmt"

	fastio "github.com/xLanStar/go-fast-io"
)

type Anime struct {
	Id          uint32   `json:"-"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Episodes    uint16   `json:"episodes,omitempty"`
	Videos      []string `json:"videos,omitempty"`
	ExEpisodes  []uint32 `json:"exepisodes,omitempty"`
	Id_if101    uint32   `json:"id_if101,omitempty"`
}

func (media *Anime) Test() {
	fmt.Println("ANIME")
}

func (media *Anime) String() string {
	return fmt.Sprintf("\n  [作品] ID:%6d 類型:%s\n  - 標題:%s\n  - 簡介:%s\n  - if101ID:%6d\n  - Videos:%v\n  - ExEpisodes:%v", media.Id, ANIME, media.Title, media.Description, media.Id_if101, media.Videos, media.ExEpisodes)
}

func (media *Anime) GetId() uint32 {
	return media.Id
}

func (media *Anime) GetType() MediaType {
	return ANIME
}

func (media *Anime) GetTitle() string {
	return media.Title
}

func (media *Anime) GetDescription() string {
	return media.Description
}

func (media *Anime) SetId(id uint32) {
	media.Id = id
}

func (media *Anime) SetTitle(title string) {
	media.Title = title
}

func (media *Anime) SetDescription(description string) {
	media.Description = description
}

func (anime *Anime) GetEpisodes() uint16 {
	return anime.Episodes
}

func (anime *Anime) GetVideos() []string {
	return anime.Videos
}

func (anime *Anime) GetExEpisodes() []uint32 {
	return anime.ExEpisodes
}

func (anime *Anime) GetId_if101() uint32 {
	return anime.Id_if101
}

func (anime *Anime) Write(fileWriter *fastio.FileWriter) {

	fileWriter.Write(byte(anime.GetType()))

	fileWriter.WriteUint16(anime.GetEpisodes())

	if anime.Episodes != 0 {
		for _, video := range anime.Videos {
			fileWriter.WriteString(video)
		}

		fileWriter.WriteUint32Array(anime.GetExEpisodes())
	}

	fileWriter.WriteString(anime.GetTitle())

	fileWriter.WriteString(anime.GetDescription())
}
