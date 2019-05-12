package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/pflag"

	_ "github.com/go-sql-driver/mysql"
)

const (
	//VERSION 版本号
	VERSION = "1.0.0"
)

func main() {
	var (
		usr     string
		pwd     string
		ip      string
		name    string
		output  string
		pkg     string
		prefix  []string
		verbose bool
		fmtTool string
	)

	fmt.Println("gormc - ", VERSION)

	log.SetFlags(log.Ltime)

	pflag.CommandLine.SortFlags = false
	pflag.StringVarP(&usr, "user", "u", "root", "数据库连接用户名")
	pflag.StringVarP(&pwd, "password", "k", "root", "数据库连接密码")
	pflag.StringVarP(&ip, "host", "h", "localhost:3306", "数据库连接主机和端口")
	pflag.StringVarP(&name, "name", "n", "", "数据库名")
	pflag.StringVarP(&output, "output", "o", "", "生成的文件路径和文件名，默认在当前目录下的 models/数据库名称.go")
	pflag.StringVar(&pkg, "pkg", "", "包名，默认是目录名称")
	pflag.StringSliceVar(&prefix, "remove-prefix", []string{"t_", "tab_", "tb_"}, "移除前缀")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "输出详细信息")
	pflag.StringVar(&fmtTool, "format-tool", "goimports", "格式化工具")
	pflag.Parse()

	if name == "" {
		pflag.Usage()
		os.Exit(1)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/", usr, pwd, ip))
	if err == nil {
		err = db.Ping()
	}
	if err != nil {
		log.Fatal(err)
	}

	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)

	tmp := filepath.Join(tmpDir, name+".go")
	w, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	if output == "" {
		output = "models/" + name + ".go"
	}

	output, _ = filepath.Abs(output)
	dir := filepath.Dir(output)
	if pkg == "" {
		pkg = filepath.Base(dir)
	}

	var d Database = &mysqlDatabase{conn: db, pkg: pkg, database: name, prefix: prefix}

	err = d.GenerateStruct(w)
	if err != nil {
		log.Fatal(err)
	}

	v, err := exec.Command(fmtTool, tmp).CombinedOutput()
	if err == nil {
		if err = os.MkdirAll(dir, 0744); err == nil {
			err = ioutil.WriteFile(output, v, 0644)
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Println("已生成：", output)
}
