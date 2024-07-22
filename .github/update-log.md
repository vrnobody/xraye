### 和XTLS/xray-core的区别
修复windows下iperf3并发测速断流问题（XTLS/Xray-core issue [#3460](https://github.com/XTLS/Xray-core/issues/3460)）  
去除routing相关的adrules/rmrules/sib命令  
添加多个API功能，详见[wiki](https://github.com/vrnobody/xraye/wiki)  

### 详细更新记录

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
