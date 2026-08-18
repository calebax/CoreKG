package excelparser

import (
	"context"
	"testing"

	"github.com/insmtx/corekg/pkgs/testutils"
)

//  Sheet1
//  全市和分区
//  1
//  2
//  3
//  4
//  5
//  6
//  7
//  8
//  9
//  10
//  11
//  12
//  13
//  14
//  指标解释

func TestExcelParser3466(t *testing.T) {
	if true {
		t.Skip("skip test")
		return
	}
	t.Errorf("not implemented ")

	ctx := context.Background()
	epr := &ExcelParser{}
	err := epr.openExcelFile(testutils.TestFilePath("testdata/3466.xlsx"))
	if err != nil {
		t.Errorf("failed to open excel: %v", err)
		return
	}
	_, err = epr.decodeSheetDSLs(ctx, "Sheet1")
	if err != nil {
		t.Errorf("failed to parse excel: %v", err)
		return
	}

	return
}

func TestExcelParserYG4431(t *testing.T) {
	if true {
		t.Skip("skip test")
		return
	}
	t.Errorf("not implemented ")

	ctx := context.Background()
	epr := &ExcelParser{}
	err := epr.openExcelFile(testutils.TestFilePath("testdata/言古研发_需求_20250914104431.xlsx"))
	if err != nil {
		t.Errorf("failed to open excel: %v", err)
		return
	}
	_, err = epr.decodeSheetDSLs(ctx, "Story_list")
	if err != nil {
		t.Errorf("failed to parse excel: %v", err)
		return
	}

}
