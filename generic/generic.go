package generic

import (
	"errors"
	"fmt"
)

type ColumnDef struct {
	Type string
}

type ColumnsDef map[string]ColumnDef

type TableDef struct {
	Columns ColumnsDef
}

type EngineVendorInfo struct {
	Name string
}

type EngineInfo struct {
	Name    string
	Version string
}

type DbOrigin struct {
	Vendor        EngineVendorInfo
	Engine        EngineInfo
	EngineVersion string
	Dialect       string
	Description   string
}

type TablesDef struct {
	Origin DbOrigin
	Tables map[string]TableDef
}

func Errorf(err error, format string, a ...any) error {
	return errors.Join(err, fmt.Errorf(format, a...))
}
