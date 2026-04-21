package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	apiclient "github.com/smartxworks/cloudtower-go-sdk/v2/client"
	"github.com/smartxworks/cloudtower-go-sdk/v2/client/cluster"
	"github.com/smartxworks/cloudtower-go-sdk/v2/client/content_library_vm_template"
	"github.com/smartxworks/cloudtower-go-sdk/v2/client/vm"
	"github.com/smartxworks/cloudtower-go-sdk/v2/client/vm_folder"
	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
	"github.com/smartxworks/cloudtower-go-sdk/v2/utils"
)

func (a *app) runGetVM(name string, cluster string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}

	vms, err := getVMs(client, name, cluster)
	if err != nil {
		return err
	}

	views := make([]vmView, 0, len(vms))
	for _, item := range vms {
		views = append(views, vmToView(item))
	}
	return a.renderVMs(out, views)
}

func (a *app) runDescribeVM(name string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}

	vms, err := getVMs(client, name, "")
	if err != nil {
		return err
	}
	if len(vms) == 0 {
		return fmt.Errorf("VM not found: %s", name)
	}

	return a.renderVMDescription(out, describeVM(vms[0]))
}

func (a *app) runDescribeFolder(name string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}

	items, err := getFolders(client, name)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("folder not found: %s", name)
	}

	if a.output == "json" {
		return renderJSON(out, folderView{
			ID:   stringValue(items[0].ID),
			Name: stringValue(items[0].Name),
		})
	}

	tw := newTabWriter(out)
	fmt.Fprintln(tw, "NAME\tID")
	fmt.Fprintf(tw, "%s\t%s\n", stringValue(items[0].Name), stringValue(items[0].ID))
	return tw.Flush()
}

func (a *app) runGetClusters(name string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}

	items, err := getClusters(client, name)
	if err != nil {
		return err
	}

	views := make([]clusterView, 0, len(items))
	for _, item := range items {
		views = append(views, clusterView{
			ID:   stringValue(item.ID),
			Name: stringValue(item.Name),
		})
	}

	if a.output == "json" {
		return renderJSON(out, views)
	}

	tw := newTabWriter(out)
	fmt.Fprintln(tw, "NAME\tID")
	for _, item := range views {
		fmt.Fprintf(tw, "%s\t%s\n", item.Name, item.ID)
	}
	return tw.Flush()
}

func (a *app) runGetTemplates(name string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}

	items, err := getTemplates(client, name)
	if err != nil {
		return err
	}

	views := make([]templateView, 0, len(items))
	for _, item := range items {
		views = append(views, templateView{
			ID:   stringValue(item.ID),
			Name: stringValue(item.Name),
		})
	}

	if a.output == "json" {
		return renderJSON(out, views)
	}

	tw := newTabWriter(out)
	fmt.Fprintln(tw, "NAME\tID")
	for _, item := range views {
		fmt.Fprintf(tw, "%s\t%s\n", item.Name, item.ID)
	}
	return tw.Flush()
}

func (a *app) runGetFolders(name string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}

	items, err := getFolders(client, name)
	if err != nil {
		return err
	}

	views := make([]folderView, 0, len(items))
	for _, item := range items {
		views = append(views, folderView{
			ID:   stringValue(item.ID),
			Name: stringValue(item.Name),
		})
	}

	if a.output == "json" {
		return renderJSON(out, views)
	}

	tw := newTabWriter(out)
	fmt.Fprintln(tw, "NAME\tID")
	for _, item := range views {
		fmt.Fprintf(tw, "%s\t%s\n", item.Name, item.ID)
	}
	return tw.Flush()
}

func (a *app) runCreateFolder(name string, cluster string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}

	clusterID, err := getClusterID(client, cluster)
	if err != nil {
		return err
	}

	params := vm_folder.NewCreateVMFolderParams()
	params.RequestBody = []*models.VMFolderCreationParams{
		{
			Name:      &name,
			ClusterID: &clusterID,
		},
	}

	resp, err := client.VMFolder.CreateVMFolder(params)
	if err != nil {
		return fmt.Errorf("create folder: %w", err)
	}
	if err := waitForFolderTasks(context.Background(), client, resp.Payload); err != nil {
		return err
	}
	return a.renderOperation(out, operationResult{Action: "create", Target: "folder", Names: []string{name}})
}

func (a *app) runCreateVM(opts createVMOptions, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}
	if opts.Template != "" {
		return a.runCreateVMFromTemplate(client, opts, out)
	}
	return a.runCreateBlankVM(client, opts, out)
}

func (a *app) runCreateBlankVM(client *apiclient.Cloudtower, opts createVMOptions, out io.Writer) error {
	clusterID, err := getClusterID(client, opts.Cluster)
	if err != nil {
		return err
	}

	vmStatus, err := parseVMStatus(opts.Status)
	if err != nil {
		return err
	}
	firmware, err := parseFirmware(opts.Firmware)
	if err != nil {
		return err
	}
	bus, err := parseBus(opts.Bus)
	if err != nil {
		return err
	}

	params := vm.NewCreateVMParams()
	memory := int64(opts.MemoryGB) * gib
	cpuCores := int32(opts.VCPU)
	cpuSockets := int32(1)

	body := &models.VMCreationParams{
		ClusterID:  &clusterID,
		Name:       &opts.Name,
		Ha:         &opts.HA,
		Memory:     &memory,
		CPUCores:   &cpuCores,
		CPUSockets: &cpuSockets,
		Status:     vmStatus.Pointer(),
		Firmware:   firmware.Pointer(),
		VMNics:     []*models.VMNicParams{},
	}

	if opts.ConnectVlanID != "" {
		body.VMNics = []*models.VMNicParams{{ConnectVlanID: &opts.ConnectVlanID}}
	}
	if opts.VMVolumeID != "" || opts.ElfImageID != "" {
		body.VMDisks = &models.VMDiskParams{}
	}
	if opts.VMVolumeID != "" {
		boot := int32(0)
		index := int32(0)
		body.VMDisks.MountDisks = []*models.MountDisksParams{{
			Boot:       &boot,
			Bus:        bus.Pointer(),
			Index:      &index,
			VMVolumeID: &opts.VMVolumeID,
		}}
	}
	if opts.ElfImageID != "" {
		boot := int32(0)
		index := int32(0)
		body.VMDisks.MountCdRoms = []*models.VMCdRomParams{{
			Boot:       &boot,
			Index:      &index,
			ElfImageID: &opts.ElfImageID,
		}}
	}
	if opts.DiskSizeGB > 0 {
		diskSize := int64(opts.DiskSizeGB) * gib
		boot := int32(0)
		index := int32(0)
		newDisk := &models.MountNewCreateDisksParams{
			Boot: &boot,
			Bus:  bus.Pointer(),
			VMVolume: &models.MountNewCreateDisksParamsVMVolume{
				Name:             &opts.DiskName,
				Size:             &diskSize,
				ElfStoragePolicy: models.VMVolumeElfStoragePolicyTypeREPLICA2THINPROVISION.Pointer(),
			},
			Index: &index,
		}
		if body.VMDisks == nil {
			body.VMDisks = &models.VMDiskParams{}
		}
		body.VMDisks.MountNewCreateDisks = append(body.VMDisks.MountNewCreateDisks, newDisk)
	}
	if len(body.VMNics) == 0 && body.VMDisks == nil {
		boot := int32(0)
		index := int32(0)
		body.VMDisks = &models.VMDiskParams{
			MountCdRoms: []*models.VMCdRomParams{{Boot: &boot, Index: &index}},
		}
	}

	params.RequestBody = []*models.VMCreationParams{body}
	resp, err := client.VM.CreateVM(params)
	if err != nil {
		return fmt.Errorf("create VM: %w", err)
	}
	if err := waitForVMTasks(context.Background(), client, resp.Payload); err != nil {
		return err
	}
	return a.renderOperation(out, operationResult{Action: "create", Target: "vm", Names: []string{opts.Name}})
}

func (a *app) runCreateVMFromTemplate(client *apiclient.Cloudtower, opts createVMOptions, out io.Writer) error {
	clusterID, err := getClusterID(client, opts.Cluster)
	if err != nil {
		return err
	}

	templates, err := getTemplates(client, opts.Template)
	if err != nil {
		return err
	}
	if len(templates) == 0 {
		return fmt.Errorf("template not found: %s", opts.Template)
	}

	vmParams := &models.VMCreateVMFromContentLibraryTemplateParams{
		TemplateID: templates[0].ID,
		ClusterID:  &clusterID,
		Name:       &opts.Name,
		IsFullCopy: &opts.IsFullCopy,
	}

	if opts.VCPU > 0 {
		cpuCores := int32(opts.VCPU)
		cpuSockets := int32(1)
		vmParams.CPUCores = &cpuCores
		vmParams.CPUSockets = &cpuSockets
	}
	if opts.MemoryGB > 0 {
		memory := int64(opts.MemoryGB) * gib
		vmParams.Memory = &memory
	}

	if opts.Password != "" || opts.Hostname != "" || opts.UserData != "" || len(opts.Nameservers) > 0 || len(opts.PublicKeys) > 0 || opts.IPAddress != "" || opts.Netmask != "" {
		nicIndex := int32(opts.NicIndex)
		vmParams.CloudInit = &models.TemplateCloudInit{}
		if opts.Password != "" {
			vmParams.CloudInit.DefaultUserPassword = &opts.Password
		}
		if opts.Hostname != "" {
			vmParams.CloudInit.Hostname = &opts.Hostname
		}
		if opts.UserData != "" {
			vmParams.CloudInit.UserData = &opts.UserData
		}
		if len(opts.Nameservers) > 0 {
			vmParams.CloudInit.Nameservers = defaultStringSlice(opts.Nameservers)
		}
		if len(opts.PublicKeys) > 0 {
			vmParams.CloudInit.PublicKeys = defaultStringSlice(opts.PublicKeys)
		}
		if opts.IPAddress != "" || opts.Netmask != "" {
			vmParams.CloudInit.Networks = []*models.CloudInitNetWork{{
				IPAddress: &opts.IPAddress,
				Netmask:   &opts.Netmask,
				NicIndex:  &nicIndex,
				Type:      models.CloudInitNetworkTypeEnumIPV4.Pointer(),
			}}
		}
	}

	params := vm.NewCreateVMFromContentLibraryTemplateParams()
	params.RequestBody = []*models.VMCreateVMFromContentLibraryTemplateParams{vmParams}

	resp, err := client.VM.CreateVMFromContentLibraryTemplate(params)
	if err != nil {
		return fmt.Errorf("create VM from template: %w", err)
	}
	if err := waitForVMTasks(context.Background(), client, resp.Payload); err != nil {
		return err
	}
	return a.renderOperation(out, operationResult{Action: "create", Target: "vm", Names: []string{opts.Name}})
}

func (a *app) runDeleteVM(name string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}
	vms, err := getVMs(client, name, "")
	if err != nil {
		return err
	}
	if len(vms) == 0 {
		return fmt.Errorf("VM not found: %s", name)
	}

	params := vm.NewDeleteVMParams()
	params.RequestBody = &models.VMDeleteParams{Where: &models.VMWhereInput{IDIn: vmIDs(vms)}}
	resp, err := client.VM.DeleteVM(params)
	if err != nil {
		return fmt.Errorf("delete VM: %w", err)
	}
	if err := waitForDeleteVMTasks(context.Background(), client, resp.Payload); err != nil {
		return err
	}
	return a.renderOperation(out, operationResult{Action: "delete", Target: "vm", Names: vmNames(vms)})
}

func (a *app) runDeleteFolder(name string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}

	folders, err := getFolders(client, name)
	if err != nil {
		return err
	}
	if len(folders) == 0 {
		return fmt.Errorf("folder not found: %s", name)
	}

	params := vm_folder.NewDeleteVMFolderParams()
	params.RequestBody = &models.VMFolderDeletionParams{Where: &models.VMFolderWhereInput{IDIn: []string{*folders[0].ID}}}

	resp, err := client.VMFolder.DeleteVMFolder(params)
	if err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}
	if err := waitForDeleteFolderTasks(context.Background(), client, resp.Payload); err != nil {
		return err
	}
	return a.renderOperation(out, operationResult{Action: "delete", Target: "folder", Names: []string{name}})
}

func (a *app) runStartVM(name string, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}
	vms, err := getVMs(client, name, "")
	if err != nil {
		return err
	}
	if len(vms) == 0 {
		return fmt.Errorf("VM not found: %s", name)
	}

	params := vm.NewStartVMParams()
	params.RequestBody = &models.VMStartParams{Where: &models.VMWhereInput{IDIn: vmIDs(vms)}}
	resp, err := client.VM.StartVM(params)
	if err != nil {
		return fmt.Errorf("start VM: %w", err)
	}
	if err := waitForVMTasks(context.Background(), client, resp.Payload); err != nil {
		return err
	}
	return a.renderOperation(out, operationResult{Action: "start", Target: "vm", Names: vmNames(vms)})
}

func (a *app) runStopVM(name string, force bool, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}
	vms, err := getVMs(client, name, "")
	if err != nil {
		return err
	}
	if len(vms) == 0 {
		return fmt.Errorf("VM not found: %s", name)
	}

	var resp interface{ GetPayload() []*models.WithTaskVM }
	if force {
		params := vm.NewPoweroffVMParams()
		params.RequestBody = &models.VMOperateParams{Where: &models.VMWhereInput{IDIn: vmIDs(vms)}}
		resp, err = client.VM.PoweroffVM(params)
	} else {
		params := vm.NewShutDownVMParams()
		params.RequestBody = &models.VMOperateParams{Where: &models.VMWhereInput{IDIn: vmIDs(vms)}}
		resp, err = client.VM.ShutDownVM(params)
	}
	if err != nil {
		return fmt.Errorf("stop VM: %w", err)
	}
	if err := waitForVMTasks(context.Background(), client, resp.GetPayload()); err != nil {
		return err
	}
	return a.renderOperation(out, operationResult{Action: "stop", Target: "vm", Names: vmNames(vms)})
}

func (a *app) runRestartVM(name string, force bool, out io.Writer) error {
	client, err := a.getClient()
	if err != nil {
		return err
	}
	vms, err := getVMs(client, name, "")
	if err != nil {
		return err
	}
	if len(vms) == 0 {
		return fmt.Errorf("VM not found: %s", name)
	}

	var resp interface{ GetPayload() []*models.WithTaskVM }
	if force {
		params := vm.NewForceRestartVMParams()
		params.RequestBody = &models.VMOperateParams{Where: &models.VMWhereInput{IDIn: vmIDs(vms)}}
		resp, err = client.VM.ForceRestartVM(params)
	} else {
		params := vm.NewRestartVMParams()
		params.RequestBody = &models.VMOperateParams{Where: &models.VMWhereInput{IDIn: vmIDs(vms)}}
		resp, err = client.VM.RestartVM(params)
	}
	if err != nil {
		return fmt.Errorf("restart VM: %w", err)
	}
	if err := waitForVMTasks(context.Background(), client, resp.GetPayload()); err != nil {
		return err
	}
	return a.renderOperation(out, operationResult{Action: "restart", Target: "vm", Names: vmNames(vms)})
}

func (o createVMOptions) validate() error {
	if strings.TrimSpace(o.Name) == "" {
		return errors.New("VM name is required")
	}
	if strings.TrimSpace(o.Cluster) == "" {
		return errors.New("cluster name is required")
	}
	if o.VCPU <= 0 || o.MemoryGB <= 0 {
		return errors.New("vcpu and memory must be positive")
	}
	if o.DiskSizeGB > 0 && strings.TrimSpace(o.DiskName) == "" {
		return errors.New("disk-name is required when disk-size is set")
	}
	if _, err := parseVMStatus(o.Status); err != nil {
		return err
	}
	if _, err := parseFirmware(o.Firmware); err != nil {
		return err
	}
	if _, err := parseBus(o.Bus); err != nil {
		return err
	}
	return nil
}

func getVMs(client *apiclient.Cloudtower, name string, cluster string) ([]*models.VM, error) {
	params := vm.NewGetVmsParams()
	where := &models.VMWhereInput{}
	if name != "" {
		where.Name = &name
	}
	if cluster != "" {
		where.Cluster = &models.ClusterWhereInput{Name: &cluster}
	}
	params.RequestBody = &models.GetVmsRequestBody{Where: where}

	resp, err := client.VM.GetVms(params)
	if err != nil {
		return nil, fmt.Errorf("get VMs: %w", err)
	}
	if len(resp.Payload) == 0 {
		return []*models.VM{}, nil
	}
	return resp.Payload, nil
}

func getClusters(client *apiclient.Cloudtower, name string) ([]*models.Cluster, error) {
	params := cluster.NewGetClustersParams()
	where := &models.ClusterWhereInput{}
	if name != "" {
		where.Name = &name
	}
	params.RequestBody = &models.GetClustersRequestBody{Where: where}

	resp, err := client.Cluster.GetClusters(params)
	if err != nil {
		return nil, fmt.Errorf("get clusters: %w", err)
	}
	return resp.Payload, nil
}

func getTemplates(client *apiclient.Cloudtower, name string) ([]*models.ContentLibraryVMTemplate, error) {
	params := content_library_vm_template.NewGetContentLibraryVMTemplatesParams()
	where := &models.ContentLibraryVMTemplateWhereInput{}
	if name != "" {
		where.Name = &name
	}
	params.RequestBody = &models.GetContentLibraryVMTemplatesRequestBody{Where: where}

	resp, err := client.ContentLibraryVMTemplate.GetContentLibraryVMTemplates(params)
	if err != nil {
		return nil, fmt.Errorf("get templates: %w", err)
	}
	return resp.Payload, nil
}

func getClusterID(client *apiclient.Cloudtower, name string) (string, error) {
	items, err := getClusters(client, name)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", fmt.Errorf("cluster not found: %s", name)
	}
	return stringValue(items[0].ID), nil
}

func getFolders(client *apiclient.Cloudtower, name string) ([]*models.VMFolder, error) {
	params := vm_folder.NewGetVMFoldersParams()
	where := &models.VMFolderWhereInput{}
	if name != "" {
		where.Name = &name
	}
	params.RequestBody = &models.GetVMFoldersRequestBody{Where: where}

	resp, err := client.VMFolder.GetVMFolders(params)
	if err != nil {
		return nil, fmt.Errorf("get folders: %w", err)
	}
	return resp.Payload, nil
}

func waitForVMTasks(ctx context.Context, client *apiclient.Cloudtower, items []*models.WithTaskVM) error {
	taskIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil && item.TaskID != nil && *item.TaskID != "" {
			taskIDs = append(taskIDs, *item.TaskID)
		}
	}
	if len(taskIDs) == 0 {
		return nil
	}
	if err := utils.WaitTasks(ctx, client, taskIDs, time.Second); err != nil {
		return fmt.Errorf("wait for tasks: %w", err)
	}
	return nil
}

func waitForFolderTasks(ctx context.Context, client *apiclient.Cloudtower, items []*models.WithTaskVMFolder) error {
	taskIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil && item.TaskID != nil && *item.TaskID != "" {
			taskIDs = append(taskIDs, *item.TaskID)
		}
	}
	if len(taskIDs) == 0 {
		return nil
	}
	if err := utils.WaitTasks(ctx, client, taskIDs, time.Second); err != nil {
		return fmt.Errorf("wait for tasks: %w", err)
	}
	return nil
}

func waitForDeleteVMTasks(ctx context.Context, client *apiclient.Cloudtower, items []*models.WithTaskDeleteVM) error {
	taskIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil && item.TaskID != nil && *item.TaskID != "" {
			taskIDs = append(taskIDs, *item.TaskID)
		}
	}
	if len(taskIDs) == 0 {
		return nil
	}
	if err := utils.WaitTasks(ctx, client, taskIDs, time.Second); err != nil {
		return fmt.Errorf("wait for tasks: %w", err)
	}
	return nil
}

func waitForDeleteFolderTasks(ctx context.Context, client *apiclient.Cloudtower, items []*models.WithTaskDeleteVMFolder) error {
	taskIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil && item.TaskID != nil && *item.TaskID != "" {
			taskIDs = append(taskIDs, *item.TaskID)
		}
	}
	if len(taskIDs) == 0 {
		return nil
	}
	if err := utils.WaitTasks(ctx, client, taskIDs, time.Second); err != nil {
		return fmt.Errorf("wait for tasks: %w", err)
	}
	return nil
}
