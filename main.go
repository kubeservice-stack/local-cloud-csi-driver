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

package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/kubeservice-stack/local-cloud-csi-driver/pkg/local"
	"github.com/kubeservice-stack/local-cloud-csi-driver/pkg/om"
	_ "github.com/kubeservice-stack/local-cloud-csi-driver/pkg/options"
	"github.com/kubeservice-stack/local-cloud-csi-driver/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func init() {
	_ = flag.Set("logtostderr", "true")
}

const (
	// LogfilePrefix prefix of log file
	LogfilePrefix = "/var/log/ecloud/"

	// MBSIZE MB size
	MBSIZE = 1024 * 1024

	// TypePluginSuffix is the suffix of all storage plugins.
	TypePluginSuffix = "plugin.csi.ecloud.cmss.com"

	// PluginServicePort default port is 11260.
	PluginServicePort = "11260"

	// ProvisionerServicePort default port is 11270.
	ProvisionerServicePort = "11270"

	// TypePluginLocal LVM type plugin
	TypePluginLocal = "local.csi.ecloud.cmss.com"

	ExtenderAgent = "agent"
)

var (
	endpoint        = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID          = flag.String("nodeid", "", "node id")
	runAsController = flag.Bool("run-as-controller", false, "Only run as controller service")
	driver          = flag.String("driver", TypePluginLocal, "CSI Driver")
	// Deprecated: rootDir is instead by KUBELET_ROOT_DIR env.
	rootDir = flag.String("rootdir", "/var/lib/kubelet/csi-plugins", "Kubernetes root directory")
)

// CSI Plugin
func main() {
	flag.Parse()
	serviceType := os.Getenv(utils.ServiceType)

	if len(serviceType) == 0 || serviceType == "" {
		serviceType = utils.PluginService
	}

	// When serviceType is neither plugin nor provisioner, the program will exits.
	if serviceType != utils.PluginService && serviceType != utils.ProvisionerService {
		log.Fatalf("Service type is unknown:%s", serviceType)
	}

	var logAttribute string
	switch serviceType {
	case utils.ProvisionerService:
		logAttribute = strings.Replace(TypePluginSuffix, utils.PluginService, utils.ProvisionerService, -1)
	case utils.PluginService:
		logAttribute = TypePluginSuffix
	default:
	}

	setLogAttribute(logAttribute)

	log.Infof("Multi CSI Driver Name: %s, nodeID: %s, endPoints: %s", *driver, *nodeID, *endpoint)

	multiDriverNames := *driver
	endPointName := *endpoint
	driverNames := strings.Split(multiDriverNames, ",")
	var wg sync.WaitGroup

	// Storage devops
	go om.StorageOM()

	for _, driverName := range driverNames {
		wg.Add(1)

		if err := createPersistentStorage(path.Join(utils.KubeletRootDir, "/csi-plugins", driverName, "controller")); err != nil {
			log.Errorf("failed to create persistent storage for controller: %v", err)
			os.Exit(1)
		}
		if err := createPersistentStorage(path.Join(utils.KubeletRootDir, "/csi-plugins", driverName, "node")); err != nil {
			log.Errorf("failed to create persistent storage for node: %v", err)
			os.Exit(1)
		}
		go func(endPoint string) {
			defer wg.Done()
			driver := lvm.NewDriver(*nodeID, endPoint)
			driver.Run()
		}(endPointName)

	}
	servicePort := os.Getenv("SERVICE_PORT")

	if len(servicePort) == 0 || servicePort == "" {
		switch serviceType {
		case utils.PluginService:
			servicePort = PluginServicePort
		case utils.ProvisionerService:
			servicePort = ProvisionerServicePort
		default:
		}
	}

	log.Info("CSI is running status.")
	server := &http.Server{Addr: ":" + servicePort}

	http.HandleFunc("/healthz", healthHandler)
	log.Infof("Metric listening on address: /healthz")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Service port listen and serve err:%s", err.Error())
	}
	wg.Wait()
	os.Exit(0)
}
func createPersistentStorage(persistentStoragePath string) error {
	log.Infof("Create Stroage Path: %s", persistentStoragePath)
	return os.MkdirAll(persistentStoragePath, os.FileMode(0755))
}

// rotate log file by 2M bytes
// default print log to stdout and file both.
func setLogAttribute(driver string) {
	logType := os.Getenv("LOG_TYPE")
	logType = strings.ToLower(logType)
	if logType != "stdout" && logType != "host" {
		logType = "both"
	}
	if logType == "stdout" {
		return
	}

	os.MkdirAll(LogfilePrefix, os.FileMode(0755))
	logFile := LogfilePrefix + driver + ".log"
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		os.Exit(1)
	}

	// rotate the log file if too large
	if fi, err := f.Stat(); err == nil && fi.Size() > 2*MBSIZE {
		f.Close()
		timeStr := time.Now().Format("-2006-01-02-15:04:05")
		timedLogfile := LogfilePrefix + driver + timeStr + ".log"
		err = os.Rename(logFile, timedLogfile)
		if err != nil {
			os.Exit(1)
		}
		f, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			os.Exit(1)
		}
	}
	if logType == "both" {
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
	} else {
		log.SetOutput(f)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	message := "Liveness probe is OK, time:" + time.Now().String()
	_, _ = w.Write([]byte(message))
}
