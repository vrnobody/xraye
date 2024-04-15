#### golang
go 1.20支持的最高版本：  
```bash
github.com/quic-go/quic-go v0.40.1
```
修改go.mod后执行`go mod tidy`更新go.sum  

#### xray version
版本信息：core\core.go  

#### git merge
update main branch  
```bash
git fetch upstream
git checkout main
git merge upstream/main
```
revert commits  
```bash
git checkout <branch-for-revert>
git merge main
git revert <commit id>  
# fix conflicts
git revert --continue
git checkout api
git merge <branch-for-revert>
# fix conflicts
```

#### protobuf
如果需要修改 protobuf，例如增加新配置项，请运行：go generate core/proto.go  
安装protoc
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

