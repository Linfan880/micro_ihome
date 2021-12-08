package main

import (
	"github.com/gin-gonic/gin"

	"bj38web/web/controller"
	"bj38web/web/model"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-contrib/sessions"
)
func LoginFilter() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 初始化 Session 对象
		s := sessions.Default(ctx)
		userName := s.Get("userName")
		if userName == nil {
			ctx.Abort()			// 从这里返回, 不必继续执行了
		} else {
			ctx.Next()			// 继续向下
		}
	}
}
// 添加gin框架开发3步骤
func main(){
	// 初始化 MySQL 链接池
	model.InitDb()
	// 初始化路由
	router := gin.Default()
	//初始化Redis连接池
	model.InitRedis()
	// 路由匹配
	//router.GET("/", func(context *gin.Context) {
	//	context.Writer.WriteString("项目开始了....")
	//})
	// 初始化容器
	store, _ := redis.NewStore(10, "tcp", "192.168.28.128:6379", "", []byte("bj38"))

	// 使用容器
	router.Use(sessions.Sessions("mysession", store))

	router.Static("/home", "view")

	r1 := router.Group("/api/v1.0")
	{
		r1.GET("/session", controller.GetSession)
		r1.GET("/imagecode/:uuid", controller.GetImageCd)
		r1.GET("/smscode/:phone", controller.GetSmscd)
		r1.POST("/users", controller.PostRet)
		r1.GET("/areas", controller.GetArea)
		r1.POST("sessions", controller.PostLogin)
		r1.Use(LoginFilter())  //以后的路由,都不需要再校验 Session 了. 直接获取数据即可!

		r1.DELETE("/session", controller.DeleteSession)
		r1.GET("/user", controller.GetUserInfo)			//显示个人信息
		r1.PUT("/user/name", controller.PutUserInfo)	//更新用户名
		r1.POST("/user/avatar", controller.PostAvatar)
		r1.POST("/user/auth", controller.PutUserAuth)	//实名认证
		// 获取实名认证
		r1.GET("/user/auth", controller.GetUserInfo)
		//获取已发布房源信息
		r1.GET("/user/houses",controller.GetUserHouses)
		//发布房源
		r1.POST("/houses",controller.PostHouses)
		////添加房源图片
		r1.POST("/houses/:id/images",controller.PostHousesImage)
		//展示房屋详情
		r1.GET("/houses/:id",controller.GetHouseInfo)
		//展示首页轮播图
		r1.GET("/house/index",controller.GetIndex)
		//搜索房屋
		r1.GET("/houses",controller.GetHouses)



		//下订单
		r1.POST("/orders",controller.PostOrders)
		//获取订单
		r1.GET("/user/orders",controller.GetUserOrder)
		//同意/拒绝订单
		r1.PUT("/orders/:id/status",controller.PutOrders)

	}

	// 启动运行
	router.Run(":8080")
}
