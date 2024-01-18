package media

import (
	"fmt"

	fastio "github.com/xLanStar/go-fast-io"
)

type Novel struct {
	Id          uint32 `json:"-"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Volumes     uint16 `json:"volumes,omitempty"`
}

func (media *Novel) Test() {
	fmt.Println("NOVEL")
}

func (media *Novel) String() string {
	return fmt.Sprintf("\n  [作品] ID:%6d 類型:%s\n  - 標題:%s\n  - 簡介:%s", media.Id, NOVEL, media.Title, media.Description)
}

func (media *Novel) GetId() uint32 {
	return media.Id
}

func (media *Novel) GetType() MediaType {
	return NOVEL
}

func (media *Novel) GetTitle() string {
	return media.Title
}

func (media *Novel) GetDescription() string {
	return media.Description
}

func (media *Novel) SetId(id uint32) {
	media.Id = id
}

func (media *Novel) SetTitle(title string) {
	media.Title = title
}

func (media *Novel) SetDescription(description string) {
	media.Description = description
}

func (novel *Novel) GetVolumes() uint16 {
	return novel.Volumes
}

func (novel *Novel) Write(fileWriter *fastio.FileWriter) {

	fileWriter.Write(byte(novel.GetType()))

	fileWriter.WriteUint16(novel.GetVolumes())

	fileWriter.WriteString(novel.GetTitle())

	fileWriter.WriteString(novel.GetDescription())
}
