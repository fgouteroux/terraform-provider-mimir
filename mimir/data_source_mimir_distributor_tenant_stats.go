package mimir

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nfx/go-htmltable"
)

type Stats struct {
	User            string `header:"User"`
	Series          int    `header:"# Series"`
	TotalIngestRate string `header:"Total Ingest Rate"`
	ApiIngestRate   string `header:"API Ingest Rate"`
	RuleIngestRate  string `header:"Rule Ingest Rate"`
}

func dataSourcemimirDistributorTenantStats() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcemimirDistributorTenantStatsRead,

		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Description: "Query specific user stats",
				ForceNew:    true,
				Optional:    true,
			},
			"stats": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"series": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"total_ingest_rate": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"api_ingest_rate": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"rule_ingest_rate": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
		}, /* End schema */

	}
}

func dataSourcemimirDistributorTenantStatsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	user := d.Get("user").(string)

	var headers map[string]string
	jobraw, err := client.sendRequest("distributor", "GET", "/all_user_stats", "", headers)

	baseMsg := fmt.Sprintf("Cannot read user stats")
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		if strings.Contains(err.Error(), "response code '404'") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	output, err := htmltable.NewSliceFromString[Stats](jobraw)

	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to decode stats data: %v", err))
	}

	// transform the output into a list of maps
	var stats []map[string]interface{}
	for _, stat := range output {

		// convert the string values to float
		totalIngestRate, err := strconv.ParseFloat(stat.TotalIngestRate, 64)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to convert total ingest rate to float: %v", err))
		}
		apiIngestRate, err := strconv.ParseFloat(stat.ApiIngestRate, 64)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to convert api ingest rate to float: %v", err))
		}
		ruleIngestRate, err := strconv.ParseFloat(stat.RuleIngestRate, 64)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to convert rule ingest rate to float: %v", err))
		}

		stats = append(stats, map[string]interface{}{
			"user":              stat.User,
			"series":            stat.Series,
			"total_ingest_rate": totalIngestRate,
			"api_ingest_rate":   apiIngestRate,
			"rule_ingest_rate":  ruleIngestRate,
		})
	}

	// if user is specified then filter the stats
	if user != "" {
		var filteredStats []map[string]interface{}
		for _, stat := range stats {
			if stat["user"] == user {
				filteredStats = append(filteredStats, stat)
			}
		}
		stats = filteredStats
	}

	if err := d.Set("stats", stats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", StringHashcode(jobraw)))

	return nil
}
