# gormc

> auto generate struct for `github.com/jinzhu/gorm` from database

## build

```
go build -ldflags "-s -w" -o /usr/local/bin/gormc && upx -9 /usr/local/bin/gormc
```

```
gorm --help

gormc -  1.0.0
Usage of gormc:
  -u, --user string             数据库连接用户名 (default "root")
  -k, --password string         数据库连接密码 (default "root")
  -h, --host string             数据库连接主机和端口 (default "localhost:3306")
  -n, --name string             数据库名
  -o, --output string           生成的文件路径和文件名，默认在当前目录下的 models/数据库名称.go
      --pkg string              包名，默认是目录名称
      --remove-prefix strings   移除前缀 (default [t_,tab_,tb_])
  -v, --verbose                 输出详细信息
```
