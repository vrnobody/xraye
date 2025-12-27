[![Unit Tests][1]][2] [![Release][3]][4] [![Total Downloads][5]][6] [![License][7]][8]  

[1]: https://github.com/vrnobody/xraye/actions/workflows/test.yml/badge.svg "Unit Tests Status Badge"
[2]: https://github.com/vrnobody/xraye/actions/workflows/test.yml "Workflow"
[3]: https://img.shields.io/github/release/vrnobody/xraye.svg "Release Badge"
[4]: https://github.com/vrnobody/xraye/releases/latest "Releases"
[5]: https://img.shields.io/github/downloads/vrnobody/xraye/total.svg "Total Downloads Badge"
[6]: https://somsubhra.github.io/github-release-stats/?username=vrnobody&repository=xraye&per_page=30 "Download Details"
[7]: https://img.shields.io/github/license/vrnobody/xraye.svg "Licence Badge"
[8]: https://github.com/vrnobody/xraye/blob/main/LICENSE "Licence"

这个 fork 在 xray-core 的基础上添加了以下子命令：  
```bash
service   prober      高并发批量测速
service   latency     测速单个配置文件

api       rmi         增加删除 tag 为空的 inbound 功能
api       rmo         增加删除 tag 为空的 outbound 功能
api       getr        获取 routing 配置，以 json 格式输出
api       setr        替换 routing 配置
```

移除了 wireguard，shadowsocks，reverse 等功能，详见下面的更新日志。

[详细用法说明wiki](https://github.com/vrnobody/xraye/wiki)  
[可执行文件下载](https://github.com/vrnobody/xraye/releases)  
[更新日志](./.github/update-log.md)  
  
这个 fork 遵循 Mozilla Public License Version 2.0 协议，简单来说你想怎么用都行。  
原版 [README.md](./README-xtls.md)  

