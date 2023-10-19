/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"net/http"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-agent/client"
)

var nodeInfo client.NodeInfo

// fetchNodeInfo populates the nodeData type from the GCE Metadata Server
func fetchNodeInfo() error {
	var err error
	n := &nodeInfo

	// Populate Name
	n.Name, err = getMetadataValue("computeMetadata/v1/instance/name", false)
	if err != nil {
		return err
	}

	// Populate Hostname
	n.Hostname, err = getMetadataValue("computeMetadata/v1/instance/hostname", false)
	if err != nil {
		return err
	}

	// Populate Instance IP Address
	n.InterfaceConfig.IPAddress, err = getMetadataValue("computeMetadata/v1/instance/network-interfaces/0/ip", false)
	if err != nil {
		return err
	}

	// Populate the Network URI
	n.InterfaceConfig.NetworkURI, err = getMetadataValue("computeMetadata/v1/instance/network-interfaces/0/network", false)
	if err != nil {
		return err
	}

	// Populate the Network Name
	n.InterfaceConfig.NetworkName = lastAfterDelimiter(n.InterfaceConfig.NetworkURI, "/")

	// Populate the Instance Zone
	n.Zone, err = getMetadataValue("computeMetadata/v1/instance/zone", true)
	if err != nil {
		return err
	}

	// Populate the Machine Type
	n.MachineType, err = getMetadataValue("computeMetadata/v1/instance/machine-type", true)
	if err != nil {
		return err
	}

	// Populate the Image
	n.Image, err = getMetadataValue("computeMetadata/v1/instance/image", false)
	if err != nil {
		return err
	}

	return nil
}

func handleNodeInfo(*http.Request) (*client.NodeInfo, error) {
	return &nodeInfo, nil
}
