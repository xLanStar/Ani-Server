package auth

import (
	"Ani-Server/internal/alert"
	"Ani-Server/internal/userManager"
	"fmt"
	"os"

	"github.com/robbert229/jwt"
)

var secret string
var algorithm jwt.Algorithm

var cache map[string]*userManager.User = make(map[string]*userManager.User)

func Init() {
	secret = os.Getenv("SECRET")
	algorithm = jwt.HmacSha256(secret)
}

func GenerateJWT(account, password string) (token string) {
	claims := jwt.NewClaim()
	claims.Set("account", account)
	claims.Set("password", password)

	token, err := algorithm.Encode(claims)
	if err != nil {
		panic(&alert.BadServer)
	}

	return token
}

func ValidateJWT(token string) *userManager.User {
	user, ok := cache[token]
	if ok {
		fmt.Printf("[帳號] 名稱:%s 已驗證\n", user.UserName)
		return user
	}

	err := algorithm.Validate(token)
	if err != nil {
		fmt.Println("algorithm validate failed")
		panic(&alert.IllegalData)
	}

	loadedClaims, err := algorithm.Decode(token)
	if err != nil {
		fmt.Println("algorithm decode failed")
		panic(&alert.IllegalData)
	}

	accountClaim, err := loadedClaims.Get("account")
	if err != nil {
		fmt.Println("token no account")
		panic(&alert.IllegalData)
	}

	account, ok := accountClaim.(string)
	if !ok {
		fmt.Println("account is not a string")
		panic(&alert.IllegalData)
	}

	// if userNameString != userName {
	// 	panic("使用者名稱有誤!")
	// }

	passwordClaim, err := loadedClaims.Get("password")
	if err != nil {
		fmt.Println("token no password")
		panic(&alert.IllegalData)
	}

	password, ok := passwordClaim.(string)
	if !ok {
		fmt.Println("password is not a string")
		panic(&alert.IllegalData)
	}

	cache[token] = userManager.ValidateAccount(account, password)

	fmt.Printf("[帳號] 名稱:%s 驗證成功\n", cache[token].UserName)

	return cache[token]
}
