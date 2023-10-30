package resources

import (
	"context"
	"errors"
	"time"

	"github.com/camptocamp/terraform-provider-puppetdb/internal/log"
	"github.com/camptocamp/terraform-provider-puppetdb/internal/provider"
	"github.com/camptocamp/terraform-provider-puppetdb/internal/puppetdb"
	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Node struct {
	provider *provider.Provider
}

type NodeModel struct {
	CertificateName              types.String `tfsdk:"certname"`
	Deactivated                  types.String `tfsdk:"deactivated"`
	Expired                      types.String `tfsdk:"expired"`
	CachedCatalogStatus          types.String `tfsdk:"cached_catalog_status"`
	CatalogEnvironment           types.String `tfsdk:"catalog_environment"`
	FactsEnvironment             types.String `tfsdk:"facts_environment"`
	ReportEnvironment            types.String `tfsdk:"report_environment"`
	CatalogTimestamp             types.String `tfsdk:"catalog_timestamp"`
	FactsTimestamp               types.String `tfsdk:"facts_timestamp"`
	ReportTimestamp              types.String `tfsdk:"report_timestamp"`
	LatestReportCorrectiveChange types.String `tfsdk:"latest_report_corrective_change"`
	LatestReportHash             types.String `tfsdk:"latest_report_hash"`
	LatestReportNoop             types.Bool   `tfsdk:"latest_report_noop"`
	LatestReportNoopPending      types.Bool   `tfsdk:"latest_report_noop_pending"`
	LatestReportStatus           types.String `tfsdk:"latest_report_status"`
}

func (r *Node) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (r *Node) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"certname": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deactivated": schema.StringAttribute{
				Computed: true,
			},
			"expired": schema.StringAttribute{
				Computed: true,
			},
			"cached_catalog_status": schema.StringAttribute{
				Computed: true,
			},
			"catalog_environment": schema.StringAttribute{
				Computed: true,
			},
			"facts_environment": schema.StringAttribute{
				Computed: true,
			},
			"report_environment": schema.StringAttribute{
				Computed: true,
			},
			"catalog_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"facts_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"report_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"latest_report_corrective_change": schema.StringAttribute{
				Computed: true,
			},
			"latest_report_hash": schema.StringAttribute{
				Computed: true,
			},
			"latest_report_noop": schema.BoolAttribute{
				Computed: true,
			},
			"latest_report_noop_pending": schema.BoolAttribute{
				Computed: true,
			},
			"latest_report_status": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *Node) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, state NodeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	certificateName := plan.CertificateName.ValueString()

	node, err := retryGetNode(ctx, r.provider.Client(), certificateName)

	if err != nil {
		resp.Diagnostics.AddError("Failed to create node", "Reason: "+err.Error())

		return
	}

	nodeToNodeModel(node, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Node) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NodeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	certificateName := state.CertificateName.ValueString()

	node, err := getNode(ctx, r.provider.Client(), certificateName)

	if err != nil {
		if errors.Is(err, puppetdb.ErrNotFound) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read node", "Reason: "+err.Error())

		return
	}

	nodeToNodeModel(node, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Node) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan NodeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	certificateName := plan.CertificateName.ValueString()

	node, err := retryGetNode(ctx, r.provider.Client(), certificateName)

	if err != nil {
		resp.Diagnostics.AddError("Failed to update node", "Reason: "+err.Error())

		return
	}

	nodeToNodeModel(node, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Node) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NodeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	certificateName := state.CertificateName.ValueString()

	err := deleteNode(ctx, r.provider.Client(), certificateName)

	if err != nil && !errors.Is(err, puppetdb.ErrNotFound) {
		resp.Diagnostics.AddError("Failed to delete node", "Reason: "+err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Node) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := NodeModel{
		CertificateName: types.StringValue(req.ID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Node) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":                              schema.StringAttribute{},
					"certname":                        schema.StringAttribute{},
					"deactivated":                     schema.StringAttribute{},
					"expired":                         schema.StringAttribute{},
					"cached_catalog_status":           schema.StringAttribute{},
					"catalog_environment":             schema.StringAttribute{},
					"facts_environment":               schema.StringAttribute{},
					"report_environment":              schema.StringAttribute{},
					"catalog_timestamp":               schema.StringAttribute{},
					"facts_timestamp":                 schema.StringAttribute{},
					"report_timestamp":                schema.StringAttribute{},
					"latest_report_corrective_change": schema.StringAttribute{},
					"latest_report_hash":              schema.StringAttribute{},
					"latest_report_noop":              schema.BoolAttribute{},
					"latest_report_noop_pending":      schema.BoolAttribute{},
					"latest_report_status":            schema.StringAttribute{},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldState struct {
					ID                           types.String `tfsdk:"id"`
					CertificateName              types.String `tfsdk:"certname"`
					Deactivated                  types.String `tfsdk:"deactivated"`
					Expired                      types.String `tfsdk:"expired"`
					CachedCatalogStatus          types.String `tfsdk:"cached_catalog_status"`
					CatalogEnvironment           types.String `tfsdk:"catalog_environment"`
					FactsEnvironment             types.String `tfsdk:"facts_environment"`
					ReportEnvironment            types.String `tfsdk:"report_environment"`
					CatalogTimestamp             types.String `tfsdk:"catalog_timestamp"`
					FactsTimestamp               types.String `tfsdk:"facts_timestamp"`
					ReportTimestamp              types.String `tfsdk:"report_timestamp"`
					LatestReportCorrectiveChange types.String `tfsdk:"latest_report_corrective_change"`
					LatestReportHash             types.String `tfsdk:"latest_report_hash"`
					LatestReportNoop             types.Bool   `tfsdk:"latest_report_noop"`
					LatestReportNoopPending      types.Bool   `tfsdk:"latest_report_noop_pending"`
					LatestReportStatus           types.String `tfsdk:"latest_report_status"`
				}

				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)

				if resp.Diagnostics.HasError() {
					return
				}

				newState := NodeModel{
					CertificateName:              oldState.CertificateName,
					Deactivated:                  oldState.Deactivated,
					Expired:                      oldState.Expired,
					CachedCatalogStatus:          oldState.CachedCatalogStatus,
					CatalogEnvironment:           oldState.CatalogEnvironment,
					FactsEnvironment:             oldState.FactsEnvironment,
					ReportEnvironment:            oldState.ReportEnvironment,
					CatalogTimestamp:             oldState.CatalogTimestamp,
					FactsTimestamp:               oldState.FactsTimestamp,
					ReportTimestamp:              oldState.ReportTimestamp,
					LatestReportCorrectiveChange: oldState.LatestReportCorrectiveChange,
					LatestReportHash:             oldState.LatestReportHash,
					LatestReportNoop:             oldState.LatestReportNoop,
					LatestReportNoopPending:      oldState.LatestReportNoopPending,
					LatestReportStatus:           oldState.LatestReportStatus,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

func NewNode(p *provider.Provider) resource.Resource {
	r := &Node{
		provider: p,
	}

	var _ resource.Resource = r
	var _ resource.ResourceWithImportState = r
	var _ resource.ResourceWithUpgradeState = r

	return r
}

func init() {
	resources = append(resources, NewNode)
}

func getNode(ctx context.Context, client *puppetdb.Client, certificateName string) (*puppetdb.Node, error) {
	logFields := log.NodeFields(certificateName)

	tflog.Trace(ctx, "Requesting node", logFields)

	node, err := client.Query("query/v4/nodes/"+certificateName, "GET", nil)

	tflog.Trace(ctx, "Requested node", log.MergeFields(logFields, log.ErrorField(err), map[string]any{
		"node": node,
	}))

	return node, err
}

func retryGetNode(ctx context.Context, client *puppetdb.Client, certificateName string) (*puppetdb.Node, error) {
	logFields := log.NodeFields(certificateName)

	getNode := func() (*puppetdb.Node, error) {
		node, err := getNode(ctx, client, certificateName)

		if err != nil && !errors.Is(err, puppetdb.ErrNotFound) {
			err = backoff.Permanent(err)
			tflog.Error(ctx, "un")
		}

		tflog.Error(ctx, "deux")
		return node, err
	}

	return backoff.RetryNotifyWithData(getNode, backoff.WithContext(backoff.NewExponentialBackOff(), ctx),
		func(err error, delay time.Duration) {
			tflog.Trace(ctx, "Will retry requesting node after backoff delay", log.MergeFields(logFields, log.ErrorField(err), map[string]any{
				"delay": delay,
			}))
		},
	)
}

func deleteNode(ctx context.Context, client *puppetdb.Client, certificateName string) error {
	logFields := log.NodeFields(certificateName)

	tflog.Trace(ctx, "Requesting node deletion", logFields)

	_, err := client.Query("cmd/v1", "POST", &puppetdb.Command{
		Command: "deactivate node",
		Version: 3,
		Payload: map[string]string{
			"certname": certificateName,
		},
	})

	tflog.Trace(ctx, "Requested node deletion", log.MergeFields(logFields, log.ErrorField(err)))

	return err
}

func nodeToNodeModel(node *puppetdb.Node, nodeModel *NodeModel) {
	nodeModel.CertificateName = types.StringValue(node.Certname)
	nodeModel.Deactivated = types.StringValue(node.Deactivated)
	nodeModel.Expired = types.StringValue(node.Expired)
	nodeModel.CachedCatalogStatus = types.StringValue(node.CachedCatalogStatus)
	nodeModel.CatalogEnvironment = types.StringValue(node.CatalogEnvironment)
	nodeModel.FactsEnvironment = types.StringValue(node.FactsEnvironment)
	nodeModel.ReportEnvironment = types.StringValue(node.ReportEnvironment)
	nodeModel.CatalogTimestamp = types.StringValue(node.CatalogTimestamp)
	nodeModel.FactsTimestamp = types.StringValue(node.FactsTimestamp)
	nodeModel.ReportTimestamp = types.StringValue(node.ReportTimestamp)
	nodeModel.LatestReportCorrectiveChange = types.StringValue(node.LatestReportCorrectiveChange)
	nodeModel.LatestReportHash = types.StringValue(node.LatestReportHash)
	nodeModel.LatestReportNoop = types.BoolValue(node.LatestReportNoop)
	nodeModel.LatestReportNoopPending = types.BoolValue(node.LatestReportNoopPending)
	nodeModel.LatestReportStatus = types.StringValue(node.LatestReportStatus)
}
