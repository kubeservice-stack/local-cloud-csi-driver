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
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
)

func TestGetPodUIDFromTargetPath(t *testing.T) {
	assert := assert.New(t)
	poduid := GetPodUIDFromTargetPath("/var/lib/kubelet/pods/c06d5521-3d9c-4517-bdc2-e6df34b9e8f1/volumes/kubernetes.io~csi/lvm-9e30e658-5f85-4ec6-ada2-c4ff308b506e/mount")
	assert.Equal(poduid, "c06d5521-3d9c-4517-bdc2-e6df34b9e8f1")

	poduide := GetPodUIDFromTargetPath("")
	assert.Equal(poduide, "")
}

func TestGetBpsLimt(t *testing.T) {
	assert := assert.New(t)
	bps, err := getBpsLimt("5K")
	assert.Nil(err)
	assert.Equal(bps, 5*1024)

	bps, err = getBpsLimt("5M")
	assert.Nil(err)
	assert.Equal(bps, 5*1024*1024)

	bps, err = getBpsLimt("5G")
	assert.Nil(err)
	assert.Equal(bps, 5*1024*1024*1024)

	bps, err = getBpsLimt("10")
	assert.Nil(err)
	assert.Equal(bps, 10)

	bps, err = getBpsLimt("aa")
	assert.NotNil(err)
	assert.Equal(bps, 0)
}

func TestSetVolumeIOLimit(t *testing.T) {
	assert := assert.New(t)
	err := SetVolumeIOLimit("", &csi.NodePublishVolumeRequest{})
	assert.Nil(err)
}
