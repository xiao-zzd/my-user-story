package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)
func TestMain(t *testing.T) {
// 创建要发送的JSON数据

data := map[string]interface{}{
	"url":        "https://stream7.iqilu.com/10339/article/202002/17/4417a27b1a656f4779eaa005ecd1a1a0.mp4",
	"start_time": 0,
	"end_time":   15,
	"user_id":    "bbn实时",
}
// 将数据编码为JSON字节流
jsonData, err := json.Marshal(data)
if err != nil {
	fmt.Println("JSON编码失败:", err)
	return
}
// 创建一个POST请求
req, err := http.NewRequest("POST", "http://127.0.0.1/clip", bytes.NewBuffer(jsonData))
if err != nil {
	fmt.Println("创建请求失败:", err)
	return
}
// 设置请求头
req.Header.Set("Content-Type", "application/json")
// 创建一个HTTP客户端并发送请求
client := &http.Client{}
resp, err := client.Do(req)
if err != nil {
	fmt.Println("发送请求失败:", err)
	return
}
defer resp.Body.Close()
// 打印响应状态码
var responseData map[string]interface{}
err = json.NewDecoder(resp.Body).Decode(&responseData)
if err != nil {
	fmt.Println("Failed to decode response data:", err)
	return
}
fmt.Println(responseData)


}
	
	
	