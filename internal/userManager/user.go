package userManager

import "fmt"

type User struct {
	Id              uint32          `json:"id"`
	Account         string          `json:"-"`
	Password        string          `json:"-"`
	UserName        string          `json:"userName"`
	Introduction    string          `json:"introduction"`
	LikeMedias      map[uint32]bool `json:"-"`
	LikeReviews     map[uint32]bool `json:"-"`
	LikeMediaCount  uint32          `json:"-"`
	LikeReviewCount uint32          `json:"-"`
	WatchedMedias   map[uint32]bool `json:"-"`
}

func (user User) String() string {
	return fmt.Sprintf("\n  [帳號] ID:%04d\n  - 帳號:%-10s 密碼:%-10s\n  - 名稱:%-10s\n  - 喜歡作品:%v\n  - 喜歡評論:%v", user.Id, user.Account, user.Password, user.UserName, user.LikeMedias, user.LikeReviews)
}

func (user *User) EditIntroduction(introduction string) {
	user.Introduction = introduction

	change[user.Id] = true
}

func (user *User) EditLikeMedia(id uint32, like bool) {
	if like {
		user.LikeMedias[id] = true
		user.LikeMediaCount++
	} else {
		delete(user.LikeMedias, id)
		user.LikeMediaCount--
	}

	change[user.Id] = true

	fmt.Printf("[帳號] 名稱:%s 對作品ID:%d 喜歡:%v\n", user.UserName, id, like)
}

func (user *User) EditLikeReview(id uint32, like bool) {
	if like {
		user.LikeReviews[id] = true
		user.LikeReviewCount++
	} else {
		delete(user.LikeReviews, id)
		user.LikeReviewCount--
	}

	change[user.Id] = true

	fmt.Printf("[帳號] 名稱:%s 對評論id:%d 喜歡:%v\n", user.UserName, id, like)
}

func (user *User) EditWatchedMedia(id uint32, watched bool) {
	if watched {
		user.WatchedMedias[id] = true
	} else {
		delete(user.WatchedMedias, id)
	}

	change[user.Id] = true

	fmt.Printf("[帳號] 名稱:%s 對作品id:%d 看過:%v\n", user.UserName, id, watched)
}

func (user *User) IsLikeMedia(id uint32) bool {
	_, ok := user.LikeMedias[id]

	return ok
}

func (user *User) IsLikeReview(id uint32) bool {
	_, ok := user.LikeReviews[id]

	return ok
}

func (user *User) IsWatchedMedia(id uint32) bool {
	_, ok := user.WatchedMedias[id]

	return ok
}
