package log

import (
	"testing"

	"github.com/hashicorp/go-hclog"
)

func TestHclog2LoggerImplementsInterfaces(t *testing.T) {
	var logger interface{} = NewHclogAdapter(L())
	if _, ok := logger.(hclog.Logger); !ok {
		t.Fatalf("logger does not implement hclog.Logger")
	}
}
