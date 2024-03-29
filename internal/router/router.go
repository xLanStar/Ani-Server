package router

import (
	"Ani-Server/internal/alert"
	"Ani-Server/internal/auth"
	"Ani-Server/internal/media"
	"Ani-Server/internal/mediaManager"
	"Ani-Server/internal/reviewManager"
	"Ani-Server/internal/userManager"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/mandrigin/gin-spa/spa"
)

var (
	webFolder     string
	profileFolder string
)

func Init(WebFolder, ProfileFolder string) {
	webFolder = WebFolder
	profileFolder = ProfileFolder
}

func Cors(c *gin.Context) {
	header := c.Writer.Header()
	header.Set("Access-Control-Allow-Origin", "http://localhost:4000")
	header.Set("Access-Control-Allow-Credentials", "true")
	header.Set("Access-Control-Allow-Headers", "Content-Type")

	c.Next()
}

func PostTITLEHandler(c *gin.Context) {
	var data []uint32
	c.Bind(&data)

	var responseData struct {
		Titles       map[uint32]string `json:"titles,omitempty"`
		HasResources []uint32          `json:"hasResources,omitempty"`
	}

	responseData.Titles, responseData.HasResources = mediaManager.GetSimpleMediaInfo(data)

	if responseData.Titles == nil && responseData.HasResources == nil {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, responseData)
}

/*
mediaId 	uint32	: 作品ID
relativeIds []uint32: 相關作品ID
*/
func PostMEDIAHandler(c *gin.Context) {
	var data struct {
		Id          uint32   `json:"id"`
		RelativeIds []uint32 `json:"relativeIds"`
	}
	c.Bind(&data)

	if data.Id == 0 {
		panic(&alert.IllegalData)
	}

	var responseData struct {
		Media   media.IMedia            `json:"media,omitempty"`
		Reviews []*reviewManager.Review `json:"reviews,omitempty"`

		User       *userManager.User     `json:"user,omitempty"`
		UserReivew *reviewManager.Review `json:"userReview,omitempty"`
		LikeReview []uint32              `json:"likeReview,omitempty"`
		LikeMedia  bool                  `json:"likeMedia,omitempty"`
		Watched    bool                  `json:"watched,omitempty"`

		Titles       map[uint32]string `json:"titles,omitempty"`
		HasResources []uint32          `json:"hasResources,omitempty"`
	}

	// 從資料庫中取得資料
	// if !mediaManager.HasMediaId(data.Id) {
	// 	fmt.Println("此作品還尚未存在於資料庫中")

	// 	mediaManager.FetchMediaById(data.Id)
	// }

	responseData.Media = mediaManager.GetMediaById(data.Id)

	responseData.Reviews = reviewManager.GetMediaReviews(data.Id)

	if len(data.RelativeIds) != 0 {
		responseData.Titles, responseData.HasResources = mediaManager.GetSimpleMediaInfo(data.RelativeIds)
	}

	token, err := c.Cookie("token")

	// 登入時，才有 LikeMedia、LikeReview 資料
	if err == nil {
		user := auth.ValidateJWT(token)

		responseData.UserReivew = reviewManager.GetUserReview(user.Id, data.Id)
		responseData.LikeMedia = user.IsLikeMedia(data.Id)
		responseData.LikeReview = make([]uint32, 0, len(responseData.LikeReview)/2)

		for _, review := range responseData.Reviews {
			if user.IsLikeReview(review.Id) {
				responseData.LikeReview = append(responseData.LikeReview, uint32(review.Id))
			}
		}

		responseData.Watched = user.IsWatchedMedia(data.Id)
	}

	c.JSON(http.StatusOK, responseData)
}

/*
id 	 uint32 				: 作品ID
data map[string]interface{} : 編輯項目
*/
func PostEDITMEDIAHandler(c *gin.Context) {
	var data struct {
		Id   uint32                 `json:"id"`
		Type media.MediaType        `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	c.BindJSON(&data)

	fmt.Println("editmedia:", data.Id, data.Data)

	// 從資料庫中取得資料
	if !mediaManager.HasMediaId(data.Id) {
		fmt.Println("此作品還尚未存在於資料庫中")

		mediaManager.FetchMediaById(data.Id)
	}

	mediaManager.EditMedia(data.Id, data.Type, data.Data)

	c.JSON(http.StatusOK, gin.H{"media": mediaManager.GetMediaById(data.Id), "alert": &alert.EditSuccess})
}

func responseSimpleUserData(c *gin.Context, user *userManager.User) {
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.Id,
			"userName": user.UserName,
		},
		"alert": &alert.LoggedInSuccess,
	})
}

// login with token and response user
func PostVALIDATEHandler(c *gin.Context) {
	token, err := c.Cookie("token")

	// 特殊例外: Validate 若接收到無 Token 請求，將忽略此請求
	if err != nil || len(token) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	user := auth.ValidateJWT(token)

	responseSimpleUserData(c, user)
}

/*
account 	string	: 帳號
password 	string	: 密碼
*/
func PostLOGINHandler(c *gin.Context) {
	var data struct {
		Account  string `json:"account"`
		Password string `json:"password"`
	}

	c.Bind(&data)

	user := userManager.ValidateAccount(data.Account, data.Password)

	c.SetCookie("token", auth.GenerateJWT(data.Account, data.Password), 60*60*24*7, "/", "", false, true)

	responseSimpleUserData(c, user)
}

/*
account 	string	: 帳號
password 	string	: 密碼
userName 	string	: 名稱
*/
func PostREGISTERHandler(c *gin.Context) {
	var data struct {
		Account  string `json:"account"`
		Password string `json:"password"`
		UserName string `json:"userName"`
	}
	c.Bind(&data)
	user := userManager.RegistryAccount(data.Account, data.Password, data.UserName)
	c.SetCookie("token", auth.GenerateJWT(data.Account, data.Password), 60*60*24*7, "/", "", false, true)
	responseSimpleUserData(c, user)
}

// logout and clear token
func PostLOGOUTHandler(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		panic(&alert.NotLoggedIn)
	}

	c.SetCookie("token", "", -1, "/", "", false, true)

	user := auth.ValidateJWT(token)

	fmt.Println("Id:", user.Id, " 使用者名稱:", user.UserName, " 登出")

	c.JSON(http.StatusOK, gin.H{
		"alert": &alert.LoggedOutSuccess,
	})
}

/*
uint32 : 帳號ID
*/
func GetUSERHandler(c *gin.Context) {
	sUserId := c.Param("userid")

	userId, err := strconv.Atoi(sUserId)

	if err != nil {
		panic(&alert.DataFormatError)
	}

	user := userManager.GetUser(uint32(userId))

	if user == nil {
		panic(&alert.NotFoundAccount)
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
func PostUSERHandler(c *gin.Context) {
	var data struct {
		Id uint32 `json:"id"`
	}
	c.BindJSON(&data)

	user := userManager.GetUser(uint32(data.Id))

	if user == nil {
		panic(&alert.NotFoundAccount)
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

/*
uint32 : 作品ID
*/
func PostLIKEMEDIAHandler(c *gin.Context) {
	var data struct {
		Id uint32 `json:"id"`
	}
	c.BindJSON(&data)

	token, err := c.Cookie("token")

	if err != nil {
		panic(&alert.NotLoggedIn)
	}

	user := auth.ValidateJWT(token)

	user.EditLikeMedia(data.Id, !user.IsLikeMedia(data.Id))

	c.JSON(http.StatusOK, gin.H{
		"like":  user.IsLikeMedia(data.Id),
		"alert": &alert.EditSuccess,
	})
}

/*
media 	uint32	: 作品ID
review 	uint32	: 評論ID
*/
func PostLIKEREVIEWHandler(c *gin.Context) {
	var data struct {
		MediaId  uint32 `json:"media"`
		ReviewId uint32 `json:"review"`
	}

	c.Bind(&data)

	// 使用者驗證
	token, err := c.Cookie("token")
	if err != nil {
		panic(&alert.NotLoggedIn)
	}
	user := auth.ValidateJWT(token)

	reviewManager.LikeReview(user.Id, data.MediaId, data.ReviewId)

	c.JSON(http.StatusOK, gin.H{
		"like":  user.IsLikeReview(data.ReviewId),
		"alert": &alert.EditSuccess,
	})
}

/*
media 	uint32 	: 作品ID
rank 	uint8	: 評級
content string	: 內容
*/
func PostEDITREVIEWHandler(c *gin.Context) {
	var data struct {
		Media   uint32             `json:"media"`
		Rank    reviewManager.Rank `json:"rank"`
		Content string             `json:"content"`
	}

	c.Bind(&data)

	// 使用者驗證
	token, err := c.Cookie("token")
	if err != nil {
		panic(&alert.NotLoggedIn)
	}
	user := auth.ValidateJWT(token)

	// 修改評論

	if reviewManager.UserHasReview(user.Id, data.Media) {
		reviewManager.EditUserReview(user, data.Media, data.Rank, data.Content)
	} else {
		reviewManager.AddReview(user, data.Media, data.Rank, data.Content)
	}
	c.JSON(http.StatusOK, gin.H{
		"review": *reviewManager.GetUserReview(user.Id, data.Media),
		"alert":  &alert.EditSuccess,
	})
}

/*
uint32 	: 作品ID
*/
func PostDELETEREVIEWHandler(c *gin.Context) {
	var data struct {
		Id uint32 `json:"id"`
	}

	c.Bind(&data)

	// 使用者驗證
	token, err := c.Cookie("token")
	if err != nil {
		panic(&alert.NotLoggedIn)
	}
	user := auth.ValidateJWT(token)

	reviewManager.DeleteUserReview(user, data.Id)

	c.JSON(http.StatusOK, gin.H{
		"alert": &alert.EditSuccess,
	})
}

func PanicHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			if responseAlert, ok := err.(*alert.Alert); ok {
				c.JSON(http.StatusBadRequest, gin.H{"alert": responseAlert})
				fmt.Println(responseAlert)
			} else {
				c.JSON(http.StatusForbidden, gin.H{"alert": alert.BadServer})
				fmt.Println("未知的錯誤:", err)
			}
		}
	}()
	c.Next()
}

func GetPROFILEHandler(c *gin.Context) {
	userid := c.Param("userid")

	c.File(profileFolder + userid)
}

func PostEDITUSERHandler(c *gin.Context) {
	// 使用者驗證
	token, err := c.Cookie("token")
	if err != nil {
		panic(&alert.NotLoggedIn)
	}
	user := auth.ValidateJWT(token)

	// 頭像
	file, err := c.FormFile("profile")

	if err == nil {
		err = c.SaveUploadedFile(file, profileFolder+strconv.Itoa(int(user.Id)))
		if err != nil {
			panic(&alert.BadServer)
		}
	}

	// 自我介紹
	introduction, ok := c.GetPostForm("introduction")
	if ok {
		user.EditIntroduction(introduction)
	}

	c.JSON(http.StatusOK, gin.H{
		"alert": &alert.EditSuccess,
	})
}

func MapRouter(server *gin.Engine, memoryStore *persist.MemoryStore) {
	// static assets folder
	server.Static("/assets", webFolder+"assets/")

	// static assets folder
	server.Static("/favicon.ico", webFolder+"favicon.ico")

	// middlewares
	server.Use(PanicHandler, Cors)

	// routers
	server.POST("/api/title/", PostTITLEHandler)
	server.POST("/api/media/", PostMEDIAHandler)
	server.POST("/api/editmedia/", PostEDITMEDIAHandler)
	server.POST("/api/validate/", PostVALIDATEHandler)
	server.POST("/api/login/", PostLOGINHandler)
	server.POST("/api/logout/", PostLOGOUTHandler)
	server.POST("/api/register/", PostREGISTERHandler)
	server.POST("/api/user/", PostUSERHandler)
	server.GET("/api/user/:userid/", GetUSERHandler)
	server.POST("/api/likemedia/", PostLIKEMEDIAHandler)
	server.POST("/api/likereview/", PostLIKEREVIEWHandler)
	server.POST("/api/editreview/", PostEDITREVIEWHandler)
	server.POST("/api/deletereview/", PostDELETEREVIEWHandler)
	server.GET("/api/profile/:userid/", GetPROFILEHandler)
	server.POST("/api/edituser/", PostEDITUSERHandler)

	// last direct to weburl
	server.Use(spa.Middleware("/", webFolder))

}
