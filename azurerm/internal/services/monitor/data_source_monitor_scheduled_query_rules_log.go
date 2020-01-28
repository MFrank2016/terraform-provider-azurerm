package monitor

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceArmMonitorScheduledQueryRulesLog() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmMonitorScheduledQueryRulesLogRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"location": azure.SchemaLocationForDataSource(),

			"authorized_resources": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"azns_action": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_group": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"custom_webhook_payload": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email_subject": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"criteria": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dimension": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
										Elem:     schema.TypeString,
									},
									"operator": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"values": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"metric_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"data_source_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"frequency": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provisioning_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"query": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"query_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"severity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"throttling": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"time_window": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"trigger": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_trigger": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metric_column": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"metric_trigger_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"operator": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"threshold": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
						},
						"operator": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"threshold": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceArmMonitorScheduledQueryRulesLogRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ScheduledQueryRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Error: Scheduled Query Rule %q was not found", name)
		}
		return fmt.Errorf("Error reading Scheduled Query Rule: %+v", err)
	}

	d.SetId(*resp.ID)
	d.Set("resource_group_name", resourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	d.Set("description", resp.Description)
	if resp.Enabled == insights.True {
		d.Set("enabled", true)
	} else {
		d.Set("enabled", false)
	}

	var action insights.LogToMetricAction
	if err = d.Set("criteria", flattenAzureRmScheduledQueryRulesLogCriteria(action.Criteria)); err != nil {
		return fmt.Errorf("Error setting `criteria`: %+v", err)
	}

	if schedule := resp.Schedule; schedule != nil {
		if schedule.FrequencyInMinutes != nil {
			d.Set("frequency", schedule.FrequencyInMinutes)
		}
		if schedule.TimeWindowInMinutes != nil {
			d.Set("time_window", schedule.TimeWindowInMinutes)
		}
	}

	if source := resp.Source; source != nil {
		if source.AuthorizedResources != nil {
			d.Set("authorized_resources", utils.FlattenStringSlice(source.AuthorizedResources))
		}
		if source.DataSourceID != nil {
			d.Set("data_source_id", source.DataSourceID)
		}
		if source.Query != nil {
			d.Set("query", source.Query)
		}
		d.Set("query_type", string(source.QueryType))
	}

	if lastUpdated := resp.LastUpdatedTime; lastUpdated != nil {
		d.Set("last_updated_time", lastUpdated.Format(time.RFC3339))
	}
	d.Set("provisioning_state", resp.ProvisioningState)

	return tags.FlattenAndSet(d, resp.Tags)
}
