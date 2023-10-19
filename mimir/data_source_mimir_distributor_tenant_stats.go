package mimir

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nfx/go-htmltable"
	"golang.org/x/net/html"
)

type Stats struct {
	User            string `header:"User"`
	Series          int    `header:"# Series"`
	TotalIngestRate string `header:"Total Ingest Rate"`
	APIIngestRate   string `header:"API Ingest Rate"`
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
			"replication_factor": {
				Type:        schema.TypeInt,
				Description: "Stats replication factor",
				Computed:    true,
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

	var headers map[string]string
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

	// get replication factor
	var replicationFactor int
	doc, err := html.Parse(strings.NewReader(jobraw))
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to parse html: %v", err))
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "b" {
					re := regexp.MustCompile(`\d+`)
					replicationFactor, _ = strconv.Atoi(re.FindString(c.FirstChild.Data))
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// get the stats
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
		apiIngestRate, err := strconv.ParseFloat(stat.APIIngestRate, 64)
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

	if err := d.Set("replication_factor", replicationFactor); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("stats", stats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", StringHashcode(jobraw)))

	return nil
}
