package userManager

import (
	"Ani-Server/internal/alert"
	"fmt"
	"os"
	"strconv"

	fastio "github.com/xLanStar/go-fast-io"
)

var (
	userCount  uint32
	users      []User           = make([]User, 0, 20)
	idMap      map[uint32]*User = make(map[uint32]*User)
	AccountMap map[string]*User = make(map[string]*User)

	change map[uint32]bool = make(map[uint32]bool)

	userFolder string
)

func Load(UserFolder string, UserCount uint32) {
	userFolder = UserFolder
	userCount = UserCount

	userFiles, err := os.ReadDir(UserFolder)
	if err != nil {
		return
	}

	var fileReader fastio.FileReader

	fileReader.Init()

	var userptr *User
	for _, userFile := range userFiles {
		err := fileReader.OpenFile(userFolder+userFile.Name(), os.O_RDONLY, 0666)

		if err != nil {
			fmt.Printf("[UserManager] [帳號] 名稱:%s 讀取失敗!\n", userFile.Name())
			continue
		}

		userId, err := strconv.Atoi(userFile.Name())

		if err != nil {
			continue
		}

		users = append(users, User{Id: uint32(userId), LikeMedias: map[uint32]bool{}, LikeReviews: map[uint32]bool{}, Introduction: ""})
		userptr = &users[len(users)-1]
		idMap[uint32(userId)] = userptr

		userptr.Account = fileReader.ReadString()
		AccountMap[userptr.Account] = userptr

		userptr.Password = fileReader.ReadString()

		userptr.UserName = fileReader.ReadString()

		userptr.Introduction = fileReader.ReadString()

		for _, mediaId := range fileReader.ReadUint32Array() {
			userptr.LikeMedias[mediaId] = true
		}

		userptr.LikeMediaCount = uint32(len(userptr.LikeMedias))

		for _, reviewId := range fileReader.ReadUint32Array() {
			userptr.LikeReviews[reviewId] = true
		}

		userptr.LikeReviewCount = uint32(len(userptr.LikeReviews))

		for _, mediaId := range fileReader.ReadUint32Array() {
			userptr.WatchedMedias[mediaId] = true
		}

		fileReader.Close()

		fmt.Printf("[UserManager] 讀取帳號%s\n", userptr)
	}
}

func Save() {
	fmt.Println("[UserManager] 存檔")

	var fileWriter fastio.FileWriter

	fileWriter.Init()

	for _, user := range users {
		if !change[user.Id] {
			continue
		}

		fileWriter.OpenFile(userFolder+strconv.Itoa(int(user.Id)), os.O_WRONLY|os.O_CREATE, 0666)

		fmt.Printf("[UserManager] 保存帳號%s\n", user)

		fileWriter.WriteString(user.Account)
		fileWriter.WriteString(user.Password)
		fileWriter.WriteString(user.UserName)
		fileWriter.WriteString(user.Introduction)
		fileWriter.WriteUint16(uint16(len(user.LikeMedias)))
		for id := range user.LikeMedias {
			fileWriter.WriteUint32(id)
		}
		fileWriter.WriteUint16(uint16(len(user.LikeReviews)))
		for id := range user.LikeReviews {
			fileWriter.WriteUint32(id)
		}
		fileWriter.WriteUint16(uint16(len(user.WatchedMedias)))
		for id := range user.WatchedMedias {
			fileWriter.WriteUint32(id)
		}
		fileWriter.Flush()
		fileWriter.Close()
	}

}

func GetUserCount() uint32 {
	return userCount
}

func ValidateAccount(Account, Password string) *User {
	// 檢查使用者帳號是否存在
	user, ok := AccountMap[Account]
	if !ok {
		panic(&alert.NotFoundAccount)
	}

	// 檢查密碼是否正確
	if Password != user.Password {
		panic(&alert.WrongPassword)
	}

	return user
}

func GetUser(Id uint32) *User {
	return idMap[Id]
}

func GetUsers() *[]User {
	return &users
}

func GetUserIsLikeMedia(Id, MediaId uint32) bool {
	_, ok := idMap[Id].LikeMedias[MediaId]
	return ok
}

func RegistryAccount(Account, Password, UserName string) *User {
	// 檢查帳號是否已經存在
	_, ok := AccountMap[Account]
	if ok {
		panic(&alert.AccountAlreadyExists)
	}

	userCount++
	users = append(users, User{Id: userCount, Account: Account, Password: Password, UserName: UserName, LikeMedias: map[uint32]bool{}, LikeReviews: map[uint32]bool{}, Introduction: ""})
	idMap[userCount] = &users[len(users)-1]
	AccountMap[Account] = idMap[userCount]
	change[userCount] = true

	fmt.Printf("註冊成功 id:%d 使用者名稱:%s 密碼:%s\n", userCount, UserName, Password)

	return idMap[userCount]
}
