#### golang
修改go.mod后执行`go mod tidy`更新go.sum  
版本信息修改：core\core.go  

更新全部protobuf: 
go run ./infra/vprotogen/main.go -pwd .

过滤merge confict: `cat conflict.txt | grep -v -P "\.pb\.go$"`  

#### protobuf
```bash
# 安装protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

# 更新单个 proto
go generate core/proto.go  
```

#### go 1.20支持的最高版本
```bash
github.com/quic-go/quic-go v0.40.1
gvisor.dev/gvisor v0.0.0-20231104011432-48a6d7d5bd0b
```

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
