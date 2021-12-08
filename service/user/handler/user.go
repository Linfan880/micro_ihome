package handler

import (
	"context"
	user "bj38web/service/user/proto/user"
	"encoding/json"
	"bj38web/service/user/model"
	"math/rand"
	"time"

	//"math/rand"
	//"time"
	"fmt"
	"bj38web/service/user/utils"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111" // 引入sms
)

type User struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *User) SendSms(ctx context.Context, req *user.Request, rsp *user.Response) error {

	// 校验图片验证码 是否正确
	result := model.CheckImgCode(req.Uuid, req.ImgCode)
	if result {
		// 发送短信
		credential := common.NewCredential(
			// os.Getenv("TENCENTCLOUD_SECRET_ID"),
			// os.Getenv("TENCENTCLOUD_SECRET_KEY"),
			"",
			"",
		)
		/* 非必要步骤:
		 * 实例化一个客户端配置对象，可以指定超时时间等配置 */
		cpf := profile.NewClientProfile()

		/* SDK默认使用POST方法。
		 * 如果你一定要使用GET方法，可以在这里设置。GET方法无法处理一些较大的请求 */
		cpf.HttpProfile.ReqMethod = "POST"

		/* SDK有默认的超时时间，非必要请不要进行调整
		 * 如有需要请在代码中查阅以获取最新的默认值 */
		// cpf.HttpProfile.ReqTimeout = 5

		/* SDK会自动指定域名。通常是不需要特地指定域名的，但是如果你访问的是金融区的服务
		 * 则必须手动指定域名，例如sms的上海金融区域名： sms.ap-shanghai-fsi.tencentcloudapi.com */
		cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"

		/* SDK默认用TC3-HMAC-SHA256进行签名，非必要请不要修改这个字段 */
		cpf.SignMethod = "HmacSHA1"

		/* 实例化要请求产品(以sms为例)的client对象
		 * 第二个参数是地域信息，可以直接填写字符串ap-guangzhou，或者引用预设的常量 */
		client, _ := sms.NewClient(credential, "ap-guangzhou", cpf)

		/* 实例化一个请求对象，根据调用的接口和实际情况，可以进一步设置请求参数
		 * 你可以直接查询SDK源码确定接口有哪些属性可以设置
		 * 属性可能是基本类型，也可能引用了另一个数据结构
		 * 推荐使用IDE进行开发，可以方便的跳转查阅各个接口和数据结构的文档说明 */
		request := sms.NewSendSmsRequest()

		/* 基本类型的设置:
		 * SDK采用的是指针风格指定参数，即使对于基本类型你也需要用指针来对参数赋值。
		 * SDK提供对基本类型的指针引用封装函数
		 * 帮助链接：
		 * 短信控制台: https://console.cloud.tencent.com/smsv2
		 * sms helper: https://cloud.tencent.com/document/product/382/3773 */

		/* 短信应用ID: 短信SdkAppId在 [短信控制台] 添加应用后生成的实际SdkAppId，示例如1400006666 */
		request.SmsSdkAppId = common.StringPtr("1400606387")
		/* 短信签名内容: 使用 UTF-8 编码，必须填写已审核通过的签名，签名信息可登录 [短信控制台] 查看 */
		request.SignName = common.StringPtr("高冷林乐乐公众号")
		/* 国际/港澳台短信 SenderId: 国内短信填空，默认未开通，如需开通请联系 [sms helper] */
		request.SenderId = common.StringPtr("")
		/* 用户的 session 内容: 可以携带用户侧 ID 等上下文信息，server 会原样返回 */
		request.SessionContext = common.StringPtr("mike_lin")
		/* 短信码号扩展号: 默认未开通，如需开通请联系 [sms helper] */
		request.ExtendCode = common.StringPtr("")
		/* 模板参数: 若无模板参数，则设置为空*/
		//生成一个随机6位数
		rand.Seed(time.Now().UnixNano())
		smsCode := fmt.Sprintf("%06d", rand.Int31n(1000000))
		request.TemplateParamSet = common.StringPtrs([]string{smsCode})
		/* 模板 ID: 必须填写已审核通过的模板 ID。模板ID可登录 [短信控制台] 查看 */
		request.TemplateId = common.StringPtr("1226667")
		/* 下发手机号码，采用 E.164 标准，+[国家或地区码][手机号]
		 * 示例如：+8613711112222， 其中前面有一个+号 ，86为国家码，13711112222为手机号，最多不要超过200个手机号*/
		request.PhoneNumberSet = common.StringPtrs([]string{"+86"+req.Phone})

		// 通过client对象调用想要访问的接口，需要传入请求对象
		response, err := client.SendSms(request)
		// 处理异常
		if _, ok := err.(*errors.TencentCloudSDKError); ok {
			fmt.Printf("An API error has returned: %s", err)

		} else{
			fmt.Println("短信发送成功！！！！！！")
			// 将 电话号:短信验证码 ,存入到 Redis 数据库
			err := model.SaveSmsCode(req.Phone, smsCode)
			if err != nil {
				fmt.Println("存储短信验证码到redis失败:", err)
				rsp.Errno = utils.RECODE_DBERR
				rsp.Errmsg = utils.RecodeText(utils.RECODE_DBERR)
			}

			rsp.Errno = utils.RECODE_OK
			rsp.Errmsg  = utils.RecodeText(utils.RECODE_OK)
		}
		// 非SDK异常，直接失败。实际代码中可以加入其他的处理。
		if err != nil {
			panic(err)
		}
		b, _ := json.Marshal(response.Response)
		// 打印返回的json字符串
		fmt.Printf("%s", b)
		} else {
		// 校验失败, 发送错误信息
		rsp.Errno = utils.RECODE_DATAERR
		rsp.Errmsg = utils.RecodeText(utils.RECODE_DATAERR)
	}
	return nil
}

func (e *User) Register(ctx context.Context, req *user.RegReq, rsp *user.Response) error {

	// 先校验短信验证码,是否正确. redis 中存储短信验证码.
	err := model.CheckSmsCode(req.Mobile, req.SmsCode)
	if err == nil {

		// 如果校验正确. 注册用户. 将数据写入到 MySQL数据库.
		err = model.RegisterUser(req.Mobile, req.Password)
		if err != nil {
			rsp.Errno = utils.RECODE_DBERR
			rsp.Errmsg = utils.RecodeText(utils.RECODE_DBERR)
		} else {
			rsp.Errno = utils.RECODE_OK
			rsp.Errmsg = utils.RecodeText(utils.RECODE_OK)
		}

	} else {  // 短信验证码错误
		rsp.Errno = utils.RECODE_DATAERR
		rsp.Errmsg = utils.RecodeText(utils.RECODE_DATAERR)
	}

	return nil
}
