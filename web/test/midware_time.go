package main

import (
	"fmt"
	"time"
	"github.com/gin-gonic/gin"
)

func Test1(ctx *gin.Context){
	fmt.Println("1111")
	t:=time.Now()
	ctx.Next()
	fmt.Println(time.Now().Sub(t))
}

func Test2() gin.HandlerFunc{
	return func(ctx *gin.Context){
		fmt.Println("3333")
		ctx.Next()
		fmt.Println("5555")
	}
}


func main(){
	router := gin.Default()

	//使用中间件
	router.Use(Test1)
	router.Use(Test2())
	router.GET("/test", func(ctx *gin.Context){
		fmt.Println("2222")
		ctx.Writer.WriteString("hello")
	})
	router.Run(":9999")
}
