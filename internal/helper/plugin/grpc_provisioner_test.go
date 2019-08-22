package plugin

import proto "github.com/hashicorp/terraform-plugin-sdk/internal/tfplugin5"

var _ proto.ProvisionerServer = (*GRPCProvisionerServer)(nil)
