package media

import fastio "github.com/xLanStar/go-fast-io"

type MediaType uint8

const (
	ANIME MediaType = iota
	MANGA
	NOVEL
	UNKNOWN MediaType = 255
)

var (
	MediaTypes []string = []string{"動畫", "漫畫", "小說"}
)

func (mediaType MediaType) String() string {
	return MediaTypes[mediaType]
}

type IMedia interface {
	GetId() uint32
	GetType() MediaType
	GetTitle() string
	GetDescription() string
	SetId(uint32)
	SetTitle(string)
	SetDescription(string)
	Write(*fastio.FileWriter)
	Test()
}
