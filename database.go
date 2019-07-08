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

var commonInitialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSRF":  true,
	"XSS":   true,
}

func convertToGoName(name string) string {
	name = regexp.MustCompile(`([A-Z]+)`).ReplaceAllString(name, "_$1")
	name = strings.ToLower(regexp.MustCompile(`[^[:alnum:]]+`).ReplaceAllString(name, "_"))
	name = regexp.MustCompile(`(_[[:alnum:]]+)`).ReplaceAllStringFunc(name, func(s string) string {
		sUp := strings.ToUpper(s[1:])
		if _, ok := commonInitialisms[sUp]; ok {
			return sUp
		}
		return strings.ToUpper(s[1:2]) + s[2:]
	})
	name = strings.TrimLeftFunc(name, func(r rune) bool { return r > '0' && r < '9' || r == '_' })
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
