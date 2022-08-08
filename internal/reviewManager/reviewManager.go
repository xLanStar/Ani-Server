package reviewManager

import (
	"Ani-Server/internal/alert"
	"Ani-Server/internal/userManager"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	fastio "github.com/xLanStar/go-fast-io"
)

var (
	reviewCount uint32                        = 0
	reviews     []Review                      = make([]Review, 0, 20)
	idMap       map[uint32]*Review            = make(map[uint32]*Review)            // reviewId 對應 *Review
	userMap     map[uint32]map[uint32]*Review = make(map[uint32]map[uint32]*Review) // userId 對應 map[mediaId]*Review
	mediaMap    map[uint32][]*Review          = make(map[uint32][]*Review)          // mediaId 對應 []*Review

	editedReviews map[uint32][]uint32 = make(map[uint32][]uint32) // userId 對應 mediaId

	reviewFolder string
)

func Load(ReviewFolder string, ReviewCount uint32) {
	reviewFolder = ReviewFolder
	reviewCount = ReviewCount

	userReviewFolders, err := ioutil.ReadDir(reviewFolder)
	if err != nil {
		return
	}

	var fileReader fastio.FileReader

	fileReader.Init()

	var reviewptr *Review
	for _, userReviewFolder := range userReviewFolders {
		if !userReviewFolder.IsDir() {
			continue
		}

		userId, err := strconv.Atoi(userReviewFolder.Name())
		if err != nil {
			fmt.Printf("[ReviewManager] 使用者ID:%s 格式不正確!", userReviewFolder.Name())
			continue
		}

		if _, ok := userMap[uint32(userId)]; !ok {
			userMap[uint32(userId)] = make(map[uint32]*Review)
		}

		userReviewFiles, err := ioutil.ReadDir(reviewFolder + userReviewFolder.Name() + "/")
		if err != nil {
			return
		}

		for _, userReviewFile := range userReviewFiles {
			mediaId, err := strconv.Atoi(userReviewFile.Name())
			if err != nil {
				continue
			}

			err = fileReader.OpenFile(reviewFolder+userReviewFolder.Name()+"/"+userReviewFile.Name(), os.O_RDONLY, 0666)
			if err != nil {
				fmt.Printf("[ReviewManager] 帳號:%s 作品ID:%s 開啟失敗\n", userReviewFolder.Name(), userReviewFile.Name())
				continue
			}

			reviews = append(reviews, Review{AuthorId: uint32(userId), Author: &userManager.GetUser(uint32(userId)).UserName})
			reviewptr = &reviews[len(reviews)-1]
			userMap[uint32(userId)][uint32(mediaId)] = reviewptr

			if _, ok := mediaMap[uint32(mediaId)]; !ok {
				mediaMap[uint32(mediaId)] = make([]*Review, 0, 1)
			}
			mediaMap[uint32(mediaId)] = append(mediaMap[uint32(mediaId)], reviewptr)

			reviewptr.Id = fileReader.ReadUint32()
			idMap[reviewptr.Id] = reviewptr

			reviewptr.Rank = Rank(fileReader.Read())

			reviewptr.Content = fileReader.ReadString()

			fileReader.Close()

			fmt.Printf("[ReviewManager] 讀取評論 作品ID:%d%s\n", mediaId, reviewptr)
		}
	}
	// for _, user := range *userManager.GetUsers() {
	// 	for reviewId := range user.LikeReviews {
	// 		idMap[reviewId].Like++
	// 	}
	// }
}

func Save() {
	fmt.Println("[ReviewManager] 存檔")

	if len(editedReviews) == 0 {
		return
	}

	var fileWriter fastio.FileWriter

	fileWriter.Init()

	var reviewptr *Review

	for userId, mediaIds := range editedReviews {
		fmt.Println("[reviewManager] 保存評論 使用者ID:", userId)

		userReviewFolder := reviewFolder + strconv.Itoa(int(userId)) + "/"

		if _, err := os.Stat(userReviewFolder); os.IsNotExist(err) {
			fmt.Println("[reviewManager] 建立使用者資料夾")
			os.Mkdir(userReviewFolder, 0755)
		}

		for _, mediaId := range mediaIds {
			reviewptr = userMap[userId][mediaId]

			if reviewptr == nil {
				continue
			}

			fileWriter.OpenFile(userReviewFolder+strconv.Itoa(int(mediaId)), os.O_WRONLY|os.O_CREATE, 0666)
			fileWriter.WriteUint32(reviewptr.Id)
			fileWriter.WriteUint8(uint8(reviewptr.Rank))
			fileWriter.WriteString(reviewptr.Content)
			fileWriter.Flush()
			fileWriter.Close()

			fmt.Printf("[ReviewManager] 保存評論%s\n", reviewptr)
		}
	}

}

func GetReviewCount() uint32 {
	return reviewCount
}

func AddReview(user *userManager.User, mediaId uint32, rank Rank, content string) {

	if _, ok := userMap[user.Id]; !ok {
		userMap[user.Id] = make(map[uint32]*Review)
	} else if _, ok := userMap[user.Id][mediaId]; ok {
		panic(&alert.IllegalBehavior)
	}

	reviewCount++

	reviews = append(reviews, Review{Id: reviewCount, Rank: rank, Content: content, AuthorId: user.Id, Author: &user.UserName})
	idMap[reviewCount] = &reviews[len(reviews)-1]

	userMap[user.Id][mediaId] = idMap[reviewCount]

	if _, ok := mediaMap[mediaId]; !ok {
		mediaMap[mediaId] = make([]*Review, 0, 1)
	}
	mediaMap[mediaId] = append(mediaMap[mediaId], idMap[reviewCount])

	MarkReviewEdited(user.Id, mediaId)

	fmt.Printf("[AddReview] 新增評論\n%s\n", idMap[reviewCount])
}

func EditUserReview(user *userManager.User, MediaId uint32, rank Rank, content string) {
	fmt.Println(MediaId)

	review, ok := userMap[user.Id][MediaId]

	if !ok {
		panic(&alert.IllegalBehavior)
	}

	review.Rank = rank
	review.Content = content

	MarkReviewEdited(user.Id, MediaId)
}

func DeleteUserReview(user *userManager.User, MediaId uint32) {
	review, ok := userMap[user.Id][MediaId]

	if !ok {
		panic(&alert.IllegalBehavior)
	}

	delete(idMap, review.Id)

	delete(userMap[user.Id], MediaId)

	for i := 0; i != len(mediaMap[MediaId]); i++ {
		if mediaMap[MediaId][i].Id == review.Id {
			mediaMap[MediaId] = append(mediaMap[MediaId][:i], mediaMap[MediaId][i+1:]...)
			break
		}
	}

	filename := reviewFolder + strconv.Itoa(int(user.Id)) + "/" + strconv.Itoa(int(MediaId))

	if _, err := os.Stat(filename); err == nil {
		err := os.Remove(filename)

		if err != nil {
			panic(&alert.BadServer)
		}
	}

}

func MarkReviewEdited(userId, mediaId uint32) {
	if _, ok := editedReviews[userId]; !ok {
		editedReviews[userId] = make([]uint32, 0, 1)
	}

	editedReviews[userId] = append(editedReviews[userId], mediaId)
	fmt.Println("[ReviewManager] 標註編輯: ", userId, mediaId)
}

func GetMediaReviews(MediaId uint32) []*Review {
	return mediaMap[MediaId]
}

func UserHasReview(UserId, MediaId uint32) bool {
	_, ok := userMap[UserId][MediaId]
	return ok
}

func GetUserReviewId(UserId, MediaId uint32) uint32 {
	if review, ok := userMap[UserId][MediaId]; ok {
		return review.Id
	}
	return 0
}

func GetUserReview(UserId, MediaId uint32) *Review {
	return userMap[UserId][MediaId]
}

func LikeReview(UserId, MediaId, ReviewId uint32) {
	if review, ok := userMap[UserId][MediaId]; ok && review.Id == ReviewId {
		panic(&alert.IllegalBehavior)
	}

	review, ok := idMap[ReviewId]

	if !ok {
		panic(&alert.IllegalBehavior)
	}

	user := userManager.GetUser(UserId)

	if user.IsLikeReview(ReviewId) {
		user.EditLikeReview(ReviewId, false)
		review.Like--
	} else {
		user.EditLikeReview(ReviewId, true)
		review.Like++
	}

}
