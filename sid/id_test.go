package sid

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	id := MustNewString(5)
	fmt.Println(id)
}
