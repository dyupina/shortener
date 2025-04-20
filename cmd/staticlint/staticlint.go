// Package staticlint implements static analyzer multichecker.
// It combines multiple analyzers from various sources:
//
//	standard static analyzers of the golang.org/x/tools/go/analysis/passes package;
//	all analyzers of the SA class of the staticcheck.io package;
//	at least one analyzer of the remaining classes of the staticcheck.io package;
//	public analyzers;
//	custom analyzer.
//
// The tool reads configuration from
// a `config.json` file to enable specific analyzers and uses
// `multichecker.Main` to execute the selected analyzers.
//
// The analyzer can be launched via Task:
//
//	task sast
//
// or manually:
//
//	go build -o cmd/staticlint/staticlint cmd/staticlint/staticlint.go
//	cmd/staticlint/staticlint ./...
//
// The included analyzers are:
// Standard Go analyzers (golang.org/x/tools/go/analysis/passes):
//
//	appends - appends defines an Analyzer that detects if there is only one variable in append.
//	asmdecl - asmdecl defines an Analyzer that reports mismatches between assembly files and Go declarations.
//	assign - assign defines an Analyzer that detects useless assignments.
//	atomic - atomic defines an Analyzer that checks for common mistakes using the sync/atomic package.
//	atomicalign - atomicalign defines an Analyzer that checks for non-64-bit-aligned arguments to sync/atomic functions.
//	bools - bools defines an Analyzer that detects common mistakes involving boolean operators.
//	buildssa defines an Analyzer that creates SSA representation of an error-free package and returns all functions within it.
//	buildtag - buildtag defines an Analyzer that checks build tags.
//	cgocall - cgocall defines an Analyzer that detects some violations of the cgo pointer passing rules.
//	composite - composite defines an Analyzer that checks for unkeyed composite literals.
//	copylock - copylock defines an Analyzer that checks for locks erroneously passed by value.
//	deepequalerrors - deepequalerrors defines an Analyzer that checks for the use of reflect.DeepEqual with error values.
//	defers - defers defines an Analyzer that checks for common mistakes in defer statements.
//	directive - directive defines an Analyzer that checks known Go toolchain directives.
//	errorsas - errorsas defines an Analyzer that checks that the second argument to errors.As is a pointer to a type implementing error.
//	fieldalignment - fieldalignment defines an Analyzer that detects structs that would use less memory if their fields were sorted.
//	findcall - findcall defines an Analyzer that serves as a trivial example and test of the Analysis API.
//	framepointer - framepointer defines an Analyzer that reports assembly code that clobbers the frame pointer before saving it.
//	httpresponse - httpresponse defines an Analyzer that checks for mistakes using HTTP responses.
//	ifaceassert - ifaceassert defines an Analyzer that flags impossible interface-interface type assertions.
//	inspect - inspect defines an Analyzer that provides an AST inspector for the syntax trees of a package.
//	loopclosure - loopclosure defines an Analyzer that checks for references to enclosing loop variables from within nested functions.
//	lostcancel - lostcancel defines an Analyzer that checks for failure to call a context cancellation function.
//	nilfunc - nilfunc defines an Analyzer that checks for useless comparisons against nil.
//	printf - printf defines an Analyzer that checks consistency of Printf format strings and arguments.
//	reflectvaluecompare defines an Analyzer that checks for accidental use of == or reflect.DeepEqual to compare reflect.Value values.
//	shadow - shadow defines an Analyzer that checks for shadowed variables.
//	shift - shift defines an Analyzer that checks for shifts that exceed the width of an integer.
//	sigchanyzer - sigchanyzer defines an Analyzer that detects misuse of unbuffered signal as argument to signal.Notify.
//	slog - slog defines an Analyzer that checks for mismatched key-value pairs in log/slog calls.
//	sortslice - sortslice defines an Analyzer that checks for calls to sort.Slice that do not use a slice type as first argument.
//	stdmethods - stdmethods defines an Analyzer that checks for misspellings in the signatures of methods similar to well-known interfaces.
//	stringintconv - stringintconv defines an Analyzer that flags type conversions from integers to strings.
//	structtag - structtag defines an Analyzer that checks struct field tags are well formed.
//	testinggoroutine - testinggoroutine defines an Analyzerfor detecting calls to Fatal from a test goroutine.
//	tests - tests defines an Analyzer that checks for common mistaken usages of tests and examples.
//	timeformat - timeformat defines an Analyzer that checks for the use of time.Format or time.Parse calls with a bad format.
//	unmarshal - unmarshal defines an Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
//	unreachable - unreachable defines an Analyzer that checks for unreachable code.
//	unsafeptr - unsafeptr defines an Analyzer that checks for invalid conversions of uintptr to unsafe.Pointer.
//	unusedresult - unusedresult defines an analyzer that checks for unused results of calls to certain pure functions.
//	usesgenerics - usesgenerics defines an Analyzer that checks for usage of generic features added in Go 1.18.
//	waitgroup - waitgroup defines an Analyzer that detects simple misuses of sync.WaitGroup.
//
// Staticcheck analyzers:
//
//	all analyzers of the SA class;
//	ST1005 - Incorrectly formatted error string
//	ST1000 - Incorrect or missing package comment
//	ST1020 - The documentation of an exported function should start with the function's name
//	ST1013 - Should use constants for HTTP error codes, not magic numbers
//	S1008 - Simplify returning boolean expression
//	S1021 - Merge variable declaration and assignment
//
// Public analyzers:
//
//	bodyclose - checks whether HTTP response body is closed and a re-use of TCP connection is not blocked
//	errcheck - checks that you checked errors
//
// Custom analyzer:
//
//	OsExitCheckAnalyzer - detects calls to os.Exit in the main function.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Config is the name of the configuration file that specifies which analyzers to enable.
const Config = `config.json`

// ConfigData contains an array of analyzers.
type ConfigData struct {
	Staticcheck []string
}

// mychecks is a slice of all analyzers that will be executed by multichecker.
var mychecks []*analysis.Analyzer

// appendChecks appends analyzers from the given list if they match the criteria.
func appendChecks(analyzers []*lint.Analyzer, checks map[string]bool) {
	for _, v := range analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") || checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
}

// appendPassesChecks adds standard Go analysis passes to the list of analyzers.
func appendPassesChecks() {
	mychecks = []*analysis.Analyzer{appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		slog.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		usesgenerics.Analyzer,
		waitgroup.Analyzer}
}

// appendStaticcheckIoChecks adds analyzers from staticcheck.io (which in config.json) to the list.
func appendStaticcheckIoChecks(checks map[string]bool) {
	appendChecks(staticcheck.Analyzers, checks)
	appendChecks(stylecheck.Analyzers, checks)
	appendChecks(simple.Analyzers, checks)
	appendChecks(quickfix.Analyzers, checks)
}

// appendOtherPublicChecks adds additional public analyzers.
func appendOtherPublicChecks() {
	mychecks = append(mychecks, bodyclose.Analyzer)
	mychecks = append(mychecks, errcheck.Analyzer)
}

// appendCustomOsExitCheck adds a custom analyzer to detect os.Exit calls in the main function.
func appendCustomOsExitCheck() {
	mychecks = append(mychecks, OsExitCheckAnalyzer)
}

// OsExitCheckAnalyzer is a custom analyzer that detects calls to os.Exit in the main function.
var OsExitCheckAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for os.Exit() calls",
	Run:  run,
}

// run implements OsExitCheckAnalyzer.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" || strings.Contains(pass.Fset.Position(file.Package).Filename, ".cache") {
			return nil, nil
		}

		for _, decl := range file.Decls {
			f, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if f.Name.Name != "main" {
				continue
			}

			ast.Inspect(f, func(node ast.Node) bool {
				if callExpr, ok := node.(*ast.CallExpr); ok {
					if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "os" && selExpr.Sel.Name == "Exit" {
							pass.Reportf(callExpr.Pos(), "osexitcheck os.Exit cannot be called in main function of main package")
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}

// main initializes the multichecker with the configured analyzers.
func main() {
	appfile, err := os.Executable()
	if err != nil {
		fmt.Printf("error")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		fmt.Printf("error")
	}
	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("error")
	}
	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}
	appendPassesChecks()
	appendStaticcheckIoChecks(checks)
	appendOtherPublicChecks()
	appendCustomOsExitCheck()

	multichecker.Main(
		mychecks...,
	)
}
