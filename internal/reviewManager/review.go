package reviewManager

import "fmt"

type Review struct {
	Id       uint32  `json:"id"`
	AuthorId uint32  `json:"authorId"`
	Author   *string `json:"author"`
	Rank     Rank    `json:"rank"`
	Content  string  `json:"content"`
	Like     uint32  `json:"like"`
}

func (review *Review) String() string {
	return fmt.Sprintf("\n  [評論] ID:%04d\n  - 作者:%-10s 評價:%d 喜歡人數:%02d\n  - 內容:%s", review.Id, *review.Author, review.Rank, review.Like, review.Content)
	// return fmt.Sprintf("評論[ID:%d 作者:%s 評價:%d 內容:%s 喜歡人數:%d]", review.Id, *review.Author, review.Rank, review.Content, review.Like)
}
