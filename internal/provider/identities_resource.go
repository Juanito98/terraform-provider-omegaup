// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"terraform-provider-omegaup/internal/apiclient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IdentitiesResource{}

func NewIdentitiesResource() resource.Resource {
	return &IdentitiesResource{}
}

// IdentitiesResource defines the resource implementation.
type IdentitiesResource struct {
	client *apiclient.Client
}

// IdentitiesResourceModel describes the resource data model.
type IdentitiesResourceModel struct {
	GroupAlias types.String            `tfsdk:"group_alias"`
	Identities []IdentityResourceModel `tfsdk:"identities"`
}

func getIdentitiesOfResource(datas IdentitiesResourceModel) []apiclient.Identity {
	identities := []apiclient.Identity{}
	for _, data := range datas.Identities {
		identities = append(identities, apiclient.Identity{
			GroupAlias: data.GroupAlias.ValueString(),
			Username:   data.Username.ValueString(),
			Name:       data.Name.ValueString(),
			Gender:     data.Gender.ValueString(),
			Password:   data.Password.ValueString(),
			SchoolName: data.SchoolName.ValueString(),
			CountryId:  data.CountryId.ValueString(),
			StateId:    data.StateId.ValueString(),
		})
	}
	return identities
}

func (r *IdentitiesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identities"
}

func (r *IdentitiesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	identitySchema := IdentityResourceSchema()

	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Creates a bulk identities associated to a group. It does not fit well with single identity resource.",

		Attributes: map[string]schema.Attribute{
			"group_alias": schema.StringAttribute{
				MarkdownDescription: "Group identifier to associate the identities.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identities": schema.ListNestedAttribute{
				MarkdownDescription: "List of identities",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username":    identitySchema.Attributes["username"],
						"name":        identitySchema.Attributes["name"],
						"gender":      identitySchema.Attributes["gender"],
						"password":    identitySchema.Attributes["password"],
						"school_name": identitySchema.Attributes["school_name"],
						"country_id":  identitySchema.Attributes["country_id"],
						"state_id":    identitySchema.Attributes["state_id"],
						// Output
						"group_alias": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *IdentitiesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IdentitiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IdentitiesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.IdentityBulkCreate(&apiclient.IdentityBulkCreateRequest{
		GroupAlias: data.GroupAlias.ValueString(),
		Identities: getIdentitiesOfResource(data),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Populate the group alias
	for k := range data.Identities {
		data.Identities[k].GroupAlias = data.GroupAlias
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentitiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IdentitiesResourceModel

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

	// Update values
	identities := []IdentityResourceModel{}
	for _, dataIdentity := range data.Identities {
		// Look it into the group members
		found := false
		for _, identity := range group.Identities {
			if dataIdentity.Username.ValueString() == identity.Username {
				dataIdentity.Name = types.StringValue(identity.Name)
				dataIdentity.CountryId = types.StringValue(identity.CountryId)
				dataIdentity.StateId = types.StringValue(identity.StateId)
				found = true
			}
		}
		if found {
			identities = append(identities, dataIdentity)
		}
	}
	data.Identities = identities

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentitiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var oldData IdentitiesResourceModel
	var data IdentitiesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Remove from group the identities no longer seen
	toRemove := []IdentityResourceModel{}
	for _, oldIdentity := range oldData.Identities {
		exists := false
		for _, identity := range data.Identities {
			if identity.Username.ValueString() == oldIdentity.Username.ValueString() {
				exists = true
			}
		}
		if !exists {
			toRemove = append(toRemove, oldIdentity)
		}
	}
	for _, identityToRemove := range toRemove {
		err := r.client.GroupRemoveUser(&apiclient.GroupRemoveUserRequest{
			GroupAlias:      data.GroupAlias.ValueString(),
			UsernameOrEmail: identityToRemove.Username.ValueString(),
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

	err := r.client.IdentityBulkCreate(&apiclient.IdentityBulkCreateRequest{
		GroupAlias: data.GroupAlias.ValueString(),
		Identities: getIdentitiesOfResource(data),
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

	// Populate the group alias
	for k := range data.Identities {
		data.Identities[k].GroupAlias = data.GroupAlias
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentitiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IdentitiesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, identityToRemove := range data.Identities {
		err := r.client.GroupRemoveUser(&apiclient.GroupRemoveUserRequest{
			GroupAlias:      data.GroupAlias.ValueString(),
			UsernameOrEmail: identityToRemove.Username.ValueString(),
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

}
