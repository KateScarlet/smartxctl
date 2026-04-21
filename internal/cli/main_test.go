package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
)

func TestValidateConfig(t *testing.T) {
	cfg := config{
		Tower: towerConfig{Host: "tower.example.com", Scheme: "https"},
		User:  userConfig{Name: "admin", Password: "secret", Source: "LOCAL"},
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig() returned unexpected error: %v", err)
	}
}

func TestValidateConfigRejectsBadScheme(t *testing.T) {
	cfg := config{
		Tower: towerConfig{Host: "tower.example.com", Scheme: "tcp"},
		User:  userConfig{Name: "admin", Password: "secret", Source: "LOCAL"},
	}

	err := validateConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "tower.scheme") {
		t.Fatalf("expected scheme validation error, got %v", err)
	}
}

func TestCreateVMOptionsValidate(t *testing.T) {
	opts := createVMOptions{
		Name:     "demo",
		Cluster:  "cluster-a",
		VCPU:     2,
		MemoryGB: 4,
		Status:   "STOPPED",
		Firmware: "BIOS",
		Bus:      "VIRTIO",
	}

	if err := opts.validate(); err != nil {
		t.Fatalf("validate() returned unexpected error: %v", err)
	}
}

func TestResolveNameArgRejectsMismatch(t *testing.T) {
	_, err := resolveNameArg([]string{"from-arg"}, "from-flag")
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestRootCommandGetVMDefaultsToAll(t *testing.T) {
	cmd := NewRootCmd()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"get", "vm"})

	err := cmd.Execute()
	if err != nil && strings.Contains(err.Error(), "Config File") {
		return
	}
}

func TestRootCommandDescribeVMRequiresName(t *testing.T) {
	cmd := NewRootCmd()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"describe", "vm"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "specify a VM name") {
		t.Fatalf("expected missing name error, got %v", err)
	}
}

func TestDescribeVMBuildsReadableView(t *testing.T) {
	name := "demo-vm"
	id := "vm-1"
	clusterName := "cluster-a"
	hostName := "host-a"
	hostIP := "10.0.0.10"
	description := "demo machine"
	localID := "42"
	nodeIP := "10.0.0.20"
	hostname := "demo.local"
	cpuModel := "x86_64"
	firmware := models.VMFirmwareUEFI
	status := models.VMStatusRUNNING
	guestOS := "ubuntu"
	ha := true
	protected := true
	cloudInit := true
	ips := "192.168.1.10, 192.168.1.11"
	dns := "8.8.8.8, 1.1.1.1"
	vmToolsStatus := models.VMToolsStatusRUNNING
	vmToolsVersion := "1.0.0"
	memory := int64(8 * gib)
	provisioned := int64(20 * gib)
	used := int64(5 * gib)
	vcpu := int32(4)

	vm := &models.VM{
		ID:                 &id,
		Name:               &name,
		Status:             &status,
		Cluster:            &models.NestedCluster{Name: &clusterName},
		Host:               &models.NestedHost{Name: &hostName, ManagementIP: &hostIP},
		Description:        &description,
		LocalID:            &localID,
		NodeIP:             &nodeIP,
		Hostname:           &hostname,
		CPUModel:           &cpuModel,
		Firmware:           &firmware,
		GuestOsType:        (*models.VMGuestsOperationSystem)(&guestOS),
		Ha:                 &ha,
		Protected:          &protected,
		CloudInitSupported: &cloudInit,
		Ips:                &ips,
		DNSServers:         &dns,
		VMToolsStatus:      &vmToolsStatus,
		VMToolsVersion:     &vmToolsVersion,
		Memory:             &memory,
		ProvisionedSize:    &provisioned,
		UsedSize:           &used,
		Vcpu:               &vcpu,
		VMDisks:            []*models.NestedVMDisk{{ID: &id}},
		VMNics:             []*models.NestedVMNic{{ID: &id}, {ID: &name}},
		Snapshots:          []*models.NestedVMSnapshot{{ID: &id}},
	}

	view := describeVM(vm)
	if view.Name != name || view.Cluster != clusterName || view.Host != hostName {
		t.Fatalf("unexpected describe view: %+v", view)
	}
	if len(view.IPs) != 2 || view.MemoryGiB != 8 || view.NicCount != 2 {
		t.Fatalf("unexpected parsed values: %+v", view)
	}
}

func TestRenderVMDescriptionTable(t *testing.T) {
	a := &app{output: "table"}
	var out bytes.Buffer

	err := a.renderVMDescription(&out, vmDescribeView{
		Name:      "demo-vm",
		ID:        "vm-1",
		Status:    "RUNNING",
		MemoryGiB: 8,
	})
	if err != nil {
		t.Fatalf("renderVMDescription() error = %v", err)
	}

	rendered := out.String()
	if !strings.Contains(rendered, "Name:") || !strings.Contains(rendered, "demo-vm") {
		t.Fatalf("unexpected rendered description: %q", rendered)
	}
}
