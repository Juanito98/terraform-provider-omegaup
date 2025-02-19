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
var _ resource.Resource = &IdentityResource{}
var _ resource.ResourceWithImportState = &IdentityResource{}

func NewIdentityResource() resource.Resource {
	return &IdentityResource{}
}

func IdentityResourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Creates an identity associated to a group. It does not fit well with bulk identities resource.",

		Attributes: map[string]schema.Attribute{
			"group_alias": schema.StringAttribute{
				MarkdownDescription: "Group identifier to associate the identity.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Identifier of the identity within a group, in the form group:user.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"gender": schema.StringAttribute{
				Required: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"school_name": schema.StringAttribute{
				Description: "Shool name of the user associated.",
				Required:    true,
			},
			"country_id": schema.StringAttribute{
				Description: "Country id based on ISO 3166-2",
				Required:    true,
			},
			"state_id": schema.StringAttribute{
				Description: "Id of the state.",
				Required:    true,
			},
		},
	}
}

// IdentityResource defines the resource implementation.
type IdentityResource struct {
	client *apiclient.Client
}

// IdentityResourceModel describes the resource data model.
type IdentityResourceModel struct {
	GroupAlias types.String `tfsdk:"group_alias"`
	Username   types.String `tfsdk:"username"`
	Name       types.String `tfsdk:"name"`
	Gender     types.String `tfsdk:"gender"`
	Password   types.String `tfsdk:"password"`
	SchoolName types.String `tfsdk:"school_name"`
	CountryId  types.String `tfsdk:"country_id"`
	StateId    types.String `tfsdk:"state_id"`
}

func (r *IdentityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity"
}

func (r *IdentityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = IdentityResourceSchema()
}

func (r *IdentityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IdentityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IdentityResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.IdentityCreate(&apiclient.IdentityCreateRequest{
		GroupAlias: data.GroupAlias.ValueString(),
		Username:   data.Username.ValueString(),
		Name:       data.Name.ValueString(),
		Gender:     data.Gender.ValueString(),
		Password:   data.Password.ValueString(),
		SchoolName: data.SchoolName.ValueString(),
		CountryId:  data.CountryId.ValueString(),
		StateId:    data.StateId.ValueString(),
	})

	if err != nil {
		// Hacky: Check if the user is already created trying to do a no op with the password
		fixed := false
		if err := r.client.IdentityUpdate(&apiclient.IdentityUpdateRequest{
			GroupAlias:       data.GroupAlias.ValueString(),
			Username:         data.Username.ValueString(),
			OriginalUsername: data.Username.ValueString(),
			Name:             data.Name.ValueString(),
			Gender:           data.Gender.ValueString(),
			SchoolName:       data.SchoolName.ValueString(),
			CountryId:        data.CountryId.ValueString(),
			StateId:          data.StateId.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while attempting to create the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: "+err.Error(),
			)
		} else {
			// We might only need to re add it!
			err := r.client.GroupAddUser(&apiclient.GroupAddUserRequest{
				GroupAlias:      data.GroupAlias.ValueString(),
				UsernameOrEmail: data.Username.ValueString(),
			})
			if err == nil {
				fixed = true
			} else {
				resp.Diagnostics.AddError(
					"Unable to Create Resource",
					"An unexpected error occurred while attempting to create the resource. "+
						"Please retry the operation or report this issue to the provider developers.\n\n"+
						"Error: "+err.Error(),
				)
			}
		}
		if !fixed {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while attempting to create the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IdentityResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GroupMembers(&apiclient.GroupMembersRequest{
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

	// Look for the alias within the group
	var apiData *apiclient.GroupIdentity
	for _, identity := range group.Identities {
		if apiclient.EqualUsername(identity.Username, data.Username.ValueString()) {
			apiData = &identity
		}
	}

	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Username = types.StringValue(apiData.Username)
	data.Gender = types.StringValue(apiData.Gender)
	data.Name = types.StringValue(apiData.Name)
	data.CountryId = types.StringValue(apiData.CountryId)
	data.StateId = types.StringValue(apiData.StateId)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var oldData IdentityResourceModel
	var data IdentityResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.IdentityUpdate(&apiclient.IdentityUpdateRequest{
		GroupAlias:       data.GroupAlias.ValueString(),
		Username:         data.Username.ValueString(),
		OriginalUsername: oldData.Username.ValueString(),
		Name:             data.Name.ValueString(),
		Gender:           data.Gender.ValueString(),
		SchoolName:       data.SchoolName.ValueString(),
		CountryId:        data.CountryId.ValueString(),
		StateId:          data.StateId.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return
	}

	if !oldData.Password.Equal(data.Password) {
		// Password has changed.
		err = r.client.IdentityChangePassword(&apiclient.IdentityChangePasswordRequest{
			Username:   data.Username.ValueString(),
			GroupAlias: data.GroupAlias.ValueString(),
			Password:   data.Password.ValueString(),
		})

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while attempting to update the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: "+err.Error(),
			)

			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IdentityResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.GroupRemoveUser(&apiclient.GroupRemoveUserRequest{
		GroupAlias:      data.GroupAlias.ValueString(),
		UsernameOrEmail: data.Username.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return
	}
}

func (r *IdentityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
