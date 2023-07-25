package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type FsSession struct {
	db           *FsDB
	tableName    string   //表名
	fieldName    []string //插入字段名
	placeHolder  []string //占位符
	values       []any    //值
	updateParams strings.Builder
	whereParams  strings.Builder
	whereValues  []any
	beginTx      bool
	tx           *sql.Tx
}

func (s *FsSession) TableName(name string) *FsSession {
	s.tableName = name
	return s
}

func (s *FsSession) Insert(data any) (int64, int64, error) {
	d := make([]any, 0)
	d = append(d, data)
	return s.InsertBatch(d)
}

func (s *FsSession) fileNames(data any, index int) {
	//反射获取数据字段名
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() != reflect.Ptr {
		panic(errors.New("data must be a pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	if index == 0 {
		if s.tableName == "" {
			s.tableName = s.db.Prefix + strings.ToLower(tVar.Name())
		}
	}
	for i := 0; i < tVar.NumField(); i++ {
		fieldName := tVar.Field(i).Name
		tag := tVar.Field(i).Tag
		sqlTag := tag.Get("fsorm")
		if sqlTag == "" {
			sqlTag = strings.ToLower(Name(fieldName))
		} else {
			if strings.Contains(sqlTag, "auto_increment") {
				//自增长ID
				continue
			}
			if strings.Contains(sqlTag, ",") {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
		}
		if strings.ToLower(sqlTag) == "id" && IsAutoId(vVar.Field(i).Interface()) {
			continue
		}
		if index == 0 {
			s.fieldName = append(s.fieldName, sqlTag)
			s.placeHolder = append(s.placeHolder, "?")
		}
		s.values = append(s.values, vVar.Field(i).Interface())
	}
}
func (s *FsSession) InsertBatch(data []any) (int64, int64, error) {
	//insert into %s (%s) values (%s),(%s)
	if len(data) <= 0 {
		return -1, -1, errors.New("no data insert")
	}
	for i, value := range data {
		s.fileNames(value, i)
	}
	query := fmt.Sprintf("insert into %s (%s) values ",
		s.tableName, strings.Join(s.fieldName, ","))
	var sb strings.Builder
	sb.WriteString(query)
	for index, _ := range data {
		sb.WriteString("(")
		sb.WriteString(strings.Join(s.placeHolder, ","))
		sb.WriteString(")")
		if index < len(data)-1 {
			sb.WriteString(",")
		}
	}
	s.db.logger.Info(sb.String())
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.db.db.Prepare(sb.String())
	}
	defer stmt.Close()
	if err != nil {
		return -1, -1, err
	}
	r, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}
	lastInsertId, err := r.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return lastInsertId, rowsAffected, nil
}
func (s *FsSession) Update(data ...any) (int64, int64, error) {
	// update("age",12) or update(user)
	if len(data) <= 0 || len(data) > 2 {
		return -1, -1, errors.New("param not valid")
	}
	//update table set age=?,name=? where id = ?
	if len(data) == 2 {
		s.updateParams.WriteString(data[0].(string))
		s.updateParams.WriteString("= ?")
	} else {
		updateData := data[0]
		t := reflect.TypeOf(updateData)
		v := reflect.ValueOf(updateData)
		if t.Kind() != reflect.Ptr {
			panic(errors.New("data must be a pointer"))
		}
		tVar := t.Elem()
		vVar := v.Elem()
		if s.tableName == "" {
			s.tableName = s.db.Prefix + strings.ToLower(tVar.Name())
		}
		for i := 0; i < tVar.NumField(); i++ {
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			sqlTag := tag.Get("fsorm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(sqlTag, "auto_increment") {
					//自增长ID
					continue
				}
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}
			if strings.ToLower(sqlTag) == "id" && IsAutoId(vVar.Field(i).Interface()) {
				continue
			}
			if s.updateParams.String() != "" {
				s.updateParams.WriteString(",")
			}
			s.updateParams.WriteString(sqlTag)
			s.updateParams.WriteString(" = ? ")
			s.values = append(s.values, vVar.Field(i).Interface())
		}
	}
	query := fmt.Sprintf("update %s set %s", s.tableName, s.updateParams)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParams.String())
	s.db.logger.Info(sb.String())
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.db.db.Prepare(sb.String())
	}
	defer stmt.Close()
	if err != nil {
		return -1, -1, err
	}
	s.values = append(s.values, s.whereValues...)
	r, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}
	lastInsertId, err := r.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return lastInsertId, rowsAffected, nil
}

func (s *FsSession) SelectOne(data any, fields ...string) error {
	typeOf := reflect.TypeOf(data)
	if typeOf.Kind() != reflect.Pointer {
		return errors.New("data must be a pointer")
	}

	//select * from table where ?
	if s.tableName == "" {
		tVar := typeOf.Elem()
		s.tableName = s.db.Prefix + strings.ToLower(tVar.Name())
	}
	fieldStr := "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	query := fmt.Sprintf("select %s from %s ", fieldStr, s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParams.String())
	s.db.logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return err
	}
	rows, err := stmt.Query(s.whereValues...)
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]any, len(columns))
	fieldScan := make([]any, len(columns))
	for i := 0; i < len(fieldScan); i++ {
		fieldScan[i] = &values[i]
	}
	if rows.Next() {
		err := rows.Scan(fieldScan)
		if err != nil {
			return err
		}
		vVar := reflect.ValueOf(data).Elem()
		tVar := typeOf.Elem()
		for i := 0; i < tVar.NumField(); i++ {
			name := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			sqlTag := tag.Get("fsorm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(name)
			} else {
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}
			for j, column := range columns {
				if column == sqlTag {
					target := values[j]
					targetValue := reflect.ValueOf(target) //获取到any类型的value
					fieldType := tVar.Field(i).Type
					value := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
					vVar.Field(i).Set(value)
				}
			}
		}
	}
	return nil
}

func (s *FsSession) Select(data any, fields ...string) ([]any, error) {
	typeOf := reflect.TypeOf(data)
	if typeOf.Kind() != reflect.Pointer {
		return nil, errors.New("data must be a pointer")
	}

	//select * from table where ?
	if s.tableName == "" {
		tVar := typeOf.Elem()
		s.tableName = s.db.Prefix + strings.ToLower(tVar.Name())
	}
	fieldStr := "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	query := fmt.Sprintf("select %s from %s ", fieldStr, s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParams.String())
	s.db.logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(s.whereValues...)
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := make([]any, 0)
	for {
		values := make([]any, len(columns))
		fieldScan := make([]any, len(columns))
		for i := 0; i < len(fieldScan); i++ {
			fieldScan[i] = &values[i]
		}
		if rows.Next() {
			err := rows.Scan(fieldScan)
			if err != nil {
				return nil, err
			}
			data := reflect.New(typeOf.Elem()).Interface()
			vVar := reflect.ValueOf(data).Elem()
			tVar := typeOf.Elem()
			for i := 0; i < tVar.NumField(); i++ {
				name := tVar.Field(i).Name
				tag := tVar.Field(i).Tag
				sqlTag := tag.Get("fsorm")
				if sqlTag == "" {
					sqlTag = strings.ToLower(name)
				} else {
					if strings.Contains(sqlTag, ",") {
						sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
					}
				}
				for j, column := range columns {
					if column == sqlTag {
						target := values[j]
						targetValue := reflect.ValueOf(target) //获取到any类型的value
						fieldType := tVar.Field(i).Type
						value := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
						vVar.Field(i).Set(value)
					}
				}
			}
			result = append(result, data)
		} else {
			break
		}
	}
	return result, nil
}

func (s *FsSession) Delete() (int64, error) {
	query := fmt.Sprintf("delete from %s", s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParams.String())
	s.db.logger.Info(sb.String())

	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.db.db.Prepare(sb.String())
	}
	if err != nil {
		return 0, err
	}
	exec, err := stmt.Exec(s.whereValues...)
	if err != nil {
		return 0, err
	}
	return exec.RowsAffected()
}

func (s *FsSession) Where(field string, value any) *FsSession {
	if s.whereParams.String() == "" {
		s.whereParams.WriteString("where ")
	}
	s.whereParams.WriteString(field)
	s.whereParams.WriteString(" = ")
	s.whereParams.WriteString("?")
	s.whereValues = append(s.whereValues, value)
	return s
}
func (s *FsSession) Like(field string, value any) *FsSession {
	if s.whereParams.String() == "" {
		s.whereParams.WriteString("where ")
	}
	s.whereParams.WriteString(field)
	s.whereParams.WriteString(" like ")
	s.whereParams.WriteString("?")
	s.whereValues = append(s.whereValues, "%"+value.(string)+"%")
	return s
}

func (s *FsSession) LikeRight(field string, value any) *FsSession {
	if s.whereParams.String() == "" {
		s.whereParams.WriteString("where ")
	}
	s.whereParams.WriteString(field)
	s.whereParams.WriteString(" like ")
	s.whereParams.WriteString("?")
	s.whereValues = append(s.whereValues, value.(string)+"%")
	return s
}
func (s *FsSession) LikeLeft(field string, value any) *FsSession {
	if s.whereParams.String() == "" {
		s.whereParams.WriteString("where ")
	}
	s.whereParams.WriteString(field)
	s.whereParams.WriteString(" like ")
	s.whereParams.WriteString("?")
	s.whereValues = append(s.whereValues, "%"+value.(string))
	return s
}
func (s *FsSession) Group(field ...string) *FsSession {
	//group by aa,bb
	s.whereParams.WriteString(" group by ")
	s.whereParams.WriteString(strings.Join(field, ","))
	return s
}
func (s *FsSession) Order(field ...string) *FsSession {
	//order by aa,bb
	s.whereParams.WriteString(" order by ")
	s.whereParams.WriteString(strings.Join(field, ","))
	return s
}
func (s *FsSession) Or(field string, value any) *FsSession {
	s.whereParams.WriteString(" or ")
	return s
}
func (s *FsSession) And(field string, value any) *FsSession {
	s.whereParams.WriteString(" and ")
	return s
}

func (s *FsSession) Aggregate(funName, field string) (int64, error) {
	var fieldSb strings.Builder
	fieldSb.WriteString(funName)
	fieldSb.WriteString("(")
	fieldSb.WriteString(field)
	fieldSb.WriteString(")")
	query := fmt.Sprintf("select %s from %s ", fieldSb.String(), s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParams.String())
	s.db.logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return -1, err
	}
	rows := stmt.QueryRow(s.whereValues...)
	if rows.Err() != nil {
		return -1, rows.Err()
	}
	var result int64
	err = rows.Scan(&result)
	if err != nil {
		return -1, err
	}
	return result, nil
}
func (s *FsSession) Count() (int64, error) {
	return s.Aggregate("count", "*")
}

// 原生sql支持
func (s *FsSession) Exec(raw string, values ...any) (int64, error) {
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(raw)
	} else {
		stmt, err = s.db.db.Prepare(raw)
	}
	if err != nil {
		return 0, err
	}
	exec, err := stmt.Exec(values...)
	if err != nil {
		return 0, err
	}
	if strings.Contains(strings.ToLower(raw), "insert") {
		return exec.LastInsertId()
	}
	return exec.RowsAffected()
}

func (s *FsSession) Begin() error {
	tx, err := s.db.db.Begin()
	if err != nil {
		return err
	}
	s.beginTx = true
	s.tx = tx
	return nil
}
func (s *FsSession) Commit() error {
	err := s.tx.Commit()
	if err != nil {
		return err
	}
	s.beginTx = false
	return nil
}

func (s *FsSession) Rollback() error {
	err := s.tx.Rollback()
	if err != nil {
		return err
	}
	s.beginTx = false
	return nil
}
