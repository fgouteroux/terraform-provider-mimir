package mimir

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Stats struct {
	User            string  `json:"UserID"`
	Series          int     `json:"numSeries"`
	TotalIngestRate float64 `json:"ingestionRate"`
	APIIngestRate   float64 `json:"APIIngestionRate"`
	RuleIngestRate  float64 `json:"RuleIngestionRate"`
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
				Type:        schema.TypeList,
				Description: "Stats list",
				Computed:    true,
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

	headers := map[string]string{"Accept": "json"}
	jobraw, err := client.sendRequest("distributor", "GET", "/all_user_stats", "", headers)

	baseMsg := "Cannot read user stats"
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		if strings.Contains(err.Error(), "response code '404'") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// unmarshal the json using json/encoding into Stats struct
	var output []Stats
	err = json.Unmarshal([]byte(jobraw), &output)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to unmarshal json: %v", err))
	}

	var stats []map[string]interface{}
	// transform the output into a list of maps
	for _, stat := range output {
		// trim float to 2 decimal places
		stat.TotalIngestRate = math.Round(stat.TotalIngestRate*100) / 100
		stat.APIIngestRate = math.Round(stat.APIIngestRate*100) / 100
		stat.RuleIngestRate = math.Round(stat.RuleIngestRate*100) / 100
		stats = append(stats, map[string]interface{}{
			"user":              stat.User,
			"series":            stat.Series,
			"total_ingest_rate": stat.TotalIngestRate,
			"api_ingest_rate":   stat.APIIngestRate,
			"rule_ingest_rate":  stat.RuleIngestRate,
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
