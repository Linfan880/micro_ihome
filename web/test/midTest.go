package main

import "github.com/gin-gonic/gin"
import "fmt"

//创建中间件
func Test1(ctx *gin.Context){
	fmt.Println("11111")
	ctx.Next()
	fmt.Println("4444")
}
//创建一种另外格式的中间件
func Test2() gin.HandlerFunc{
	return func(context *gin.Context){
			fmt.Println("3333")
			context.Next()
			fmt.Println("5555")
	}
}

func main(){
	router :=gin.Default()
	//使用中间件
	router.Use(Test1)
	router.Use(Test2())
	router.GET("/test",func(context *gin.Context){
			fmt.Println("2222")
			context.Writer.WriteString("hello world!")
	})
	router.Run(":9989")
}
