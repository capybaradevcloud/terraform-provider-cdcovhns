package provider

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/capybaradevcloud/terraform-provider-cdcovhns/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CDCOvhNSResource{}
var _ resource.ResourceWithImportState = &CDCOvhNSResource{}

func NewCDCOvhNSResource() resource.Resource {
	return &CDCOvhNSResource{}
}

type CDCOvhNSResource struct {
	client *api.APIClient
}

type CDCOvhNSResourceModel struct {
	ServiceName types.String                   `tfsdk:"service_name"`
	Type        types.String                   `tfsdk:"type"`
	NameServers map[string]CDCNameServersModel `tfsdk:"name_servers"`
}

type CDCNameServersModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Host     types.String `tfsdk:"host"`
	IP       types.String `tfsdk:"ip"`
	IsUsed   types.Bool   `tfsdk:"is_used"`
	ToDelete types.Bool   `tfsdk:"to_delete"`
}

func (r *CDCOvhNSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_name_servers"
}

func (r *CDCOvhNSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "OVH Name Server resource",

		Attributes: map[string]schema.Attribute{
			"service_name": schema.StringAttribute{
				MarkdownDescription: "Domain name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "OVH Name Servers type - 'external' (if external name servers like Cloudflare) or 'hosted' if OVH Name Servers.",
				Default:     stringdefault.StaticString("external"),
			},
			"name_servers": schema.MapNestedAttribute{
				Required: true,
				Validators: []validator.Map{

					mapvalidator.SizeAtLeast(2),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"host": schema.StringAttribute{
							Required:    true,
							Description: "DNS Hostname",
						},
						"ip": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "DNS IP address",
							Default:     stringdefault.StaticString(""),
						},
						"is_used": schema.BoolAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"to_delete": schema.BoolAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r CDCOvhNSResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data *CDCOvhNSResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Type.IsNull() && (data.Type.ValueString() != api.NSExternal && data.Type.ValueString() != api.NSHosted) {
		resp.Diagnostics.AddAttributeError(
			path.Root("type"),
			"Wrong Name Servers type",
			"Choose 'external' or 'hosted'",
		)
	}

	for key, NameServer := range data.NameServers {
		if (strings.Trim(NameServer.IP.ValueString(), `"`) != "" || !NameServer.IP.IsNull()) && net.ParseIP(NameServer.IP.ValueString()) == nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("name_servers").AtMapKey(key),
				"IP is not valid",
				"Provide real ip address",
			)
		}
	}
}

func (r *CDCOvhNSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *CDCOvhNSResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, state *CDCOvhNSResourceModel
	var serviceName string

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if req.State.Raw.IsNull() {
		resp.Diagnostics.AddError(
			"Nothing in state!",
			fmt.Sprint(
				"Please import Name Servers first, eg:\n",
				"terraform import cdcovhns_name_servers.<YOUR_RESOURCE_NAME> <DOMAIN_NAME>\n",
				"See documentantion for more information",
			),
		)
	}

	if plan != nil {
		serviceName = plan.ServiceName.ValueString()
	} else {
		serviceName = state.ServiceName.ValueString()
	}

	currentTasks := r.client.CheckCurrentTaskState(serviceName)
	if currentTasks != nil {
		resp.Diagnostics.AddError(
			"Some task are already in operation",
			"Cannot do anything because some task on Name Servers are already in TODO or DOING state: \n"+currentTasks.Error(),
		)
		return
	}

	// If the entire plan is null, the resource is planned for destruction.
	if req.Plan.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"Resource Destruction Considerations",
			"Destroy operation will revert the Name Servers to default OVH Name Servers",
		)
	}
}

func (r *CDCOvhNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Cannot create Name Servers on OVH",
		fmt.Sprint(
			"Please import name servers first:\n",
			"terraform import cdcovhns_name_servers.<YOUR_RESOURCE_NAME> <DOMAIN_NAME>\n",
			"See documentantion for more information",
		),
	)
}

func (r *CDCOvhNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *CDCOvhNSResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()

	currentTasks := r.client.CheckCurrentTaskState(serviceName)
	if currentTasks != nil {
		resp.Diagnostics.AddWarning(
			"Some task already in operation",
			"READ: Some task are performed when reading state. Watch out and refresh state later \n"+currentTasks.Error(),
		)
	}

	nameServers, err := r.client.GetNameServersFromAPI(serviceName)
	nsTypeResponse, nsTypeErr := r.client.GetNameServersType(serviceName)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Name Servers",
			"READ: Could not read current name servers, unexpected error: "+err.Error(),
		)
		return
	}

	if nsTypeErr != nil {
		resp.Diagnostics.AddError(
			"Error reading Name Servers",
			"READ: Could not read current name servers, unexpected error: "+nsTypeErr.Error(),
		)
		return
	}

	data.Type = types.StringValue(nsTypeResponse.NameServerType)
	data.NameServers = convertReponseToResourceNS(nameServers)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *CDCOvhNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *CDCOvhNSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	serviceName := plan.ServiceName.ValueString()

	currentTasks := r.client.CheckCurrentTaskState(serviceName)
	if currentTasks != nil {
		resp.Diagnostics.AddError(
			"Some task already in operation",
			"UPDATE: Cannot do anything because some task on Name Servers are already in TODO or doing state: \n"+currentTasks.Error(),
		)
		return
	}

	if plan.Type.ValueString() != state.Type.ValueString() {
		nsSetTypeErr := r.client.SetNameServerType(
			serviceName,
			plan.Type.ValueString(),
		)

		if nsSetTypeErr != nil {
			resp.Diagnostics.AddError(
				"Error updating name servers",
				"UPDATE: Could not update current name servers, unexpected error: "+nsSetTypeErr.Error(),
			)
		}
	}

	nameServerCreatePayloads := []*api.NameServerCreatePayload{}
	for _, NameServer := range plan.NameServers {
		newNsPayload := &api.NameServerCreatePayload{
			Host: NameServer.Host.ValueString(),
			IP:   NameServer.IP.ValueString(),
		}
		nameServerCreatePayloads = append(nameServerCreatePayloads, newNsPayload)
	}

	updatedNsData := &api.NameServerUpdateRequest{
		NameServers: nameServerCreatePayloads,
	}

	generatedApiTask, err := r.client.UpdateNameServers(serviceName, updatedNsData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Name Servers",
			"UPDATE_NS: Could not update current name servers, unexpected error: "+err.Error(),
		)
		return
	}

	// Wait for update in API
	taskErr := make(chan error)
	go r.client.CheckOVHTask(taskErr, generatedApiTask.ServiceName, generatedApiTask.ID)
	if err := <-taskErr; err != nil {
		resp.Diagnostics.AddError(
			"Error updating name servers",
			"UPDATE: Could not update current name servers, unexpected error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *CDCOvhNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *CDCOvhNSResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()

	currentTasks := r.client.CheckCurrentTaskState(serviceName)
	if currentTasks != nil {
		resp.Diagnostics.AddError(
			"Some task already in operation",
			"DELETE: Cannot do anything because some task on Name Servers are already in TODO or DOING state: \n"+currentTasks.Error(),
		)
		return
	}

	err := r.client.DeleteNameServers(serviceName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error when deleting name servers",
			"DELETE: Could not delete current Name Servers, unexpected error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.AddWarning(
		fmt.Sprintf(
			"Name Servers for domain %s reseted to default OVH Name Servers.",
			serviceName,
		),
		"Run 'terraform import cdcovhns_name_servers.<RESOURCE_NAME> <DOMAIN_NAME>' to import current state",
	)
}

func (r *CDCOvhNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serviceName := req.ID

	currentTasks := r.client.CheckCurrentTaskState(serviceName)
	if currentTasks != nil {
		resp.Diagnostics.AddError(
			"Some task already in operation",
			"IMPORT: Some task are performed when importing state. Should nod import now. Wait for task end \n"+currentTasks.Error(),
		)
		return
	}

	nameServers, err := r.client.GetNameServersFromAPI(serviceName)
	nsType, nsTypeErr := r.client.GetNameServersType(serviceName)

	if err != nil || nsTypeErr != nil {
		resp.Diagnostics.AddError(
			"Error reading endpoint",
			"IMPORT: Could not read current name servers, unexpected error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_name"), serviceName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), types.StringValue(nsType.NameServerType))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name_servers"), convertReponseToResourceNS(nameServers))...)
}

func convertReponseToResourceNS(nameServers map[string]api.NameServerOvhResponse) map[string]CDCNameServersModel {
	resourceNameServers := make(map[string]CDCNameServersModel)

	for key, data := range nameServers {
		resourceNameServers[key] = CDCNameServersModel{
			ID:       types.Int64Value(int64(data.Id)),
			Host:     types.StringValue(data.GetHost()),
			IP:       types.StringValue(data.GetIP()),
			IsUsed:   types.BoolValue(data.IsUsed),
			ToDelete: types.BoolValue(data.ToDelete),
		}
	}
	return resourceNameServers
}
