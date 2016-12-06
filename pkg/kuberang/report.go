package kuberang

import (
	"io"

	"github.com/apprenda/kuberang/pkg/util"
)

type report interface {
	addSuccess(string)
	addIgnored(string)
	addError(string)
	isSuccess() bool
}

// CheckResult contains the results from executing a check in Kuberang
type CheckResult struct {
	Name    string
	Success bool
	Ignored bool
}

type simpleReport struct {
	CheckResults []CheckResult
}

func (r *simpleReport) isSuccess() bool {
	for _, c := range r.CheckResults {
		if !c.Success && !c.Ignored { // return false if a check failed and is not ignored
			return false
		}
	}
	return true
}

func (r *simpleReport) addSuccess(name string) {
	r.CheckResults = append(r.CheckResults, CheckResult{Name: name, Success: true})
}

func (r *simpleReport) addError(name string) {
	r.CheckResults = append(r.CheckResults, CheckResult{Name: name, Success: true})
}

func (r *simpleReport) addIgnored(name string) {
	r.CheckResults = append(r.CheckResults, CheckResult{Name: name, Success: false, Ignored: true})
}

type echoReport struct {
	report simpleReport
	out    io.Writer
}

func (r *echoReport) addSuccess(name string) {
	r.report.addSuccess(name)
	util.PrettyPrintOk(r.out, name)
}

func (r *echoReport) addError(name string) {
	r.report.addError(name)
	util.PrettyPrintErr(r.out, name)
}

func (r *echoReport) addIgnored(name string) {
	r.report.addIgnored(name)
	util.PrettyPrintErrorIgnored(r.out, name)
}

func (r *echoReport) isSuccess() bool {
	return r.report.isSuccess()
}
