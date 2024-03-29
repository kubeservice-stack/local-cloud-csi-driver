/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	MacMountedPath     = "/System/Volumes/Data"
	MacUnmountedPath   = "./tmp"
	MacCSIPluginFlag   = "./tmp/csiPluginFlag"
	LinuxMountedPath   = "/sys/fs/cgroup"
	LinuxUnmountedPath = "./tmp/"
)

func TestResultStatus(t *testing.T) {
	testParams := []string{"a", "b", "c"}
	for _, param := range testParams {
		resultSucceed := Succeed(param)
		assert.Equal(t, param, resultSucceed.Message)

		resultNotSupport := NotSupport(param)
		assert.Equal(t, param, resultNotSupport.Message)

		resultFail := Fail(param)
		assert.Equal(t, param, resultFail.Message)
	}

}

func TestCreateDest(t *testing.T) {
	existsPath := "./tmp/test1/"
	_ = os.MkdirAll(existsPath, 0777)
	normalPath := "./tmp/test2"

	err := CreateDest(existsPath)
	assert.Nil(t, err)
	err = CreateDest(normalPath)
	assert.Nil(t, err)

	err = RemoveContents(MacUnmountedPath)
	assert.Nil(t, err)
}

func TestIsMounted(t *testing.T) {
	currentSystem := runtime.GOOS
	if currentSystem == "darwin" {
		assert.True(t, IsMounted(MacMountedPath))
		assert.False(t, IsMounted(MacUnmountedPath))
	} else if currentSystem == "linux" {
		assert.True(t, IsMounted(LinuxMountedPath))
		assert.False(t, IsMounted(LinuxUnmountedPath))
	}
}

func TestWriteJsonFile(t *testing.T) {
	var jsonData struct {
		mountfile string
		runtime   string
	}
	jsonData.mountfile = "tmp/ecloudcsiplugin.json"
	jsonData.runtime = "runv"

	err := os.MkdirAll(MacUnmountedPath, 0777)
	assert.Nil(t, err)
	resultFile := filepath.Join(MacCSIPluginFlag, CsiPluginRunTimeFlagFile)
	assert.Nil(t, WriteJSONFile(jsonData, resultFile))
}

func TestIsMountPointRunv(t *testing.T) {

	assert.False(t, IsMountPointRunv(MacMountedPath))
	err := os.MkdirAll(MacCSIPluginFlag, 0777)
	assert.Nil(t, err)
	WriteTestContent(MacCSIPluginFlag)
	assert.True(t, IsMountPointRunv(MacCSIPluginFlag))

	err = RemoveContents(MacUnmountedPath)
	assert.Nil(t, err)
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer os.Remove(dir)
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteTestContent(path string) string {
	d1 := []byte(`{"mountfile": "tmp/ecloudcsiplugin.json","runtime":"runv"}`)
	_ = os.MkdirAll(path, 0777)
	fileName := fmt.Sprintf("%s/%s", path, CsiPluginRunTimeFlagFile)
	err := os.WriteFile(fileName, d1, 0644)
	if err != nil {
		panic(err)
	}
	return fileName
}

func TestIsHostFileExist(t *testing.T) {
	assert := assert.New(t)
	aa := IsHostFileExist("tmp/ecloudcsiplugin.json")
	assert.True(aa)
}

/*
func TestPodRunTime(t *testing.T) {
	var restConfig *rest.Config
	dummyClient, err := kubernetes.NewForConfig(restConfig)
	req := &csi.NodePublishVolumeRequest{
		VolumeContext: map[string]string{
			"csi.storage.k8s.io/pod.name":      "podName",
			"csi.storage.k8s.io/pod.namespace": "podNamespace",
		},
	}
	_, err := GetPodRunTime(req, dummyClient)
	assert.NotNil(t, err)
}
*/
