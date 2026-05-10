#### merge upstream steps
```bash
git fetch upstream
git checkout -b <temp-branch>
git merge upstream/main | tee debug/mlog.txt

# fix merge conflicts and commit changes

# 版本信息修改
core/core.go

go mod tidy

# 更新全部protobuf:
go run -v ./infra/vprotogen/main.go -pwd .

# build and test
# commit changes

git checkout main
git merge <temp-branch>
git branch -d <temp-branch>
```

#### protobuf

```bash
# 安装 protoc-gen
go install -v google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.0

# 下载protoc，解压到$PATH的任意可执行目录内
https://github.com/protocolbuffers/protobuf/releases/download/v28.2/protoc-28.2-linux-x86_64.zip

# 更新全部protobuf:
go run -v ./infra/vprotogen/main.go -pwd .

# 更新单个 proto
go generate core/proto.go
```

#### git thing

```bash
# 查看差异文件名：
git diff origin/main --name-only | grep -v -P "\.pb\.go$" | less
```

```bash
# merge upstream

git fetch upstream
git checkout -b <temp-branch>
git merge upstream/main | tee debug/mlog.txt

# fix merge conflicts and commit changes

git checkout main
git merge <temp-branch>
git branch -d <temp-branch>
```

```bash
# revert commits

git checkout <branch-for-revert>
git merge main
git revert <commit id>
# fix conflicts
git revert --continue
git checkout api
git merge <branch-for-revert>
# fix conflicts
```

#### go 1.20支持的最高版本

```bash
github.com/quic-go/quic-go v0.40.1
gvisor.dev/gvisor v0.0.0-20231104011432-48a6d7d5bd0b
```
