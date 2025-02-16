// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-omegaup/internal/apiclient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &GroupMemberResource{}
var _ resource.ResourceWithImportState = &GroupMemberResource{}

func NewGroupMemberResource() resource.Resource {
	return &GroupMemberResource{}
}

// GroupMemberResource defines the resource implementation.
type GroupMemberResource struct {
	client *apiclient.Client
}

// GroupMemberResourceModel describes the resource data model.
type GroupMemberResourceModel struct {
	GroupAlias types.String `tfsdk:"group_alias"`
	Username   types.String `tfsdk:"username"`
}

func (r *GroupMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_member"
}

func (r *GroupMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Creates a new user permission for a group.",

		Attributes: map[string]schema.Attribute{
			"group_alias": schema.StringAttribute{
				MarkdownDescription: "The alias used to identify the group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "OmegaUp username to add to the group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *GroupMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	r.client = client
}

func (r *GroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupMemberResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	addReq := &apiclient.GroupAddUserRequest{
		GroupAlias:      data.GroupAlias.ValueString(),
		UsernameOrEmail: data.Username.ValueString(),
	}

	err := r.client.GroupAddUser(addReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.GroupAlias = types.StringValue(addReq.GroupAlias)
	data.Username = types.StringValue(addReq.UsernameOrEmail)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupMemberResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	members, err := r.client.GroupMembers(&apiclient.GroupMembersRequest{
		GroupAlias: data.GroupAlias.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Refresh Resource",
			"An unexpected error occurred while attempting to refresh resource state. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	userExists := false
	for _, identity := range members.Identities {
		if identity.Username == data.Username.ValueString() {
			userExists = true
		}
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	if !userExists {
		resp.Diagnostics.Append(resp.State.Set(ctx, nil)...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Does not support [Update]
	resp.Diagnostics.AddError("Unable to Update Resource", "Resource does not support update. Please retry the operation or report this issue to the provider developers.")
}

func (r *GroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupMemberResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Convert from Terraform data model into API data model
	deleteReq := &apiclient.GroupRemoveUserRequest{
		GroupAlias:      data.GroupAlias.ValueString(),
		UsernameOrEmail: data.Username.ValueString(),
	}

	err := r.client.GroupRemoveUser(deleteReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

func (r *GroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: group_alias,username. Got: %q", req.ID),
		)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	var data GroupMemberResourceModel
	data.GroupAlias = types.StringValue(idParts[0])
	data.Username = types.StringValue(idParts[1])

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
