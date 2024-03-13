#### golang
修改go.mod后执行`go mod tidy`更新go.sum  

#### xray version
版本信息：core\core.go  

#### git merge upstream
```bash
git fetch upstream
git checkout main
git merge upstream/main
```

#### protobuf
如果需要修改 protobuf，例如增加新配置项，请运行：go generate core/proto.go  

