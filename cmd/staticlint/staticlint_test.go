package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func TestOsExitCheckAnalyzer(t *testing.T) {
	// функция analysistest.Run применяет тестируемый анализатор OsExitCheckAnalyzer
	// к пакетам из папки testdata и проверяет ожидания
	// ./... — проверка всех поддиректорий в testdata
	analysistest.Run(t, analysistest.TestData()+"/osexit", OsExitCheckAnalyzer, "./...")
}

func TestAppendChecks(t *testing.T) {
	checks := map[string]bool{
		"ST1005": true,
		"ST1000": true,
		"ST1020": true,
		"ST1013": true,
		"S1008":  true,
		"S1021":  true,
	}
	appendChecks(staticcheck.Analyzers, checks)
	appendChecks(stylecheck.Analyzers, checks)
	appendChecks(simple.Analyzers, checks)
	appendChecks(quickfix.Analyzers, checks)
}

func TestAppendOtherPublicChecks(t *testing.T) {
	appendOtherPublicChecks()
}
func TestAppendStaticcheckIoChecks(t *testing.T) {
	checks := map[string]bool{
		"ST1005": true,
		"ST1000": true,
		"ST1020": true,
		"ST1013": true,
		"S1008":  true,
		"S1021":  true,
	}
	appendStaticcheckIoChecks(checks)
}

func TestAppendPassesChecks(t *testing.T) {
	appendPassesChecks()
}
func TestAppendCustomOsExitCheck(t *testing.T) {
	appendCustomOsExitCheck()
}

func TestMain(t *testing.T) {
	a := copylock.Analyzer
	analysistest.Run(t, analysistest.TestData()+"/pkg", a)
}
