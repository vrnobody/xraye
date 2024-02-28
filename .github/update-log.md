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

修改go.mod后执行`go mod tidy`更新go.sum  
版本信息：core\core.go  