package cli

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func (a *app) renderVMs(out io.Writer, items []vmView) error {
	if a.output == "json" {
		return renderJSON(out, items)
	}

	tw := newTabWriter(out)
	fmt.Fprintln(tw, "NAME\tSTATUS\tVCPU\tMEMORY(GiB)\tCLUSTER\tID")
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\t%s\n", item.Name, item.Status, item.VCPU, item.MemoryGiB, item.Cluster, item.ID)
	}
	return tw.Flush()
}

func (a *app) renderVMDescription(out io.Writer, item vmDescribeView) error {
	if a.output == "json" {
		return renderJSON(out, item)
	}

	tw := newTabWriter(out)
	rows := [][2]string{
		{"Name", item.Name},
		{"ID", item.ID},
		{"Status", item.Status},
		{"Cluster", item.Cluster},
		{"Host", item.Host},
		{"Host Management IP", item.HostManagementIP},
		{"Local ID", item.LocalID},
		{"Node IP", item.NodeIP},
		{"Hostname", item.Hostname},
		{"Description", item.Description},
		{"Guest OS", item.GuestOS},
		{"CPU Model", item.CPUModel},
		{"vCPU", fmt.Sprintf("%d", item.VCPU)},
		{"Memory (GiB)", fmt.Sprintf("%d", item.MemoryGiB)},
		{"Firmware", item.Firmware},
		{"HA", formatBool(item.HA)},
		{"Protected", formatBool(item.Protected)},
		{"In Recycle Bin", formatBool(item.InRecycleBin)},
		{"Cloud-Init Supported", formatBool(item.CloudInitSupported)},
		{"VM Tools Status", item.VMToolsStatus},
		{"VM Tools Version", item.VMToolsVersion},
		{"IPs", strings.Join(item.IPs, ", ")},
		{"DNS Servers", strings.Join(item.DNSServers, ", ")},
		{"Disk Count", fmt.Sprintf("%d", item.DiskCount)},
		{"NIC Count", fmt.Sprintf("%d", item.NicCount)},
		{"Snapshot Count", fmt.Sprintf("%d", item.SnapshotCount)},
		{"Provisioned Size (GiB)", fmt.Sprintf("%d", item.ProvisionedSizeGiB)},
		{"Used Size (GiB)", fmt.Sprintf("%d", item.UsedSizeGiB)},
		{"Last Shutdown Time", item.LastShutdownTime},
		{"Created At", item.LocalCreatedAt},
	}
	for _, row := range rows {
		fmt.Fprintf(tw, "%s:\t%s\n", row[0], emptyIfBlank(row[1]))
	}
	return tw.Flush()
}

func (a *app) renderOperation(out io.Writer, op operationResult) error {
	if a.output == "json" {
		return renderJSON(out, op)
	}
	fmt.Fprintf(out, "%s %s: %s\n", op.Action, op.Target, strings.Join(op.Names, ", "))
	return nil
}

func newTabWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
}

func renderJSON(out io.Writer, value any) error {
	enc := jsontext.NewEncoder(out, jsontext.WithIndent("  "))
	return jsonv2.MarshalEncode(enc, value)
}
