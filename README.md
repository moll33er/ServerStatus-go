# ServerStatus

基于[https://github.com/P3terChan/ServerStatus-V](https://github.com/P3terChan/ServerStatus-V)的客户端改版。

由于原版使用Python编写，在实际使用中会出现内存、硬盘、流量等统计不准确或无法统计情况，故此使用GoLang重新编写，可与原版服务端完美结合，无需更改。

### 一键安装

```shell
wget -qO- --no-check-certificate https://raw.githubusercontent.com/moll33er/ServerStatus-go/master/install.sh | bash
```



## 存在问题

程序在Alpine下，获取Swap时不准确，获取到的大小都是15.98G，这个问题后期再改。

## 使用说明

直接下载对应的可执行程序，并下载status.ini至同目录下，修改status.ini后执行即可。



## Fork说明

移除了失效的网络流量计算方法，改为直接读取内核信息，流量统计为开机到现在所有的流量，重启服务器清零。
