package versionparams_test

import (
	"testing"

	"github.com/nickwells/param.mod/v6/paramset"
	"github.com/nickwells/testhelper.mod/v2/testhelper"
	"github.com/nickwells/versionparams.mod/versionparams"
)

func TestAddParams(t *testing.T) {
	testCase := struct {
		testhelper.ID
		testhelper.ExpPanic
	}{
		ID: testhelper.MkID("ensure AddParams doesn't panic"),
	}
	panicked, panicVal := testhelper.PanicSafe(func() {
		_ = paramset.NewNoHelpNoExitNoErrRptOrPanic(versionparams.AddParams)
	})
	testhelper.CheckExpPanic(t, panicked, panicVal, testCase)
}
