[![Unit Tests][1]][2] [![Release][3]][4] [![Total Downloads][5]][6] [![License][7]][8]  

[1]: https://github.com/vrnobody/xraye/actions/workflows/test.yml/badge.svg "Unit Tests Status Badge"
[2]: https://github.com/vrnobody/xraye/actions/workflows/test.yml "Workflow"
[3]: https://img.shields.io/github/release/vrnobody/xraye.svg "Release Badge"
[4]: https://github.com/vrnobody/xraye/releases/latest "Releases"
[5]: https://img.shields.io/github/downloads/vrnobody/xraye/total.svg "Total Downloads Badge"
[6]: https://somsubhra.github.io/github-release-stats/?username=vrnobody&repository=xraye&per_page=30 "Download Details"
[7]: https://img.shields.io/github/license/vrnobody/xraye.svg "Licence Badge"
[8]: https://github.com/vrnobody/xraye/blob/main/LICENSE "Licence"

这个fork在xray-core的基础上添加了以下api子命令：  
```bash
rmi           增加删除tag为空的inbound功能
rmo           增加删除tag为空的outbound功能
getr          获取routing配置，以json格式输出
setr          替换routing配置
adu           向inbounds添加用户
rmu           从inbounds删除用户
```
[详细用法说明wiki](https://github.com/vrnobody/xraye/wiki)  
[可执行文件下载](https://github.com/vrnobody/xraye/releases)  
[更新日志](./.github/update-log.md)  
  
这个fork遵循Mozilla Public License Version 2.0协议，简单来说你想怎么用都行。  
这个只是试验性质的fork，不会跟随xray-core同步更新。   
原版[README.md](./README-xtls.md)  
