package eccco73

import (
	"testing"
	"fmt"
)

func TestTest(t *testing.T) {
	fmt.Println( EnsureAppDataDir() )
	fmt.Println( EnsureAppDir() )
}
