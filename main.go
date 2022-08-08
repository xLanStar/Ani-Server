package main

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"time"

	fastio "github.com/xLanStar/go-fast-io"

	"Ani-Server/internal/auth"
	"Ani-Server/internal/config"
	"Ani-Server/internal/mediaManager"
	"Ani-Server/internal/reviewManager"
	"Ani-Server/internal/router"
	"Ani-Server/internal/userManager"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-colorable"
	"github.com/nanmu42/gzip"
)

func NowTime() uint32 {
	now := time.Now()
	return uint32(now.Year()*1000000 + int(now.Month())*10000 + now.Day()*100 + now.Hour())
}

type ServerData struct {
	UserCount       uint32
	ReviewCount     uint32
	LastUpdateif101 uint32
}

func (serverData *ServerData) Load(filePath string) {

	var fileReader fastio.FileReader
	fileReader.Init()
	fileReader.OpenFile(filePath, os.O_RDONLY, 0666)
	serverData.UserCount = fileReader.ReadUint32()
	serverData.ReviewCount = fileReader.ReadUint32()
	serverData.LastUpdateif101 = fileReader.ReadUint32()
	fileReader.Close()
}

func (serverData *ServerData) Save(filePath string) {
	var fileWriter fastio.FileWriter
	fileWriter.Init()
	fileWriter.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	fmt.Printf("保存伺服器資料 帳號數量:%d 評論數量:%d 上次更新if101時間:%d\n", serverData.UserCount, serverData.ReviewCount, serverData.LastUpdateif101)
	fileWriter.WriteUint32(serverData.UserCount)
	fileWriter.WriteUint32(serverData.ReviewCount)
	fileWriter.WriteUint32(serverData.LastUpdateif101)
	fileWriter.Close()
}

func run() {
	// MIME
	mime.AddExtensionType(".js", "application/javascript")

	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	serverData := new(ServerData)
	serverData.Load(cfg.DataFile)
	defer serverData.Save(cfg.DataFile)

	updateIf101Timer := time.NewTimer(time.Hour)
	go func() {
		for {
			<-updateIf101Timer.C
			mediaManager.UpdateIf101(serverData.LastUpdateif101)
			serverData.LastUpdateif101 = NowTime()
		}
	}()

	auth.Init()

	router.Init(cfg.WebFolder, cfg.ProfileFolder)

	mediaManager.Load(cfg.MediaFolder)
	if NowTime() > serverData.LastUpdateif101 {
		mediaManager.UpdateIf101(serverData.LastUpdateif101)
		serverData.LastUpdateif101 = NowTime()
	}
	defer mediaManager.Save()

	userManager.Load(cfg.UserFolder, serverData.UserCount)
	defer func() {
		userManager.Save()
		serverData.UserCount = userManager.GetUserCount()
	}()

	reviewManager.Load(cfg.ReviewFolder, serverData.ReviewCount)
	defer func() {
		reviewManager.Save()
		serverData.ReviewCount = reviewManager.GetReviewCount()
	}()

	// debug
	// user := userManager.RegistryAccount("as95624268", "aa95624268", "Lanstar")
	// user.LikeMedia(31706)
	// user.LikeMedia(124194)

	// user.Introduction = "123\n654\nASD"

	// reviewManager.AddReview(user, 20954, 0, "神作")

	// user := userManager.GetUser(1)
	// reviewManager.AddReview(user, 140960, 0, "神作")

	// process color
	gin.DefaultWriter = colorable.NewColorableStdout()
	gin.ForceConsoleColor()

	// release mode
	gin.SetMode(gin.ReleaseMode)

	// create server
	server := gin.Default()

	// Use Gzip
	server.Use(gzip.DefaultHandler().Gin)

	// Cached in local memory
	memoryStore := persist.NewMemoryStore(time.Hour)

	// router
	router.MapRouter(server, memoryStore)

	// favicon
	// server.Use(favicon.New("./favicon.ico"))

	// graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: server,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// var input uint8
	// for {
	// 	fmt.Scanln(&input)
	// 	if input == 0 {
	// 		break
	// 	} else if input == 1 {
	// 		mediaManager.UpdateIf101(serverData.LastUpdateif101)
	// 		serverData.LastUpdateif101 = NowTime()
	// 	} else if input == 2 {
	// 		mediaManager.Save()
	// 	}
	// }
	var quit chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("關閉伺服器中...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("伺服器崩潰:", err)
		mediaManager.Save()
	}
	log.Println("伺服器關閉")
}

func main() {
	run()

	var quit chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
