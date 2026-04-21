package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
	"github.com/spf13/cobra"
)

const gib int64 = 1024 * 1024 * 1024

func addVMSelectionFlags(cmd *cobra.Command, opts *vmQueryOptions, help string) {
	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", help)
	cmd.Flags().BoolVarP(&opts.All, "all", "A", false, "Select all VMs")
}

func resolveNameArg(args []string, flagValue string) (string, error) {
	if len(args) == 0 {
		return flagValue, nil
	}
	if flagValue != "" && args[0] != flagValue {
		return "", errors.New("positional name and --name refer to different resources")
	}
	return args[0], nil
}

func defaultStringSlice(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}

func redactSecret(value string) string {
	if value == "" {
		return ""
	}
	return strings.Repeat("*", 8)
}

func parseVMStatus(v string) (models.VMStatus, error) {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "RUNNING":
		return models.VMStatusRUNNING, nil
	case "STOPPED":
		return models.VMStatusSTOPPED, nil
	default:
		return "", fmt.Errorf("invalid status %q, must be STOPPED or RUNNING", v)
	}
}

func parseFirmware(v string) (models.VMFirmware, error) {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "BIOS":
		return models.VMFirmwareBIOS, nil
	case "UEFI":
		return models.VMFirmwareUEFI, nil
	default:
		return "", fmt.Errorf("invalid firmware %q, must be BIOS or UEFI", v)
	}
}

func parseBus(v string) (models.Bus, error) {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "VIRTIO":
		return models.BusVIRTIO, nil
	case "SCSI":
		return models.BusSCSI, nil
	default:
		return "", fmt.Errorf("invalid bus %q, must be VIRTIO or SCSI", v)
	}
}

func nilIfVMNil[T any](item *models.VM, field *T) *T {
	if item == nil {
		return nil
	}
	return field
}

func vmStatusPtr(item *models.VM) models.VMStatus {
	if item == nil || item.Status == nil {
		return models.VMStatus("")
	}
	return *item.Status
}

func vmIDs(items []*models.VM) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil && item.ID != nil {
			ids = append(ids, *item.ID)
		}
	}
	return ids
}

func vmNames(items []*models.VM) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil && item.Name != nil {
			names = append(names, *item.Name)
		}
	}
	return names
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func int32Value(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func int64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func boolValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func nestedClusterName(v *models.NestedCluster) string {
	if v == nil {
		return ""
	}
	return stringValue(v.Name)
}

func enumString[T ~string](v *T) string {
	if v == nil {
		return ""
	}
	return string(*v)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func formatBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func emptyIfBlank(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}
