package main

import (
	"time"
)

// MyConfig is an example struct highlighting how an implementation can have
// automatic support in the apcore configuration file.
//
// Note that values put into the configuration are ones that remain constant
// throughout the running lifetime of the apcore.Application.
type MyConfig struct {
	FieldS string    `ini:"my_test_app_s" comment:"First test field"`
	FieldT int       `ini:"my_test_app_t" comment:"Second test field"`
	FieldU time.Time `ini:"my_test_app_u" comment:"Third test field"`
}
