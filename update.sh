# 定义版本变量
tag=$(wget -qO- -t1 -T2 "https://api.github.com/repos/moll33er/ServerStatus-go/releases/latest" | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g')
# 下载链接替换为变量
wget https://github.com/moll33er/ServerStatus-go/releases/download/${tag}/ServerStatus_go_linux.tgz

systemctl stop serverstatus

tar -zxvf ./ServerStatus_go_linux.tgz

systemctl start serverstatus
