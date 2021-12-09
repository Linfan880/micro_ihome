# 基于微服务的房屋租赁系统

## 一、简介

​	单体式架构逐渐暴露出一些缺点，如耦合度太高、维护成本大、交付周期长以及可扩展性差等。近年来的微服务架构成为了主流，微服务顾名思义：将单个应用程序作为一组小型服务开发的方法，每个服务程序都在自己的进程中运行，并与轻量级机制（通常是HTTP资源API）进行通信。这些服务是围绕业务功能构建的，即，“分而治之，合而用之”。每个服务只负责单一职责即可，假如某一个服务宕机，并不会影响其他服务正常运行。因此以微服务的理念开发一个房屋租赁系统，可以满足人们日益增长的租房需求，缓解租房紧张。

## 二、技术栈

​	**Golang + Go-micro + Consul + Grpc + Protobuf + Gin + Mysql + Redis + FastDFS + Nginx**

​	以**Golang**为开发语言，微服务框架采用**Go-micro**，该框架集成了用于**Grpc** 的远程服务调用。采用**Consul**作为服务发现，每一个模块的服务会注册到**Consul**上使其为客户端提供服务。**Protobuf**作为前后端的交流协议。相比于**Json**更加轻量，并且效率较高。**Gin**作为前端框架负责路由的跳转，并提供后端需要处理的格式数据。**FastDFS**作为分布式的存储系统用于存储各类上传图片，最后**Nginx**作为反向代理服务器作为转发请求的中间件。**Mysql**作为持久化的数据库存储用户数据以及订单数据，**Redis**作为临时数据用于图片验证码、短信验证码的校验以及session存储。

## 三、开发细节

​	该项目是以RESTFUL风格的微服务架构为主。以用户模块的注册功能为例，其余微服务流程可模仿此章节进行开发。首先前端路由到注册模块后，前端接收到表单数据或者ajax发来的数据需要通过Gin框架进行参数解析，进一步在F12的网页资源界面提示对应功能没有实现并标记为红色，因此在后端的处理过程中，首先要创建Go-mcrio的微服务，在.proto文件中分析需要实现的功能以及传入传出参数分别是什么，以短信验证码为例，我们传入的一定是一个电话号及获取到的短信验证信息，传出的一般是返回前端的状态码，生成.pb文件后在handler中给出具体实现，同时main函数中一定要将该服务注册到服务发现（Consul）上，才可以对外实现服务。由于项目中的前后端分离，可以在web目录下的controller中模拟客户端，初始化对象并调用远程函数，即在后端实现的handler方法，实现了在像调用本地函数一样调用远程函数。

## 四、页面展示

**注册模块**
![Register](https://github.com/Linfan880/micro_ihome/blob/master/pic/register.png)

**个人信息**
![Userinfo](https://github.com/Linfan880/micro_ihome/blob/master/pic/userinfo.png)

**房屋模块**
![Userinfo](https://github.com/Linfan880/micro_ihome/blob/master/pic/house.png)

**搜索房屋模块**
![Userinfo](https://github.com/Linfan880/micro_ihome/blob/master/pic/search.png)

**服务发现UI页面**
![Userinfo](https://github.com/Linfan880/micro_ihome/blob/master/pic/consul.png)

## 五、如何运行

在运行之前，一定要配置好所有的信息，一定要将go.mod中的包即对应的版本匹配好，关于Nginx、FastDFS的相关配置也一定要做相应检查。

1.打开一个终端,启动consul 命令为：consul agent -dev
2.启动本目录中的脚本：sudo ./start.sh
3.启动微服务(in service, open every services in terminal, then go run main)
4.去浏览器输入：Ip:port/home，就可以看到所得到的页面
5.登录方面的用户名第一次是电话号，可以进入用户模块进行更改用户名操作。
