package media

import (
	"fmt"

	fastio "github.com/xLanStar/go-fast-io"
)

type Manga struct {
	Id           uint32   `json:"-"`
	Title        string   `json:"title,omitempty"`
	Description  string   `json:"description,omitempty"`
	Volumes      []uint32 `json:"volumes,omitempty"`
	CartoonmadId uint32   `json:"id_cartoonmad,omitempty"`
}

func (media *Manga) Test() {
	fmt.Println("MANGA")
}

func (media *Manga) String() string {
	return fmt.Sprintf("\n  [作品] ID:%6d 類型:%s\n  - 標題:%s\n  - 簡介:%s\n  - 動漫狂ID:%6d", media.Id, MANGA, media.Title, media.Description, media.CartoonmadId)
}

func (media *Manga) GetId() uint32 {
	return media.Id
}

func (media *Manga) GetType() MediaType {
	return MANGA
}

func (media *Manga) GetTitle() string {
	return media.Title
}

func (media *Manga) GetDescription() string {
	return media.Description
}

func (media *Manga) SetId(id uint32) {
	media.Id = id
}

func (media *Manga) SetTitle(title string) {
	media.Title = title
}

func (media *Manga) SetDescription(description string) {
	media.Description = description
}

func (manga *Manga) GetVolumes() []uint32 {
	return manga.Volumes
}

func (manga *Manga) GetCartoonmadId() uint32 {
	return manga.CartoonmadId
}

func (manga *Manga) Write(fileWriter *fastio.FileWriter) {

	fileWriter.Write(byte(manga.GetType()))

	fileWriter.WriteUint32Array(manga.GetVolumes())

	fileWriter.WriteUint32(manga.GetCartoonmadId())

	fileWriter.WriteString(manga.GetTitle())

	fileWriter.WriteString(manga.GetDescription())
}
