// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/pivotal-cf/om/api"
)

type InstallationAssetImporterService struct {
	ImportStub        func(api.ImportInstallationInput) error
	importMutex       sync.RWMutex
	importArgsForCall []struct {
		arg1 api.ImportInstallationInput
	}
	importReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *InstallationAssetImporterService) Import(arg1 api.ImportInstallationInput) error {
	fake.importMutex.Lock()
	fake.importArgsForCall = append(fake.importArgsForCall, struct {
		arg1 api.ImportInstallationInput
	}{arg1})
	fake.recordInvocation("Import", []interface{}{arg1})
	fake.importMutex.Unlock()
	if fake.ImportStub != nil {
		return fake.ImportStub(arg1)
	} else {
		return fake.importReturns.result1
	}
}

func (fake *InstallationAssetImporterService) ImportCallCount() int {
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	return len(fake.importArgsForCall)
}

func (fake *InstallationAssetImporterService) ImportArgsForCall(i int) api.ImportInstallationInput {
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	return fake.importArgsForCall[i].arg1
}

func (fake *InstallationAssetImporterService) ImportReturns(result1 error) {
	fake.ImportStub = nil
	fake.importReturns = struct {
		result1 error
	}{result1}
}

func (fake *InstallationAssetImporterService) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	return fake.invocations
}

func (fake *InstallationAssetImporterService) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
