package cli

import (
	"strings"

	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
)

type vmView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	VCPU      int32  `json:"vcpu"`
	MemoryGiB int64  `json:"memoryGiB"`
	Cluster   string `json:"cluster"`
}

type vmDescribeView struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Status             string   `json:"status"`
	Cluster            string   `json:"cluster"`
	Host               string   `json:"host"`
	HostManagementIP   string   `json:"hostManagementIP"`
	LocalID            string   `json:"localID"`
	NodeIP             string   `json:"nodeIP"`
	Hostname           string   `json:"hostname"`
	Description        string   `json:"description"`
	GuestOS            string   `json:"guestOS"`
	CPUModel           string   `json:"cpuModel"`
	VCPU               int32    `json:"vcpu"`
	MemoryGiB          int64    `json:"memoryGiB"`
	Firmware           string   `json:"firmware"`
	HA                 bool     `json:"ha"`
	Protected          bool     `json:"protected"`
	InRecycleBin       bool     `json:"inRecycleBin"`
	CloudInitSupported bool     `json:"cloudInitSupported"`
	VMToolsStatus      string   `json:"vmToolsStatus"`
	VMToolsVersion     string   `json:"vmToolsVersion"`
	IPs                []string `json:"ips"`
	DNSServers         []string `json:"dnsServers"`
	DiskCount          int      `json:"diskCount"`
	NicCount           int      `json:"nicCount"`
	SnapshotCount      int      `json:"snapshotCount"`
	ProvisionedSizeGiB int64    `json:"provisionedSizeGiB"`
	UsedSizeGiB        int64    `json:"usedSizeGiB"`
	LastShutdownTime   string   `json:"lastShutdownTime"`
	LocalCreatedAt     string   `json:"localCreatedAt"`
}

type clusterView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type templateView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type folderView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type operationResult struct {
	Action string   `json:"action"`
	Target string   `json:"target"`
	Names  []string `json:"names"`
}

func vmToView(item *models.VM) vmView {
	clusterName := ""
	if item != nil && item.Cluster != nil && item.Cluster.Name != nil {
		clusterName = *item.Cluster.Name
	}
	return vmView{
		ID:        stringValue(nilIfVMNil(item, item.ID)),
		Name:      stringValue(nilIfVMNil(item, item.Name)),
		Status:    string(vmStatusPtr(item)),
		VCPU:      int32Value(nilIfVMNil(item, item.Vcpu)),
		MemoryGiB: int64Value(nilIfVMNil(item, item.Memory)) / gib,
		Cluster:   clusterName,
	}
}

func describeVM(item *models.VM) vmDescribeView {
	hostName := ""
	hostManagementIP := ""
	if item != nil && item.Host != nil {
		hostName = stringValue(item.Host.Name)
		hostManagementIP = stringValue(item.Host.ManagementIP)
	}

	return vmDescribeView{
		ID:                 stringValue(nilIfVMNil(item, item.ID)),
		Name:               stringValue(nilIfVMNil(item, item.Name)),
		Status:             string(vmStatusPtr(item)),
		Cluster:            nestedClusterName(nilIfVMNil(item, item.Cluster)),
		Host:               hostName,
		HostManagementIP:   hostManagementIP,
		LocalID:            stringValue(nilIfVMNil(item, item.LocalID)),
		NodeIP:             stringValue(nilIfVMNil(item, item.NodeIP)),
		Hostname:           stringValue(nilIfVMNil(item, item.Hostname)),
		Description:        stringValue(nilIfVMNil(item, item.Description)),
		GuestOS:            firstNonEmpty(stringValue(nilIfVMNil(item, item.Os)), enumString(item.GuestOsType)),
		CPUModel:           stringValue(nilIfVMNil(item, item.CPUModel)),
		VCPU:               int32Value(nilIfVMNil(item, item.Vcpu)),
		MemoryGiB:          int64Value(nilIfVMNil(item, item.Memory)) / gib,
		Firmware:           enumString(nilIfVMNil(item, item.Firmware)),
		HA:                 boolValue(nilIfVMNil(item, item.Ha)),
		Protected:          boolValue(nilIfVMNil(item, item.Protected)),
		InRecycleBin:       boolValue(nilIfVMNil(item, item.InRecycleBin)),
		CloudInitSupported: boolValue(nilIfVMNil(item, item.CloudInitSupported)),
		VMToolsStatus:      enumString(nilIfVMNil(item, item.VMToolsStatus)),
		VMToolsVersion:     stringValue(nilIfVMNil(item, item.VMToolsVersion)),
		IPs:                splitCSV(stringValue(nilIfVMNil(item, item.Ips))),
		DNSServers:         splitCSV(stringValue(nilIfVMNil(item, item.DNSServers))),
		DiskCount:          len(item.VMDisks),
		NicCount:           len(item.VMNics),
		SnapshotCount:      len(item.Snapshots),
		ProvisionedSizeGiB: int64Value(nilIfVMNil(item, item.ProvisionedSize)) / gib,
		UsedSizeGiB:        int64Value(nilIfVMNil(item, item.UsedSize)) / gib,
		LastShutdownTime:   stringValue(nilIfVMNil(item, item.LastShutdownTime)),
		LocalCreatedAt:     stringValue(nilIfVMNil(item, item.LocalCreatedAt)),
	}
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
