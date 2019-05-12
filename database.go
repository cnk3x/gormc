package main

import (
	"io"
	"regexp"
	"strings"
)

//Database 数据库接口
type Database interface {
	GetSchema(database string) ([]*TableSchema, error)
	GenerateStruct(w io.Writer) error
}

//TableSchema 表结构
type TableSchema struct {
	Name    string
	Comment string
	Columns []*ColumnSchema
}

//ColumnSchema 列结构
type ColumnSchema struct {
	TableName  string
	ColName    string
	ColType    string
	Nullable   bool
	PrimaryKey bool
	AutoIncr   bool
	Comment    string
}

func convertToGoName(name string) string {
	if name == "" {
		return ""
	}
	name = strings.Trim(strings.ToLower(name), "_")
	name = regexp.MustCompile(`(_[[:alnum:]]+?)`).ReplaceAllStringFunc(name,
		func(s string) string {
			return strings.ToUpper(strings.Trim(s, "_"))
		})
	return strings.ToUpper(name[:1]) + name[1:]
}

func convertToGoType(dataTypeString string) string {
	dataType := dataTypeString
	if idx := strings.Index(dataType, "("); idx != -1 {
		dataType = dataType[:idx]
	}

	unsigned := strings.Contains(dataTypeString, "unsigned")

	switch dataType {
	case "varchar", "nvarchar", "text", "longtext", "char", "tinytext":
		return "string"
	case "bigint":
		if unsigned {
			return "uint64"
		}
		return "int64"
	case "int", "smallint", "mediumint":
		if unsigned {
			return "uint"
		}
		return "int"
	case "tinyint":
		if strings.Contains(dataTypeString, "(1)") {
			return "bool"
		}
		if unsigned {
			return "uint8"
		}
		return "byte"
	case "decimal", "float", "double":
		return "float64"
	case "datetime", "date", "timestamp", "year":
		return "time.Time"
	default:
		return "interface{}"
	}
}
