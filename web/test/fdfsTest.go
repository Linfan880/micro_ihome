package main

import "github.com/tedcy/fdfs_client"
import "fmt"

func main(){
	clt, err := fdfs_client.NewClientWithConfig("/etc/fdfs/client.conf")
	if err !=nil{
		fmt.Println("初始化客户端错误！err: ",err)
		return
	}

	//上传文件到storage
	resp, err := clt.UploadByFilename("修勾.jpg")
	fmt.Println(resp, err)
}
