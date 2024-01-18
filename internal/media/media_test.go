package media

import (
	"os"
	"testing"

	fastio "github.com/xLanStar/go-fast-io"
)

func TestMediaType(t *testing.T) {
	anime := &Anime{Id: 1234, Description: "1234", Title: "1234"}
	manga := &Manga{Id: 1234, Description: "1234", Title: "1234"}
	novel := &Novel{Id: 1234, Description: "1234", Title: "1234"}

	var fileWriter fastio.FileWriter
	fileWriter.Init()
	fileWriter.OpenFile("test", os.O_WRONLY|os.O_CREATE, 0466)
	anime.Write(&fileWriter)
	manga.Write(&fileWriter)
	novel.Write(&fileWriter)
	fileWriter.Flush()
}

func Benchmark_NormalSwitch(b *testing.B) {
	var fileWriter fastio.FileWriter
	fileWriter.Init()
	// fileWriter.OpenFile("test1", os.O_WRONLY|os.O_CREATE, 0466)
	anime := &Anime{Id: 1234, Description: "1234", Title: "1234"}
	manga := &Manga{Id: 1234, Description: "1234", Title: "1234"}
	novel := &Novel{Id: 1234, Description: "1234", Title: "1234"}

	for i := 0; i < b.N; i++ {
		anime.Write(&fileWriter)
		manga.Write(&fileWriter)
		novel.Write(&fileWriter)
		anime.Write(&fileWriter)
		manga.Write(&fileWriter)
		novel.Write(&fileWriter)
		fileWriter.Flush()
	}
}

func Benchmark_TypeSwitch(b *testing.B) {
	var fileWriter fastio.FileWriter
	fileWriter.Init()
	anime := &Anime{Id: 1234, Description: "1234", Title: "1234"}
	manga := &Manga{Id: 1234, Description: "1234", Title: "1234"}
	novel := &Novel{Id: 1234, Description: "1234", Title: "1234"}

	for i := 0; i < b.N; i++ {
		WriteMedia(&fileWriter, anime)
		WriteMedia(&fileWriter, manga)
		WriteMedia(&fileWriter, novel)
		WriteMedia(&fileWriter, anime)
		WriteMedia(&fileWriter, manga)
		WriteMedia(&fileWriter, novel)
		fileWriter.Flush()
	}
}
