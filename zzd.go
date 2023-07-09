package main

import (
	"database/sql"
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"os/exec"
	"strconv"
	"os"
)

//剪辑请求
//1.url剪辑视频地址 2.StartTime起始剪辑时间 3.EndTime结束剪辑时间 4.用户id（用于鉴别用户身份）
type ClipRequest struct {
	URL        string `json:"url"`
	StartTime  int `json:"start_time"`
	EndTime    int `json:"end_time"`
	UserID     string `json:"user_id"`
	
}

//剪辑后返回数据，1.code状态码 2.UserID用户id 3.剪辑完成后视频下载地址 4.Msg消息
type ClipReposonse struct {
	Code       int `json:"code"`
	UserID     string `json:"user_id"`
	URL        string `json:"url"`
	Msg			string `json:"msg"`
	
}
//剪辑查询数据，1.UserID为用户id 2.Schedule为该用户的上传视频下载了多少字节的值
type ProgressQuery struct {
	UserID     string `json:"user_id"`
	Schedule   int64 `json:"schedule"`
}

var db *sql.DB

func main() {

	// 初始化数据库
	initDB()
	
	// 初始化Gin
	r := gin.Default()

	//剪辑接口
	//使用postman进行测试，接口测试地址为http://127.0.0.1/clip/，方式为POST，如下为请求JSON测试模板
	//{
	//     "url":"https://stream7.iqilu.com/10339/article/202002/17/4417a27b1a656f4779eaa005ecd1a1a0.mp4",
	//     "start_time":0,
	//     "end_time":15,
	//     "user_id":"4563"
	// }
	r.POST("/clip", clipHandler)

	//查询进度接口，如http://127.0.0.1/query/zzds即可查询用户zzds所上传视频的进度
	r.GET("/query/:UserID", getClipProgress)

	//设置静态路径
	r.StaticFS("/video", http.Dir("video"))
	
	//设置下载视频路径
	r.GET("/download/:filename", func(c *gin.Context) {
        filename := c.Param("filename")
		c.Header("Content-Type", "application/octet-stream")   // 表示是文件流，唤起浏览器下载，一般设置了这个，就要设置文件名
		c.Header("Content-Transfer-Encoding", "binary")      
        c.File("./video/" + filename)
    })

	if err := r.Run(":80"); err != nil {
		log.Fatal(err)
	}
}

func initDB() {
	var err error
    //创建与mysql的连接
    //username:password@tcp(127.0.0.1:3306)/database_name中的username、password和database_name替换为你的MySQL数据库的实际连接信息。
	db, err = sql.Open("mysql", "root:zhanzhaodong@tcp(127.0.0.1:3306)/zzd")
	if err != nil {
		log.Fatal(err)
	}
	// 创建表 ，如果表已经存在，则IF NOT EXISTS语句将防止重新创建该表。
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS clip_requests (
		id INTEGER PRIMARY KEY AUTO_INCREMENT,
		url TEXT,
		start_time INT,
		end_time INT,
		user_id TEXT
	);
	`
    //执行createTableSQL的创建表sql语法
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}


func clipHandler(c *gin.Context) {
	var request ClipRequest
	
	//将json数据和ClipRequest结构体绑定
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 持久化用户提交记录
	//storeClipRequest用于将用户提交的记录保存到数据库中
	err = storeClipRequest(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 剪辑视频
	//视频测试地址"https://stream7.iqilu.com/10339/article/202002/17/4417a27b1a656f4779eaa005ecd1a1a0.mp4"

	// 设置输出视频文件路径
	outputFile := "./video/new"+request.UserID+".mp4"

	//clipVideo为剪辑视频函数
	videoerr := clipVideo(request.URL,outputFile, request.StartTime, request.EndTime)
	if videoerr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	

	//响应返回数据
	var reposonse ClipReposonse
	reposonse.Code=200
	reposonse.UserID=request.UserID
	reposonse.URL="127.0.0.1/download/new" + request.UserID+".mp4"
	reposonse.Msg="成功"
	c.JSON(200,reposonse)

	

}

//将用户提交的记录保存到数据库中
func storeClipRequest(request ClipRequest) error {
	// 插入数据库
	insertSQL := `
	INSERT INTO clip_requests (url, start_time, end_time, user_id) 
	VALUES (?, ?, ?, ?)
	`
	_, err := db.Exec(insertSQL, request.URL, request.StartTime, request.EndTime, request.UserID)
	if err != nil {
		return err
	}

	return nil
}


//剪辑视频函数，传递正确的输入和输出文件路径，以及剪辑的起始时间和持续时间，可以执行视频剪辑操作。可以在终端查看进度
func clipVideo(inputFile string, outputFile string, startTime int, duration int) error {
	cmdArgs := []string{
		"-i", inputFile,
		"-ss", strconv.Itoa(startTime),
		"-t", strconv.Itoa(duration),
		"-c:v", "copy",
		"-c:a", "copy",
		"-y",
		outputFile,
	}

	cmd := exec.Command("ffmpeg", cmdArgs...)
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

//查询进度接口
func getClipProgress(c *gin.Context){

	//获取用户ID
	ID := c.Param("UserID")

	//根据ID生成对应路劲
	outputFile := "./video/new"+ID+".mp4"

	//使用os.Stat去获取已下载的文件字节大小
	fi, _ := os.Stat(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	totalSize := fi.Size()

	//返回数据格式如下{"user_id":"zzds","schedule":6143952}，user_id为用户ID，Schedule为该用户的上传视频下载了多少字节的值
	var progressquery ProgressQuery
	progressquery.UserID=ID
	progressquery.Schedule=totalSize

	c.JSON(http.StatusOK, progressquery)

}
