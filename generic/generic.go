package generic

import (
	"errors"
	"fmt"
)

type Grant struct {
	Type  string
	Where string
	Who   string
}

func (d *TableDef) String() string {
	result := ""
	for cname, c := range d.Columns {
		result = fmt.Sprintf("%s\n%q: %q", result, cname, c.Type)
	}
	return result
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
	Tables map[string]*TableDef
}

type ColumnDef struct {
	Name      string
	Type      string
	Default   string
	Precision int
	Scale     int
}

type ColumnsDef map[string]*ColumnDef

type TableDef struct {
	Name            string
	Columns         ColumnsDef
	SelectStatement string
}

func Errorf(err error, format string, a ...any) error {
	return errors.Join(err, fmt.Errorf(format, a...))
}
