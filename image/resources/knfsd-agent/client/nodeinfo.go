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

package client

type NodeInfo struct {
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

func (c *KnfsdAgentClient) NodeInfo() (*NodeInfo, error) {
	var v *NodeInfo
	err := c.get("api/v1/nodeInfo", &v)
	return v, err
}
