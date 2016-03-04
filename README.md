
# Azure VHD utilities for Go.

This project provides a Go package to read Virtual Hard Disk (VHD) file, a CLI interface to upload local VHD to Azure storage and to inspect a local VHD.

# Installation

    go get github.com/Microsoft/azure-vhd-utils-for-go

# Usage

### Upload local VHD to Azure storage as page blob

```bash
USAGE:
   vhd upload [command options] [arguments...]

OPTIONS:
   --localvhdpath       Path to source VHD in the local machine.
   --stgaccountname     Azure storage account name.
   --stgaccountkey      Azure storage account key.
   --containername      Name of the container holding destination page blob. (Default: vhds)
   --blobname           Name of the destination page blob.
```

The upload command uploads local VHD to Azure storage as page blob. Once uploaded, you can use Azure portal to register an image based on this page blob and use it to create Azure Virtual Machines.

Azure requires VHD to be in Fixed Disk format. The command converts Dynamic and Differencing Disk to Fixed Disk during upload process, the conversion will not consume any additional space in local machine.

In case of Fixed Disk, the command detects blocks containing zeros and those will not be uploaded. In case of expandable disks (dynamic and differencing) only the blocks those are marked as non-empty in
the Block Allocation Table (BAT) will be uploaded.

The blocks containing data will be uploaded as chunks of 2 MB pages. Consecutive blocks will be merged to create 2 MB pages if the block size of disk is less than 2 MB. If the block size is greater than 2 MB, 
tool will split them as 2 MB pages.  

### Inspect local VHD

A subset of command are exposed under inspect command for inspecting various segments of VHD in the local machine.

#### Show VHD footer

```bash
USAGE:
   vhd inspect footer [command options] [arguments...]

OPTIONS:
   --path   Path to VHD.
```

#### Show VHD header of an expandable disk

```bash
USAGE:
   vhd inspect header [command options] [arguments...]

OPTIONS:
   --path   Path to VHD.
```

Only expandable disks (dynamic and differencing) VHDs has header.

#### Show Block Allocation Table (BAT) of an expandable disk

```bash
USAGE:
   vhd inspect bat [command options] [arguments...]

OPTIONS:
   --path           Path to VHD.
   --start-range    Start range.
   --end-range      End range.
   --skip-empty     Do not show BAT entries pointing to empty blocks.
```

Only expandable disks (dynamic and differencing) VHDs has BAT.

#### Show block general information

```bash
USAGE:
   vhd inspect block info [command options] [arguments...]

OPTIONS:
   --path   Path to VHD.
```

This command shows the total number blocks, block size and size of block sector

### Show sector bitmap of an expandable disk's block

```bash
USAGE:
   vhd inspect block bitmap [command options] [arguments...]

OPTIONS:
   --path           Path to VHD.
   --block-index    Index of the block.
   
```

# License

This project is published under [MIT License](LICENSE).