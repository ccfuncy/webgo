package dialect

import "reflect"

var dialectsMap = map[string]Dialect{}

type Dialect interface {
	DataTypeOf(typ reflect.Value) string                    // 将相应的数据类型转化为该数据库的数据类型
	TableExistSQL(tableName string) (string, []interface{}) // 返回判断某个表是否存在的sql语句
}

func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}
func GetDialect(name string) (Dialect, bool) {
	dialect, ok := dialectsMap[name]
	return dialect, ok
}
