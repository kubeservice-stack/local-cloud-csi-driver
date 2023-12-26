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

package om

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kubeservice-stack/local-cloud-csi-driver/pkg/utils"
	log "github.com/sirupsen/logrus"
	k8smount "k8s.io/utils/mount"
)

var (
	// FixedPodList fix pod
	FixedPodList = map[string]string{}
	// K8sMounter mounter
	K8sMounter = k8smount.New("")
)

// FixOrphanedPodIssue Pod Like:
// Jul 10 18:55:49 kubelet: E0710 18:55:49.251132    7643 kubelet_volumes.go:154] orphaned pod "a60244b2-e6ee-4a63-b311-13f7b29ef49a"
// found, but volume paths are still present on disk : There were a total of 1 errors similar to this. Turn up verbosity to see them.
func FixOrphanedPodIssue(line string) bool {
	splitStr := strings.Split(line, "rphaned pod")
	if len(splitStr) < 2 {
		log.Warnf("OrphanPod: Error orphaned line format: %s", line)
		return false
	}
	partStr := strings.Split(splitStr[1], "\"")
	if len(partStr) < 2 {
		log.Warnf("OrphanPod: Error line format: %s", line)
		return false
	}
	orphanUID := partStr[1]
	if len(strings.Split(orphanUID, "-")) != 5 {
		log.Warnf("OrphanPod: Error Pod Uid format: %s, %s", orphanUID, line)
		return false
	}

	// break fixed orphaned pod
	if value, ok := FixedPodList[orphanUID]; ok && value == "fixed" {
		return true
	}

	// check kubernetes csi volumes
	csiPodPath := filepath.Join(utils.KubeletRootDir, "/pods", orphanUID, "volumes/kubernetes.io~csi")
	volumes, err := os.ReadDir(csiPodPath)
	if err != nil {
		log.Warnf("OrphanPod: List Volumes with error: %s, line: %s", err.Error(), line)
		return false
	}
	if len(volumes) == 0 {
		FixedPodList[orphanUID] = "fixed"
	}
	for _, volume := range volumes {
		volumePath := filepath.Join(csiPodPath, volume.Name())
		volumeMountPath := filepath.Join(volumePath, "mount")
		volumeJSONPath := filepath.Join(volumePath, "vol_data.json")
		if utils.IsFileExisting(volumeMountPath) {
			if err := k8smount.CleanupMountPoint(volumeMountPath, K8sMounter, false); err != nil {
				log.Errorf("OrphanPod: CleanupMountPoint %s, with Error: %s, Log: %s", volumeMountPath, err.Error(), line)
				continue
			} else {
				log.Infof("OrphanPod: Successful Remove Path(%s).", volumeMountPath)
			}
		}
		if IsFileExisting(volumeJSONPath) {
			err = os.Remove(volumeJSONPath)
			if err != nil {
				log.Errorf("OrphanPod: Remove Json File %s with error %s", volumeJSONPath, err.Error())
			} else {
				log.Infof("OrphanPod: Remove Json File %s Successful", volumeJSONPath)
			}
		}
		if empty, _ := utils.IsDirEmpty(volumePath); empty {
			err = os.Remove(volumePath)
			if err != nil {
				log.Errorf("OrphanPod: Remove Volume Path %s with error %s", volumePath, err.Error())
			} else {
				log.Infof("OrphanPod: Remove Volume Path %s Successful", volumePath)
			}
		}
	}
	return true
}
