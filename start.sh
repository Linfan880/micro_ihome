#! /bin/bash

#启动redis/tracker/storage/nginx/
echo "正在启动redis......."
sudo redis-server /home/linfan/redis.conf
echo "redis启动完成"
echo "正在启动tracker......"
sudo fdfs_trackerd /etc/fdfs/tracker.conf
echo "tracker启动完成"
echo "正在启动storage........"
sudo fdfs_storaged /etc/fdfs/storage.conf
echo "storage启动完成"
echo "正在启动nginx.........."
sudo /usr/local/nginx/sbin/nginx
echo "nginx启动完成"
