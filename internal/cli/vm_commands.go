package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

func buildGetCmd(a *app) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or many resources",
	}

	getVMCmd := &cobra.Command{
		Use:     "vm [name]",
		Aliases: []string{"vms"},
		Short:   "Get VM details",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			cluster, _ := cmd.Flags().GetString("cluster")
			return a.runGetVM(name, cluster, cmd.OutOrStdout())
		},
	}
	getVMCmd.Flags().StringP("cluster", "c", "", "Filter by cluster name")

	getClusterCmd := &cobra.Command{
		Use:     "cluster [name]",
		Aliases: []string{"clusters"},
		Short:   "Get clusters",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := resolveNameArg(args, "")
			if err != nil {
				return err
			}
			return a.runGetClusters(name, cmd.OutOrStdout())
		},
	}

	getTemplateCmd := &cobra.Command{
		Use:     "template [name]",
		Aliases: []string{"templates"},
		Short:   "Get content library VM templates",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := resolveNameArg(args, "")
			if err != nil {
				return err
			}
			return a.runGetTemplates(name, cmd.OutOrStdout())
		},
	}

	getFolderCmd := &cobra.Command{
		Use:     "folder [name]",
		Aliases: []string{"folders"},
		Short:   "Get VM folders",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			return a.runGetFolders(name, cmd.OutOrStdout())
		},
	}

	getCmd.AddCommand(getVMCmd, getClusterCmd, getTemplateCmd, getFolderCmd)
	return getCmd
}

func buildDescribeCmd(a *app) *cobra.Command {
	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Show detailed information about a resource",
	}

	describeVMCmd := &cobra.Command{
		Use:   "vm [name]",
		Short: "Describe a VM",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a VM name")
			}
			return a.runDescribeVM(args[0], cmd.OutOrStdout())
		},
	}

	describeFolderCmd := &cobra.Command{
		Use:   "folder [name]",
		Short: "Describe a VM folder",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a folder name")
			}
			return a.runDescribeFolder(args[0], cmd.OutOrStdout())
		},
	}

	describeCmd.AddCommand(describeVMCmd, describeFolderCmd)
	return describeCmd
}

func buildCreateCmd(a *app) *cobra.Command {
	opts := &createVMOptions{
		VCPU:     2,
		MemoryGB: 4,
		Status:   "STOPPED",
		Firmware: "BIOS",
		Bus:      "VIRTIO",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources",
	}

	createVMCmd := &cobra.Command{
		Use:   "vm [name]",
		Short: "Create a VM from a template or as a blank VM",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := resolveNameArg(args, opts.Name)
			if err != nil {
				return err
			}
			opts.Name = name
			if err := opts.validate(); err != nil {
				return err
			}
			return a.runCreateVM(*opts, cmd.OutOrStdout())
		},
	}

	createVMCmd.Flags().StringVarP(&opts.Name, "name", "n", "", "VM name")
	createVMCmd.Flags().StringVar(&opts.Cluster, "cluster", "", "Cluster name")
	createVMCmd.Flags().StringVar(&opts.Template, "template", "", "Content library template name")
	createVMCmd.Flags().BoolVar(&opts.IsFullCopy, "is-full-copy", false, "Create a full copy from the template")
	createVMCmd.Flags().IntVar(&opts.VCPU, "vcpu", opts.VCPU, "Total vCPUs")
	createVMCmd.Flags().IntVar(&opts.MemoryGB, "memory", opts.MemoryGB, "Memory in GiB")
	createVMCmd.Flags().BoolVar(&opts.HA, "ha", false, "Enable HA")
	createVMCmd.Flags().StringVar(&opts.Status, "status", opts.Status, "Initial power status: STOPPED or RUNNING")
	createVMCmd.Flags().StringVar(&opts.Firmware, "firmware", opts.Firmware, "Firmware: BIOS or UEFI")
	createVMCmd.Flags().StringVar(&opts.Bus, "bus", opts.Bus, "Disk bus: VIRTIO or SCSI")
	createVMCmd.Flags().StringVar(&opts.ConnectVlanID, "connect-vlan-id", "", "VLAN ID used by the first NIC")
	createVMCmd.Flags().StringVar(&opts.VMVolumeID, "vm-volume-id", "", "Existing volume ID to mount")
	createVMCmd.Flags().StringVar(&opts.ElfImageID, "elf-image-id", "", "ELF image ID for the CD-ROM")
	createVMCmd.Flags().IntVar(&opts.DiskSizeGB, "disk-size", 0, "New disk size in GiB")
	createVMCmd.Flags().StringVar(&opts.DiskName, "disk-name", "", "New disk name")
	createVMCmd.Flags().StringVar(&opts.Password, "password", "", "Cloud-init default user password")
	createVMCmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Cloud-init hostname")
	createVMCmd.Flags().StringVar(&opts.UserData, "user-data", "", "Raw cloud-init user data")
	createVMCmd.Flags().StringSliceVar(&opts.Nameservers, "nameservers", nil, "Cloud-init nameservers")
	createVMCmd.Flags().StringSliceVar(&opts.PublicKeys, "public-key", nil, "Cloud-init SSH public keys")
	createVMCmd.Flags().StringVar(&opts.IPAddress, "ip-address", "", "Cloud-init IP address")
	createVMCmd.Flags().StringVar(&opts.Netmask, "netmask", "", "Cloud-init netmask")
	createVMCmd.Flags().IntVar(&opts.NicIndex, "nic-index", 0, "Cloud-init NIC index")

	createFolderCmd := &cobra.Command{
		Use:   "folder [name]",
		Short: "Create a VM folder",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a folder name")
			}
			cluster, _ := cmd.Flags().GetString("cluster")
			return a.runCreateFolder(args[0], cluster, cmd.OutOrStdout())
		},
	}
	createFolderCmd.Flags().String("cluster", "", "Cluster name (required)")

	createCmd.AddCommand(createVMCmd, createFolderCmd)
	return createCmd
}

func buildDeleteCmd(a *app) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
	}

	deleteVMCmd := &cobra.Command{
		Use:   "vm [name]",
		Short: "Delete a VM",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a VM name")
			}
			return a.runDeleteVM(args[0], cmd.OutOrStdout())
		},
	}

	deleteFolderCmd := &cobra.Command{
		Use:   "folder [name]",
		Short: "Delete a VM folder",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a folder name")
			}
			return a.runDeleteFolder(args[0], cmd.OutOrStdout())
		},
	}

	deleteCmd.AddCommand(deleteVMCmd, deleteFolderCmd)
	return deleteCmd
}

func buildStartCmd(a *app) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start resources",
	}

	startVMCmd := &cobra.Command{
		Use:   "vm [name]",
		Short: "Start a VM",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a VM name")
			}
			return a.runStartVM(args[0], cmd.OutOrStdout())
		},
	}
	startCmd.AddCommand(startVMCmd)
	return startCmd
}

func buildStopCmd(a *app) *cobra.Command {
	var force bool
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop resources",
	}

	stopVMCmd := &cobra.Command{
		Use:   "vm [name]",
		Short: "Force power off a VM",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a VM name")
			}
			return a.runStopVM(args[0], force, cmd.OutOrStdout())
		},
	}
	stopVMCmd.Flags().BoolVarP(&force, "force", "f", false, "Force power off")
	stopCmd.AddCommand(stopVMCmd)
	return stopCmd
}

func buildRestartCmd(a *app) *cobra.Command {
	var force bool
	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart resources",
	}

	restartVMCmd := &cobra.Command{
		Use:   "vm [name]",
		Short: "Restart a VM",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a VM name")
			}
			return a.runRestartVM(args[0], force, cmd.OutOrStdout())
		},
	}
	restartVMCmd.Flags().BoolVarP(&force, "force", "f", false, "Force restart")
	restartCmd.AddCommand(restartVMCmd)
	return restartCmd
}
