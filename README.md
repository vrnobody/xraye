这个fork在xray-core的基础上添加了以下api子命令：  
```bash
geti          获取全部inbounds配置，以json格式输出
geto          获取全部outbounds配置以json格式输出
rmi           增加删除tag为空的inbound功能
rmo           增加删除tag为空的outbound功能
getr          获取routing配置，以json格式输出
setr          替换routing配置
getu          获取inbound的用户信息，只支持vmess, vless, trojan三种协议
adu           向inbounds添加用户
rmu           从inbounds删除用户
```
[详细用法说明wiki](https://github.com/vrnobody/Xray-core/wiki)  
[可执行文件下载](https://github.com/vrnobody/Xray-core/releases)  
[更新日志](https://github.com/vrnobody/Xray-core/blob/api/.github/update-log.md)  
  
这个fork遵循Mozilla Public License Version 2.0协议，简单来说你想怎么用都行。  
这个只是试验性质的fork，不会跟随xray-core同步更新。   
原版[README.md](https://github.com/vrnobody/Xray-core/blob/api/README-xtls.md)  