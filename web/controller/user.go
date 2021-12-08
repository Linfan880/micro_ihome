package controller

import (
	"bj38web/web/model"
	"fmt"
	"github.com/afocus/captcha"
	"github.com/gomodule/redigo/redis"

	"github.com/gin-gonic/gin"
	"bj38web/web/utils"
	"github.com/micro/go-micro"
	"github.com/micro/go-plugins/registry/consul"
	"image/png"
	"net/http"
	getCaptcha "bj38web/web/proto/getCaptcha"   // 给包起别名
	houseMicro "bj38web/web/proto/house"
	"context"
	"encoding/json"
	userMicro "bj38web/web/proto/user"
	orderMicro "bj38web/web/proto/userOrder"
	"github.com/gin-contrib/sessions"
	"github.com/tedcy/fdfs_client"
	"path"
)

// 获取 session 信息.
func GetSession(ctx *gin.Context) {
	resp := make(map[string]interface{})
	// 获取 Session 数据
	s := sessions.Default(ctx) // 初始化 Session 对象
	userName := s.Get("userName")

	// 用户没有登录.---没存在 MySQL中, 也没存在 Session 中
	if userName == nil {
		resp["errno"] = utils.RECODE_SESSIONERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_SESSIONERR)
	} else {
		resp["errno"] = utils.RECODE_OK
		resp["errmsg"]  = utils.RecodeText(utils.RECODE_OK)

		//var nameData struct {
		//	Name string `json:"name"`
		//}
		//nameData.Name = userName.(string) // 类型断言
		nameData := make(map[string]interface{})
		nameData["name"] = userName

		resp["data"] = nameData
	}

	ctx.JSON(http.StatusOK, resp)
}
// 获取图片信息
func GetImageCd(ctx *gin.Context) {
	// 获取图片验证码 uuid
	uuid := ctx.Param("uuid")

	// 指定 consul 服务发现
	consulReg := consul.NewRegistry()
	consulService := micro.NewService(
		micro.Registry(consulReg),
	)

	// 初始化客户端
	microClient := getCaptcha.NewGetCaptchaService("getCaptcha", consulService.Client())

	// 调用远程函数
	resp, err := microClient.Call(context.TODO(), &getCaptcha.Request{Uuid:uuid})
	if err != nil {
		fmt.Println("未找到远程服务...")
		return
	}

	// 将得到的数据,反序列化,得到图片数据
	var img captcha.Image
	json.Unmarshal(resp.Img, &img)

	// 将图片写出到 浏览器.
	png.Encode(ctx.Writer, img)

	fmt.Println("================图片验证码的uuid============= ", uuid)
}
func GetSmscd(ctx *gin.Context) {
	// 获取短信验证码
	phone := ctx.Param("phone")
	// 拆分 GET 请求中 的 URL === 格式: 资源路径?k=v&k=v&k=v
	imgCode := ctx.Query("text")	//图片验证码的值
	uuid := ctx.Query("id")			//图片验证码的一个序列号
	fmt.Println("==================out==============:",phone, imgCode, uuid)
	////创建一个容器，存储返回信息
	//resp := make(map[string]string)
	// 指定Consul 服务发现
	consulReg := consul.NewRegistry()
	consulService := micro.NewService(
		micro.Registry(consulReg),
		)
	// 初始化客户端
	microClient := userMicro.NewUserService("go.micro.srv.user", consulService.Client())

	// 调用远程函数:
	resp, err := microClient.SendSms(context.TODO(), &userMicro.Request{
				Phone: phone,
				ImgCode: imgCode,
				Uuid: uuid})
	if err != nil {
		fmt.Println("调用远程函数 SendSms 失败:", err)
		return
	}

	// 发送校验结果 给 浏览器
	ctx.JSON(http.StatusOK, resp)
}
func PostRet(ctx *gin.Context) {
	/*	mobile := ctx.PostForm("mobile")
	pwd := ctx.PostForm("password")
	sms_code := ctx.PostForm("sms_code")
	fmt.Println("m = ", mobile, "pwd = ", pwd, "sms_code = ",sms_code)*/
	// 获取数据
	var regData struct {
		Mobile   string `json:"mobile"`
		PassWord string `json:"password"`
		SmsCode  string `json:"sms_code"`
	}
	ctx.Bind(&regData)
	fmt.Println("==================获取到的数据为================", regData)


	// 初始化consul
	microService := utils.InitMicro()
	microClient := userMicro.NewUserService("go.micro.srv.user", microService.Client())

	// 调用远程函数
	resp, err := microClient.Register(context.TODO(), &userMicro.RegReq{
			Mobile:   regData.Mobile,
			SmsCode:  regData.SmsCode,
			Password: regData.PassWord,
	})
	if err != nil {
		fmt.Println("注册用户, 找不到远程服务!", err)
		return
	}
	// 写给浏览器
	ctx.JSON(http.StatusOK, resp)
}

// 获取地域信息
func GetArea(ctx *gin.Context) {
	// 先从MySQL中获取数据.
	var areas []model.Area

	// 从缓存redis 中, 获取数据
	conn := model.RedisPool.Get()
	// 当初使用 "字节切片" 存入, 现在使用 切片类型接收
	areaData, _ := redis.Bytes(conn.Do("get", "areaData"))
	// 没有从 Redis 中获取到数据
	if len(areaData) == 0 {
		fmt.Println("从 MySQL 中 获取数据...")
		model.GlobalConn.Find(&areas)
		// 把数据写入到 redis 中. , 存储结构体序列化后的 json 串
		areaBuf, _ := json.Marshal(areas)
		conn.Do("set", "areaData", areaBuf)

	} else {
		fmt.Println("从 redis 中 获取数据...")
		// redis 中有数据
		json.Unmarshal(areaData, &areas)
	}

	resp := make(map[string]interface{})
	resp["errno"] = "0"
	resp["errmsg"] = utils.RecodeText(utils.RECODE_OK)
	resp["data"] = areas

	ctx.JSON(http.StatusOK, resp)
}

// 处理登录业务
func PostLogin(ctx *gin.Context) {
	// 获取前端数据
	var loginData struct {
		Mobile   string `json:"mobile"`
		PassWord string `json:"password"`
	}
	ctx.Bind(&loginData)

	resp := make(map[string]interface{})

	//获取 数据库数据, 查询是否和数据的数据匹配
	userName, err := model.Login(loginData.Mobile, loginData.PassWord)
	if err == nil {
		// 登录成功!
		resp["errno"] = utils.RECODE_OK
		resp["errmsg"] = utils.RecodeText(utils.RECODE_OK)

		// 将登录状态, 保存到Session中
		s := sessions.Default(ctx)		// 初始化session
		s.Set("userName", userName)   // 将用户名设置到session中.
		s.Save()
	} else {
		// 登录失败!
		resp["errno"] = utils.RECODE_LOGINERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_LOGINERR)
	}

	ctx.JSON(http.StatusOK, resp)
}

// 退出登录
func DeleteSession(ctx *gin.Context) {

	resp := make(map[string]interface{})

	// 初始化 Session 对象
	s := sessions.Default(ctx)
	// 删除 Session 数据
	s.Delete("userName") // 没有返回值
	// 必须使用 Save 保存
	err := s.Save() // 有返回值

	if err != nil {
		resp["errno"] = utils.RECODE_IOERR // 没有合适错误,使用 IO 错误!
		resp["errmsg"] = utils.RecodeText(utils.RECODE_IOERR)

	} else {
		resp["errno"] = utils.RECODE_OK
		resp["errmsg"] = utils.RecodeText(utils.RECODE_OK)
	}
	ctx.JSON(http.StatusOK, resp)
}

// 获取用户基本信息
func GetUserInfo(ctx *gin.Context) {
	resp := make(map[string]interface{})
	defer ctx.JSON(http.StatusOK, resp)

	// 获取 Session, 得到 当前 用户信息
	s := sessions.Default(ctx)
	userName := s.Get("userName")
	// 判断用户名是否存在.
	if userName == nil { // 用户没登录, 但进入该页面, 恶意进入.
		resp["errno"] = utils.RECODE_SESSIONERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_SESSIONERR)
		return // 如果出错, 报错, 退出
	}

	// 根据用户名, 获取 用户信息  ---- 查 MySQL 数据库  user 表.
	user, err := model.GetUserInfo(userName.(string))
	if err != nil {
		resp["errno"] = utils.RECODE_DBERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_DBERR)
		return // 如果出错, 报错, 退出
	}

	resp["errno"] = utils.RECODE_OK
	resp["errmsg"] = utils.RecodeText(utils.RECODE_OK)
	temp := make(map[string]interface{})
	temp["user_id"] = user.ID
	temp["name"] = user.Name
	temp["mobile"] = user.Mobile
	temp["real_name"] = user.Real_name
	temp["id_card"] = user.Id_card
	temp["avatar_url"] = "http://192.168.28.128:8888/" + user.Avatar_url
	resp["data"] = temp
}

// 更新用户名
func PutUserInfo(ctx *gin.Context) {
	// 获取当前用户名
	s := sessions.Default(ctx) // 初始化Session 对象
	userName := s.Get("userName")

	// 获取新用户名		---- 处理 Request Payload 类型数据. Bind()
	var nameData struct {
		Name string `json:"name"`
	}
	ctx.Bind(&nameData)

	// 更新用户名
	resp := make(map[string]interface{})
	defer ctx.JSON(http.StatusOK, resp)

	// 更新数据库中的 name
	err := model.UpdateUserName(nameData.Name, userName.(string))
	if err != nil {
		resp["errno"] = utils.RECODE_DBERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_DBERR)
		return
	}
	// 更新 Session 数据
	s.Set("userName", nameData.Name)
	err = s.Save() // 必须保存
	if err != nil {
		resp["errno"] = utils.RECODE_SESSIONERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_SESSIONERR)
		return
	}
	resp["errno"] = utils.RECODE_OK
	resp["errmsg"] = utils.RecodeText(utils.RECODE_OK)
	resp["data"] = nameData
}

// 上传头像
func PostAvatar(ctx *gin.Context) {
	/*// 获取图片文件, 静态文件对象
	file, _ := ctx.FormFile("avatar")
	// 上传文件到项目中
	err := ctx.SaveUploadedFile(file, "test/"+file.Filename)
	fmt.Println(err)*/
	file,_ := ctx.FormFile("avatar")
	clt,_ := fdfs_client.NewClientWithConfig("/etc/fdfs/client.conf")
	f, _ := file.Open()

	buf := make([]byte, file.Size)
	f.Read(buf)

	//go语言根据文件名获取文件后缀
	fileExt := path.Ext(file.Filename)	//传文件名字，获取文件后缀-----默认带有“.”
	remotedId,_ := clt.UploadByBuffer(buf, fileExt[1:])
	//获取session，得到对应的用户名
	userName := sessions.Default(ctx).Get("userName")
	//根据用户名，更新用户头像
	model.UpdateAvatar(userName.(string), remotedId)

	fmt.Println("remotedId is :", remotedId) //凭证
	resp := make(map[string]interface{})
	resp["errno"] = utils.RECODE_OK
	resp["errmsg"] = utils.RecodeText(utils.RECODE_OK)
	temp := make(map[string]interface{})
	temp["avatar_url"] = "http://192.168.28.128:8888/" + remotedId
	resp["data"] = temp

	ctx.JSON(http.StatusOK, resp)
}


type AuthStu struct {
	IdCard   string `json:"id_card"`
	RealName string `json:"real_name"`
}

// 上传实名认证
func PutUserAuth(ctx *gin.Context) {
	//获取数据
	var auth AuthStu
	err := ctx.Bind(&auth)
	//校验数据
	if err != nil {
		fmt.Println("获取数据错误", err)
		return
	}

	session := sessions.Default(ctx)
	userName := session.Get("userName")


	resp := make(map[string]interface{})
	err = model.SaveRealName(userName.(string),auth.RealName,auth.IdCard)
	if err != nil{
		resp["errno"] = utils.RECODE_DBERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_DBERR)
	} else{
		resp["errno"] = utils.RECODE_OK
		resp["errmsg"] = utils.RecodeText(utils.RECODE_OK)
	}


	////处理数据 微服务
	//microClient := userMicro.NewUserService("go.micro.srv.user", utils.GetMicroClient())
	//
	////调用远程服务
	//resp, _ := microClient.AuthUpdate(context.TODO(), &userMicro.AuthReq{
	//	UserName: userName.(string),
	//	RealName: auth.RealName,
	//	IdCard:   auth.IdCard,
	//})

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

//获取已发布房源信息  假数据
func GetUserHouses(ctx *gin.Context) {
	//获取用户名
	userName := sessions.Default(ctx).Get("userName")

	microClient := houseMicro.NewHouseService("go.micro.srv.house", utils.InitMicro().Client())
	//调用远程服务
	resp, _ := microClient.GetHouseInfo(context.TODO(), &houseMicro.GetReq{UserName: userName.(string)})

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

type HouseStu struct {
	Acreage   string   `json:"acreage"`
	Address   string   `json:"address"`
	AreaId    string   `json:"area_id"`
	Beds      string   `json:"beds"`
	Capacity  string   `json:"capacity"`
	Deposit   string   `json:"deposit"`
	Facility  []string `json:"facility"`
	MaxDays   string   `json:"max_days"`
	MinDays   string   `json:"min_days"`
	Price     string   `json:"price"`
	RoomCount string   `json:"room_count"`
	Title     string   `json:"title"`
	Unit      string   `json:"unit"`
}

//发布房源
func PostHouses(ctx *gin.Context) {
	//获取数据		bind数据的时候不带自动转换
	var house HouseStu
	err := ctx.Bind(&house)

	//校验数据
	if err != nil {
		fmt.Println("获取数据错误", err)
		return
	}

	//获取用户名
	userName := sessions.Default(ctx).Get("userName")

	//处理数据  服务端处理
	microClient := houseMicro.NewHouseService("go.micro.srv.house", utils.InitMicro().Client())

	//调用远程服务
	resp, _ := microClient.PubHouse(context.TODO(), &houseMicro.Request{
		Acreage:   house.Acreage,
		Address:   house.Address,
		AreaId:    house.AreaId,
		Beds:      house.Beds,
		Capacity:  house.Capacity,
		Deposit:   house.Deposit,
		Facility:  house.Facility,
		MaxDays:   house.MaxDays,
		MinDays:   house.MinDays,
		Price:     house.Price,
		RoomCount: house.RoomCount,
		Title:     house.Title,
		Unit:      house.Unit,
		UserName:  userName.(string),
	})

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}
//上传房屋图片
func PostHousesImage(ctx *gin.Context) {
	//获取数据
	houseId := ctx.Param("id")
	fmt.Println("=================houseId=================", houseId)
	fileHeader, err := ctx.FormFile("house_image")

	//校验数据
	if houseId == "" || err != nil {
		fmt.Println("传入数据不完整", err)
		return
	}

	//三种校验 大小,类型,防止重名  fastdfs
	if fileHeader.Size > 50000000 {
		fmt.Println("文件过大,请重新选择")
		return
	}

	fileExt := path.Ext(fileHeader.Filename)
	if fileExt != ".png" && fileExt != ".jpg" {
		fmt.Println("文件类型错误,请重新选择")
		return
	}

	//获取文件字节切片
	file, _ := fileHeader.Open()
	buf := make([]byte, fileHeader.Size)
	file.Read(buf)

	//处理数据  服务中实现
	microClient := houseMicro.NewHouseService("go.micro.srv.house", utils.InitMicro().Client())

	//调用远程服务
	resp, _ := microClient.UploadHouseImg(context.TODO(), &houseMicro.ImgReq{
		HouseId: houseId,
		ImgData: buf,
		FileExt: fileExt,
	})

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

//获取房屋详情
func GetHouseInfo(ctx *gin.Context) {
	//获取数据
	houseId := ctx.Param("id")
	//校验数据
	if houseId == "" {
		fmt.Println("获取数据错误")
		return
	}
	userName := sessions.Default(ctx).Get("userName")
	//处理数据
	microClient := houseMicro.NewHouseService("go.micro.srv.house", utils.InitMicro().Client())
	//调用远程服务
	resp, _ := microClient.GetHouseDetail(context.TODO(), &houseMicro.DetailReq{
		HouseId:  houseId,
		UserName: userName.(string),
	})

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}
func GetIndex(ctx *gin.Context) {
	//处理数据
	microClient := houseMicro.NewHouseService("go.micro.srv.house", utils.InitMicro().Client())

	//调用远程服务
	resp, _ := microClient.GetIndexHouse(context.TODO(), &houseMicro.IndexReq{})

	ctx.JSON(http.StatusOK, resp)
}
//搜索房屋
func GetHouses(ctx *gin.Context) {
	//获取数据
	//areaId
	aid := ctx.Query("aid")
	//start day
	sd := ctx.Query("sd")
	//end day
	ed := ctx.Query("ed")
	//排序方式
	sk := ctx.Query("sk")
	//page  第几页
	//ctx.Query("p")
	//校验数据
	if aid == "" || sd == "" || ed == "" || sk == "" {
		fmt.Println("传入数据不完整")
		return
	}

	//处理数据   服务端  把字符串转换为时间格式,使用函数time.Parse()  第一个参数是转换模板,需要转换的二字符串,两者格式一致
	/*sdTime ,_:=time.Parse("2006-01-02 15:04:05",sd+" 00:00:00")
	edTime,_ := time.Parse("2006-01-02",ed)*/

	/*sdTime,_ :=time.Parse("2006-01-02",sd)
	edTime,_ := time.Parse("2006-01-02",ed)
	d := edTime.Sub(sdTime)
	fmt.Println(d.Hours())*/

	microClient := houseMicro.NewHouseService("go.micro.srv.house", utils.InitMicro().Client())
	//调用远程服务
	resp, _ := microClient.SearchHouse(context.TODO(), &houseMicro.SearchReq{
			Aid: aid,
			Sd:  sd,
			Ed:  ed,
			Sk:  sk,
	})

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}


////////////////////订单相关/////////////////
type OrderStu struct {
	EndDate   string `json:"end_date"`
	HouseId   string `json:"house_id"`
	StartDate string `json:"start_date"`
}
//下订单
func PostOrders(ctx *gin.Context) {
	//获取数据
	var order OrderStu
	err := ctx.Bind(&order)

	//校验数据
	if err != nil {
		fmt.Println("获取数据错误", err)
		return
	}
	//获取用户名
	userName := sessions.Default(ctx).Get("userName")

	//处理数据  服务端处理业务
	microClient := orderMicro.NewUserOrderService("go.micro.srv.userOrder", utils.InitMicro().Client())
	//调用服务
	resp, _ := microClient.CreateOrder(context.TODO(), &orderMicro.Request{
		StartDate: order.StartDate,
		EndDate:   order.EndDate,
		HouseId:   order.HouseId,
		UserName:  userName.(string),
	})

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

//获取订单信息
func GetUserOrder(ctx *gin.Context) {
	//获取get请求传参
	role := ctx.Query("role")
	//校验数据
	if role == "" {
		fmt.Println("获取数据失败")
		return
	}



	//处理数据  服务端
	microClient := orderMicro.NewUserOrderService("go.micro.srv.userOrder", utils.InitMicro().Client())
	//调用远程服务
	resp,_ :=microClient.GetOrderInfo(context.TODO(),&orderMicro.GetReq{
		Role:role,
		UserName:sessions.Default(ctx).Get("userName").(string),
	})

	//返回数据
	ctx.JSON(http.StatusOK,resp)
}


type StatusStu struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}
//更新订单状态
func PutOrders(ctx*gin.Context){
	//获取数据
	id := ctx.Param("id")
	var statusStu StatusStu
	err := ctx.Bind(&statusStu)

	//校验数据
	if err != nil || id == "" {
		fmt.Println("获取数据错误",err)
		return
	}

	//处理数据   更新订单状态
	microClient := orderMicro.NewUserOrderService("go.micro.srv.userOrder", utils.InitMicro().Client())
	//调用元和产能服务
	resp,_ := microClient.UpdateStatus(context.TODO(),&orderMicro.UpdateReq{
		Action:statusStu.Action,
		Reason:statusStu.Reason,
		Id:id,
	})

	//返回数据
	ctx.JSON(http.StatusOK,resp)
}