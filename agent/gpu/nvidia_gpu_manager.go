// +build linux

// Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package gpu

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/pkg/errors"
)

// GPUManager encompasses methods to get information on GPUs and their driver
type GPUManager interface {
	Initialize() error
	SetGPUIDs([]string)
	GetGPUIDs() []string
	SetDriverVersion(string)
	GetDriverVersion() string
	SetRuntimeVersion(string)
	GetRuntimeVersion() string
}

// NvidiaGPUManager is used as a wrapper for NVML APIs and implements GPUManager
// interface
type NvidiaGPUManager struct {
	DriverVersion       string   `json:"DriverVersion"`
	NvidiaDockerVersion string   `json:"NvidiaDockerVersion"`
	GPUIDs              []string `json:"GPUIDs"`
	lock                sync.RWMutex
}

const (
	// GPUInfoDirPath is the directory where gpus and driver info are saved
	GPUInfoDirPath = "/var/lib/ecs/gpu"
	// NvidiaGPUInfoFilePath is the file path where gpus and driver info are saved
	NvidiaGPUInfoFilePath = GPUInfoDirPath + "/nvidia-gpu-info.json"
)

// NewNvidiaGPUManager is used to obtain NvidiaGPUManager handle
func NewNvidiaGPUManager() GPUManager {
	return &NvidiaGPUManager{}
}

// Initialize sets the fields of Nvidia GPU Manager struct
func (n *NvidiaGPUManager) Initialize() error {
	if GPUInfoFileExists() {
		// GPU info file found
		gpuJSON, err := GetGPUInfoJSON()
		if err != nil {
			return errors.Wrapf(err, "could not read GPU file content")
		}
		var nvidiaGPUInfo NvidiaGPUManager
		err = json.Unmarshal(gpuJSON, &nvidiaGPUInfo)
		if err != nil {
			return errors.Wrapf(err, "could not unmarshal GPU file content")
		}
		n.SetDriverVersion(nvidiaGPUInfo.GetDriverVersion())
		n.SetGPUIDs(nvidiaGPUInfo.GetGPUIDs())
		n.SetRuntimeVersion(nvidiaGPUInfo.GetRuntimeVersion())
	}
	return nil
}

var GPUInfoFileExists = CheckForGPUInfoFile

func CheckForGPUInfoFile() bool {
	_, err := os.Stat(NvidiaGPUInfoFilePath)
	return !os.IsNotExist(err)
}

var GetGPUInfoJSON = GetGPUInfo

func GetGPUInfo() ([]byte, error) {
	gpuInfo, err := os.Open(NvidiaGPUInfoFilePath)
	if err != nil {
		return nil, err
	}
	defer gpuInfo.Close()

	gpuJSON, err := ioutil.ReadAll(gpuInfo)
	if err != nil {
		return nil, err
	}
	return gpuJSON, nil
}

// SetGPUIDs sets the GPUIDs
func (n *NvidiaGPUManager) SetGPUIDs(gpuIDs []string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.GPUIDs = gpuIDs
}

// GetGPUIDs returns the GPUIDs
func (n *NvidiaGPUManager) GetGPUIDs() []string {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.GPUIDs
}

// SetDriverVersion is a setter for nvidia driver version
func (n *NvidiaGPUManager) SetDriverVersion(version string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.DriverVersion = version
}

// GetDriverVersion is a getter for nvidia driver version
func (n *NvidiaGPUManager) GetDriverVersion() string {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.DriverVersion
}

// SetRuntimeVersion is a setter for nvidia docker version
func (n *NvidiaGPUManager) SetRuntimeVersion(version string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.NvidiaDockerVersion = version
}

// GetRuntimeVersion is a getter for nvidia docker version
func (n *NvidiaGPUManager) GetRuntimeVersion() string {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.NvidiaDockerVersion
}
