package media

import (
	fastio "github.com/xLanStar/go-fast-io"
)

/*
讀取一個 Media 物件，可能是 Anime 或 Manga 型別

僅從緩衝區讀取，需自行開啟、關閉檔案
*/
func ReadMedia(fileReader *fastio.FileReader) IMedia {

	mediaType := MediaType(fileReader.ReadUint8())

	if mediaType == ANIME {
		anime := &Anime{}

		anime.Episodes = fileReader.ReadUint16()

		if anime.Episodes != 0 {
			anime.Videos = make([]string, anime.Episodes&32767)

			for i := 0; i != len(anime.Videos); i++ {
				anime.Videos[i] = fileReader.ReadString()
			}

			anime.ExEpisodes = fileReader.ReadUint32Array()
		}

		anime.Id_if101 = fileReader.ReadUint32()

		anime.Title = fileReader.ReadString()

		anime.Description = fileReader.ReadString()

		return anime
	} else if mediaType == NOVEL {
		novel := &Novel{}

		novel.Volumes = fileReader.ReadUint16()

		novel.Title = fileReader.ReadString()

		novel.Description = fileReader.ReadString()

		return novel
	} else {
		manga := &Manga{}

		manga.Volumes = fileReader.ReadUint32Array()

		manga.Id_cartoonmad = fileReader.ReadUint32()

		manga.Title = fileReader.ReadString()

		manga.Description = fileReader.ReadString()

		return manga
	}
}

// func ReadMedia_MIN(fileReader *fastio.FileReader) IMedia {

// 	mediaType := MediaType(fileReader.ReadUint8())

// 	if mediaType == ANIME {
// 		anime := &Anime{Type: mediaType}

// 		anime.Episodes = fileReader.ReadUint16()

// 		if anime.Episodes != 0 {
// 			anime.Videos = make([]string, anime.Episodes&32767)

// 			buffer := make([]byte, (anime.Episodes&32767)*64)

// 			fileReader.Reader.Read(buffer)

// 			for i := 0; i != len(anime.Videos); i++ {
// 				anime.Videos[i] = string(buffer[i*64 : (i+1)*64])
// 			}

// 			anime.ExEpisodes = fileReader.ReadUint32Array()
// 		}

// 		anime.Id_if101 = fileReader.ReadUint32()

// 		anime.Title = fileReader.ReadString()

// 		anime.Description = fileReader.ReadString()

// 		return anime
// 	} else if mediaType == NOVEL {
// 		novel := &Novel{Type: mediaType}

// 		novel.Volumes = fileReader.ReadUint16()

// 		novel.Title = fileReader.ReadString()

// 		novel.Description = fileReader.ReadString()

// 		return novel
// 	} else {
// 		manga := &Manga{Type: mediaType}

// 		manga.Volumes = fileReader.ReadUint32Array()

// 		manga.Id_cartoonmad = fileReader.ReadUint32()

// 		manga.Title = fileReader.ReadString()

// 		manga.Description = fileReader.ReadString()

// 		return manga
// 	}
// }

/*
寫入一個 Media 物件，可接受 Anime 及 Manga 型別

僅寫入至緩衝區，需自行開啟、關閉檔案
*/
func WriteMedia_MIN(fileWriter *fastio.FileWriter, media IMedia) {
	fileWriter.Write(byte(media.GetType()))

	if media.GetType() == ANIME {
		fileWriter.WriteUint16(media.(*Anime).GetEpisodes())

		if media.(*Anime).Episodes != 0 {
			temp := fileWriter.Buffer_p[0:0]

			for _, video := range media.(*Anime).Videos {
				temp = append(temp, video...)
			}

			fileWriter.Buffer_p = fileWriter.Buffer_p[len(temp):]

			fileWriter.WriteUint32Array(media.(*Anime).GetExEpisodes())
		}

		fileWriter.WriteUint32(media.(*Anime).GetId_if101())
	} else if media.GetType() == NOVEL {
		fileWriter.WriteUint16(media.(*Novel).GetVolumes())
	} else {
		fileWriter.WriteUint32Array(media.(*Manga).GetVolumes())

		fileWriter.WriteUint32(media.(*Manga).GetId_cartoonmad())
	}

	fileWriter.WriteString(media.GetTitle())

	fileWriter.WriteString(media.GetDescription())
}

func WriteMedia(fileWriter *fastio.FileWriter, media IMedia) {
	fileWriter.Write(byte(media.GetType()))

	if media.GetType() == ANIME {
		fileWriter.WriteUint16(media.(*Anime).GetEpisodes())

		if media.(*Anime).Episodes != 0 {
			for _, video := range media.(*Anime).Videos {
				fileWriter.WriteString(video)
			}

			fileWriter.WriteUint32Array(media.(*Anime).GetExEpisodes())
		}

		fileWriter.WriteUint32(media.(*Anime).GetId_if101())
	} else if media.GetType() == NOVEL {
		fileWriter.WriteUint16(media.(*Novel).GetVolumes())
	} else {
		fileWriter.WriteUint32Array(media.(*Manga).GetVolumes())

		fileWriter.WriteUint32(media.(*Manga).GetId_cartoonmad())
	}

	fileWriter.WriteString(media.GetTitle())

	fileWriter.WriteString(media.GetDescription())
}
