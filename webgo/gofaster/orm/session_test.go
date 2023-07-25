package orm

import (
	"fmt"
	"testing"
)

type User struct {
}

func TestName(t *testing.T) {
	fmt.Println(Name("UserName"))
	fmt.Println(Name("UserNameH"))

}
