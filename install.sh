#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
SKYBLUE='\033[0;36m'
PLAIN='\033[0m'

if [ `whoami` != 'root' ];then
    echo -e "${RED} 错误: ${PLAIN} 请使用root权限运行当前脚本! "
    exit 1
fi

mkdir -p /root/tool/serverstatus/

# 定义版本变量
tag=$(wget -qO- -t1 -T2 "https://api.github.com/repos/moll33er/ServerStatus-go/releases/latest" | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g')
# 下载链接替换为变量
wget -O /root/tool/serverstatus/ServerStatus.tgz https://github.com/moll33er/ServerStatus-go/releases/download/${tag}/ServerStatus_go_linux_all.tgz

tar -zxvf /root/tool/serverstatus/ServerStatus.tgz -C /root/tool/serverstatus/

rm -rf /root/tool/serverstatus/ServerStatus.tgz

chmod +x /root/tool/serverstatus/update.sh /root/tool/serverstatus/ServerStatus-go

cp -f /root/tool/serverstatus/serverstatus.service /usr/lib/systemd/system/serverstatus.service

systemctl start serverstatus

systemctl enable serverstatus

echo -e "${GREEN} 安装成功! ${PLAIN}"
