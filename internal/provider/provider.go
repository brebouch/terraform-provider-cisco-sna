package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &snaProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &snaProvider{
			version: version,
		}
	}
}

// snaProvider is the provider implementation.
type snaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// snaProviderModel maps provider schema data to a Go type.
type snaProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// Metadata returns the provider type name.
func (p *snaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sna"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *snaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Secure Network Analytics.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for Secure Network Analytics API. May also be provided via SNA_HOST environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for Secure Network Analytics API. May also be provided via SNA_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for Secure Network Analytics API. May also be provided via SNA_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *snaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Secure Network Analytics client")
	// Retrieve provider data from configuration
	var config snaProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Secure Network Analytics API Host",
			"The provider cannot create the Secure Network Analytics API client as there is an unknown configuration value for the Secure Network Analytics API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SNA_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Secure Network Analytics API Username",
			"The provider cannot create the Secure Network Analytics API client as there is an unknown configuration value for the Secure Network Analytics API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SNA_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Secure Network Analytics API Password",
			"The provider cannot create the Secure Network Analytics API client as there is an unknown configuration value for the Secure Network Analytics API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SNA_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("SNA_HOST")
	username := os.Getenv("SNA_USERNAME")
	password := os.Getenv("SNA_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Secure Network Analytics API Host",
			"The provider cannot create the Secure Network Analytics API client as there is a missing or empty value for the Secure Network Analytics API host. "+
				"Set the host value in the configuration or use the SNA_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Secure Network Analytics API Username",
			"The provider cannot create the Secure Network Analytics API client as there is a missing or empty value for the Secure Network Analytics API username. "+
				"Set the username value in the configuration or use the SNA_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Secure Network Analytics API Password",
			"The provider cannot create the Secure Network Analytics API client as there is a missing or empty value for the Secure Network Analytics API password. "+
				"Set the password value in the configuration or use the SNA_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "sna_host", host)
	ctx = tflog.SetField(ctx, "sna_username", username)
	ctx = tflog.SetField(ctx, "sna_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "sna_password")

	tflog.Debug(ctx, "Creating Secure Network Analytics client")

	// Create a new Secure Network Analytics client using the configuration values
	client, err := sna.NewClient(&host, &username, &password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Secure Network Analytics API Client",
			"An unexpected error occurred when creating the Secure Network Analytics API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Secure Network Analytics Client Error: "+err.Error(),
		)
		return
	}

	// Make the Secure Network Analytics client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Secure Network Analytics client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *snaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCoffeesDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *snaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrderResource,
	}
}
