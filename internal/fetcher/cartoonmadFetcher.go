package fetcher

import (
	Media "Ani-Server/internal/media"
	"fmt"
	"strconv"

	"github.com/gocolly/colly"
)

var (
	cartoonmadDetailUrlFormat string
	cartoonmadDetailC         *colly.Collector
	cartoonmadDetailResult    []uint32
)

func initCartoonmadDetail() {
	cartoonmadDetailUrlFormat = "https://www.cartoonmad.com/comic/%d.html"
	cartoonmadDetailC = colly.NewCollector()

	cartoonmadDetailC.OnHTML("table[width='800']", func(c *colly.HTMLElement) {
		var first uint32 = 0
		var last uint32 = 65535

		c.ForEach("a[href]", func(_ int, c *colly.HTMLElement) {
			page, _ := strconv.Atoi(c.Text[3 : len(c.Text)-3])

			if last == 65535 {
				first = uint32(page)
			} else if uint32(page) != last+1 {
				cartoonmadDetailResult = append(cartoonmadDetailResult, (first<<16)+last)
				first = uint32(page)
			}

			last = uint32(page)
		})
		if last != 65535 {
			cartoonmadDetailResult = append(cartoonmadDetailResult, (first<<16)+last)
		}
	})
}

func FetchCartoonmadDetail(media *Media.Manga) error {
	if media == nil || media.CartoonmadId == 0 {
		return nil
	}

	cartoonmadDetailResult = make([]uint32, 0, 100)

	if cartoonmadDetailC == nil {
		initCartoonmadDetail()
	}

	fmt.Println(fmt.Sprintf(cartoonmadDetailUrlFormat, media.CartoonmadId))
	cartoonmadDetailC.Visit(fmt.Sprintf(cartoonmadDetailUrlFormat, media.CartoonmadId))

	if len(cartoonmadDetailResult) != 0 {
		media.Volumes = make([]uint32, len(cartoonmadDetailResult))
		for i := 0; i != len(cartoonmadDetailResult); i++ {
			media.Volumes[i] = uint32(cartoonmadDetailResult[i])
		}
	}
	fmt.Println(cartoonmadDetailResult)

	return nil
}
