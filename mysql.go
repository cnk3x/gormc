package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"strings"
)

type mysqlDatabase struct {
	conn     *sql.DB
	database string
	pkg      string
	prefix   []string
}

func (d *mysqlDatabase) GetSchema(database string) ([]*TableSchema, error) {
	var tables = make([]*TableSchema, 0)
	tabMap := make(map[string]*TableSchema)

	rows, err := d.conn.Query(`select 
		TABLE_NAME as name,
		TABLE_COMMENT as comment 
		from information_schema.TABLES 
        where TABLE_SCHEMA=?;`, database)

	if err != nil {
		if err == sql.ErrNoRows {
			return tables, nil
		}
		return tables, err
	}

	for rows.Next() {
		tab := new(TableSchema)
		err = rows.Scan(&tab.Name, &tab.Comment)
		if err != nil {
			log.Println(err)
		}
		tables = append(tables, tab)
		tabMap[tab.Name] = tab
	}

	rows, err = d.conn.Query(`select 
		TABLE_NAME as table_name,
		COLUMN_NAME as col_name, 
		COLUMN_TYPE as col_type, 
		(case IS_NULLABLE when 'NO' then 0 else 1 end) as nullable, 
		(case COLUMN_KEY when 'PRI' then 1 else 0 end) as primary_key,
		(case EXTRA when 'auto_increment' then 1 else 0 end) as auto_incr,
		COLUMN_COMMENT as comment
		from information_schema.COLUMNS 
		where TABLE_SCHEMA=?
        order by TABLE_NAME, ORDINAL_POSITION;`, database)

	if err != nil {
		if err == sql.ErrNoRows {
			return tables, nil
		}
		return tables, err
	}

	for rows.Next() {
		col := new(ColumnSchema)
		err = rows.Scan(&col.TableName, &col.ColName, &col.ColType, &col.Nullable, &col.PrimaryKey, &col.AutoIncr, &col.Comment)
		if tab, find := tabMap[col.TableName]; find {
			tab.Columns = append(tab.Columns, col)
		}
	}

	return tables, nil
}

func (d *mysqlDatabase) GenerateStruct(w io.Writer) error {
	tables, err := d.GetSchema(d.database)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, `/* auto generate by gormc, http://github.com/shuxs/gormc */`)
	fmt.Fprintln(w, "package ", d.pkg)
	fmt.Fprintln(w, `import ("time")`)

	for _, tab := range tables {
		name := tab.Name

		if len(d.prefix) > 0 {
			for _, prefix := range d.prefix {
				name = strings.TrimPrefix(name, prefix)
			}
		}

		var (
			hasCreateAt = false
			hasUpdateAt = false
		)

		structName := convertToGoName(name)
		log.Printf("%s -> %s: %s", tab.Name, structName, tab.Comment)
		fmt.Fprintf(w, "//%s -> %s %s\n", structName, tab.Name, tab.Comment)
		fmt.Fprintf(w, "type %s struct {\n", structName)

		for _, col := range tab.Columns {
			if len(col.Comment) >= 50 {
				fmt.Fprintf(w, "// %s\n", col.Comment)
			}

			goFieldName := convertToGoName(col.ColName)
			goFieldType := convertToGoType(col.ColType)

			fmt.Fprintf(w, " %s %s", goFieldName, goFieldType)

			fmt.Fprint(w, "`gorm:")
			fmt.Fprint(w, `"`)
			fmt.Fprintf(w, "column:%s;", col.ColName)
			//主键
			if col.PrimaryKey {
				fmt.Fprint(w, "primary_key;")
			}
			//自增
			if col.AutoIncr {
				fmt.Fprint(w, "auto_increment;")
			}
			//数据类型
			fmt.Fprintf(w, "%s;", col.ColType)
			//是否允许空
			if !col.Nullable {
				fmt.Fprint(w, "not null;")
			}
			fmt.Fprintf(w, `" json:"%s,omitempty"`, col.ColName)
			fmt.Fprint(w, "`")

			if col.Comment != "" && len(col.Comment) < 50 {
				fmt.Fprintf(w, "// %s", col.Comment)
			}
			fmt.Fprintln(w)

			if goFieldName == "UpdateAt" {
				hasUpdateAt = goFieldType == "int64"
			}

			if goFieldName == "CreateAt" {
				hasCreateAt = goFieldType == "int64"
			}
		}

		fmt.Fprintln(w, `}`)

		fmt.Fprintf(w, `func (%s) TableName() string { return "%s" }`, structName, tab.Name)
		fmt.Fprintln(w)

		if hasCreateAt {
			fmt.Fprintf(w, `func (i %s) BeforeCreate() error { i.CreateAt = time.Now().Unix(); return nil }`, structName)
			fmt.Fprintln(w)
		}

		if hasUpdateAt {
			fmt.Fprintf(w, `func (i %s) BeforeUpdate() error { i.UpdateAt = time.Now().Unix(); return nil }`, structName)
			fmt.Fprintln(w)
		}
	}
	return nil
}