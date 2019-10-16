package main

import (
	"time"
)

type MyConfig struct {
	FieldS string    `ini:"my_test_app_s" comment:"First test field"`
	FieldT int       `ini:"my_test_app_t" comment:"Second test field"`
	FieldU time.Time `ini:"my_test_app_u" comment:"Third test field"`
}
