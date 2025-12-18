### 和XTLS/xray-core的区别
修改一些 API 命令，详见 [wiki](https://github.com/vrnobody/xraye/wiki)  
移除 wireguard 相关协议  
移除 shadowsocks 相关协议  
移除 reverse 功能  
移除 vless 的 reverse 功能  
移除 yaml/toml 配置格式，只支持json配置格式  
移除 routing 相关的 adrules/rmrules/sib 命令（和现有命令冲突）  

### 详细更新记录

#### exp #28
更新 xray-core 至 v25.12.8 (commit 81f8f398)  

#### exp #27
更新 xray-core 至 v25.07.26（commit c569f478）  

#### exp #26
更新 xray-core 至 v25.07.24（commit 4f45c5fa）  

#### exp #25
更新 xray-core 至 v25.05.16（commit b80e3196）  
移除 geti geto 命令，xray-core已添加相应命令  

#### exp #24
更新 xray-core 至 v25.04.30（commit 87ab8e5）  

#### exp #23
更新 xray-core 至 v25.03.31（commit ab5d7cf3）  
屏蔽 websocket 传输协议的“过时”警告  
移除 [#3460](https://github.com/XTLS/Xray-core/issues/3460) 补丁  
移除 win7 发布，请到 [Xray-core](https://github.com/XTLS/Xray-core) 下载原版  

#### exp #22
更新xray-core至v24.10.31 (commit 2c72864)  
删除getu(查询inbound用户)命令（因为xray-core新增了相关命令）  

#### exp #21
更新xray-core至v24.10.16 (commit e4939dc)  

#### exp #20
更新xray-core至v24.9.19 (commit 3632e83)  
移除wireguard相关协议  
移除shadowsocks相关协议  
移除reverse功能  
exe文件从26MiB减少到21MiB  

#### exp #19
更新xray-core至v1.8.23 (commit c0c23fd)  

#### exp #18
更新xray-core至v1.8.21 (commit c27d652)  

#### exp #17
更新xray-core至v1.8.20 (commit 8deb953)  

#### exp #16
更新xray-core至v1.8.19 (commit b277bac)  

#### exp #15
更新xray-core至v1.8.17 (commit 558cfcc)  
添加单独的win7发布文件  
加回Mac arm系列发布文件  

#### exp #14
更新xray-core至v1.8.16+ (commit e13f9f5)  
修复windows下iperf3并发测速断流问题 Xray-core issue #3460  

#### exp #13
更新xray-core至v1.8.15  
移除Mac arm系列发布文件，因为golang v1.21.4不支持新的Mac CPU  

#### exp #12
切换至golang v1.21.4  
恢复MultipathTCP功能  

#### exp #11
更新xray-core至v1.8.13  

#### exp #10
更新xray-core至v1.8.10  
去除routing相关的adrules/rmrules/sib命令，因为和getr/setr命令冲突  

#### exp #9
更新xray-core至v1.8.9  

#### exp #8
更新xray-core至v1.8.8  

#### exp #7
降级golang至1.20（支持win7）  
禁用MultipathTCP功能  
移除slices包依赖  
修复负载均衡并发报错退出的bug  

#### exp #6
修改round-robin均衡策略  
更新xray-core至v1.8.7  
修复内存泄露  

#### exp #5
给增删in(out)bounds加上线程锁  

#### exp #4
修复bug，基本可用  
