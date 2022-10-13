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
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ReadFileLinesFromHost read file from /var/log/messages
func ReadFileLinesFromHost(fname string) []string {
	lineList := []string{}
	out := ""
	var err error
	cmd := fmt.Sprintf("tail -n %d %s", GlobalConfigVar.MessageFileTailLines, fname)
	if out, err = Run(cmd); err != nil {
		return lineList
	}
	return strings.Split(out, "\n")
}

// Run run shell command
func Run(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to run cmd: " + cmd + ", with out: " + string(out) + ", with error: " + err.Error())
	}
	return string(out), nil
}

// IsFileExisting check file exist in volume driver;
func IsFileExisting(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
