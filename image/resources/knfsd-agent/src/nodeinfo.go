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
	"encoding/json"
	"net/http"
)

// nodeData is returned as a response to /api/v1.0/nodeinfo
type nodeData struct {
	Name            string `json:"name"`
	Hostname        string `json:"hostname"`
	InterfaceConfig struct {
		IPAddress   string `json:"ipAddress"`
		NetworkName string `json:"networkName"`
		NetworkURI  string `json:"networkURI"`
	} `json:"interfaceConfig"`
	Zone        string `json:"zone"`
	MachineType string `json:"machineType"`
	Image       string `json:"image"`
}

// fetch populates the nodeData type from the GCE Metadata Server
func (n *nodeData) fetch() error {

	var err error

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
	n.InterfaceConfig.NetworkName, err = lastAfterDelimiter(n.InterfaceConfig.NetworkURI, "/")
	if err != nil {
		return err
	}

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

// handleProcess handles requests sent to the /api/vx.x/nodeinfo endpoint
func handleProcess(w http.ResponseWriter, r *http.Request) (int, error) {

	// Encode JSON
	resp, err := json.MarshalIndent(&nodeInfo, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return http.StatusInternalServerError, err
	}

	// Write response
	w.WriteHeader(http.StatusOK)
	w.Write(resp)

	// Return status code and error
	return http.StatusOK, nil

}
