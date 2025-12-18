package mimir

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

// RuleGroups represents the complete YAML structure for Prometheus rules
type RuleGroups struct {
	Groups []RuleGroup `yaml:"groups"`
}

// RuleGroup represents a single rule group
type RuleGroup struct {
	Name            string   `yaml:"name"`
	Interval        string   `yaml:"interval,omitempty"`
	PartialResponse *bool    `yaml:"partial_response_strategy,omitempty"`
	Rules           []Rule   `yaml:"rules"`
	SourceTenants   []string `yaml:"source_tenants,omitempty"`
	Limit           int      `yaml:"limit,omitempty"`
}

// Rule represents both alerting and recording rules
type Rule struct {
	// Common fields
	Expr   string            `yaml:"expr"`
	Labels map[string]string `yaml:"labels,omitempty"`

	// Alerting rule fields
	Alert       string            `yaml:"alert,omitempty"`
	For         string            `yaml:"for,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`

	// Recording rule fields
	Record string `yaml:"record,omitempty"`
}

// resourceMimirRules creates the enhanced multi-group rules resource
func resourceMimirRules() *schema.Resource {
	return &schema.Resource{
		Description: `Manages multiple Grafana Mimir rule groups within a namespace. 
		This resource is designed to handle YAML files containing multiple rule groups, 
		such as those exported from mimirtool or monitoring mixins. Each rule group 
		is managed individually via the Mimir API, but they are tracked together as 
		a single Terraform resource for easier bulk management.`,

		CreateContext: resourceMimirRulesCreate,
		ReadContext:   resourceMimirRulesRead,
		UpdateContext: resourceMimirRulesUpdate,
		DeleteContext: resourceMimirRulesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceMimirRulesImport,
		},

		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The namespace for the rule groups",
				ValidateFunc: validation.StringLenBetween(1, 100),
			},

			"org_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The Organization ID. If not set, the Org ID defined in the provider block will be used.",
			},

			// Content input methods (mutually exclusive)
			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "YAML content containing rule groups. Mutually exclusive with 'content_file'.",
				ValidateFunc:  validateYAMLContent,
				ConflictsWith: []string{"content_file"},
			},

			"content_file": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Path to YAML file containing rule groups. Mutually exclusive with 'content'.",
				ValidateFunc:  validation.StringIsNotEmpty,
				ConflictsWith: []string{"content"},
			},

			// Management options
			"only_groups": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Explicit list of rule group names to manage. If not specified, all groups in the content will be managed. Use this to manage only specific groups from a larger YAML file.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"ignore_groups": {
				Type:          schema.TypeSet,
				Optional:      true,
				Description:   "List of rule group names to ignore from the content. Useful when you want to manage most groups but exclude specific ones.",
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"only_groups"},
			},

			// Read-only computed fields
			"managed_groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of rule group names actually managed by this resource",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"rule_names": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of all rule names actually managed by this resource",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"total_rules": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of rules across all managed groups",
			},

			"groups_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of rule groups managed by this resource",
			},

			"content_hash": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Hash of the rule configuration content",
			},

			// Detailed state for each group (computed)
			"groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Details of all managed rule groups",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Rule group name",
						},
						"interval": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Evaluation interval",
						},
						"rules_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of rules in this group",
						},
						"alerting_rules_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of alerting rules in this group",
						},
						"recording_rules_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of recording rules in this group",
						},
					},
				},
			},
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
			if err := validateRuleGroupsConfiguration(diff); err != nil {
				return err
			}

			// Calculate managed groups during plan phase for better diff output
			if diff.HasChange("content") || diff.HasChange("content_file") || diff.HasChange("only_groups") || diff.HasChange("ignore_groups") || diff.Id() == "" {
				// Parse the configuration to determine what will be managed
				var ruleGroups RuleGroups
				var err error

				if content := diff.Get("content").(string); content != "" {
					err = yaml.Unmarshal([]byte(content), &ruleGroups)
				} else if contentFile := diff.Get("content_file").(string); contentFile != "" {
					data, readErr := os.ReadFile(contentFile)
					if readErr == nil {
						err = yaml.Unmarshal(data, &ruleGroups)
					}
				}

				if err == nil && len(ruleGroups.Groups) > 0 {
					// Determine which groups will be managed
					allGroupNames := make([]string, len(ruleGroups.Groups))
					for i, group := range ruleGroups.Groups {
						allGroupNames[i] = group.Name
					}

					var managedGroups []string
					if onlyGroups := diff.Get("only_groups").(*schema.Set); onlyGroups != nil && onlyGroups.Len() > 0 {
						for _, name := range onlyGroups.List() {
							groupName := name.(string)
							if contains(allGroupNames, groupName) {
								managedGroups = append(managedGroups, groupName)
							}
						}
					} else if ignoreGroups := diff.Get("ignore_groups").(*schema.Set); ignoreGroups != nil && ignoreGroups.Len() > 0 {
						var ignored []string
						for _, name := range ignoreGroups.List() {
							ignored = append(ignored, name.(string))
						}
						for _, groupName := range allGroupNames {
							if !contains(ignored, groupName) {
								managedGroups = append(managedGroups, groupName)
							}
						}
					} else {
						managedGroups = allGroupNames
					}

					// Set the computed fields so they appear in the plan
					diff.SetNew("managed_groups", managedGroups)
					diff.SetNew("groups_count", len(managedGroups))

					// Calculate total rules and collect rule names
					totalRules := 0
					var ruleNames []string
					for _, group := range ruleGroups.Groups {
						if contains(managedGroups, group.Name) {
							totalRules += len(group.Rules)
							for _, rule := range group.Rules {
								if rule.Alert != "" {
									ruleNames = append(ruleNames, rule.Alert)
								} else if rule.Record != "" {
									ruleNames = append(ruleNames, rule.Record)
								}
							}
						}
					}
					diff.SetNew("total_rules", totalRules)
					diff.SetNew("rule_names", ruleNames)
				}
			}

			return nil
		},
	}
}

// Validation functions

func validateYAMLContent(val interface{}, key string) (warns []string, errs []error) {
	content := val.(string)
	if content == "" {
		errs = append(errs, fmt.Errorf("%q cannot be empty", key))
		return
	}

	var ruleGroups RuleGroups
	if err := yaml.Unmarshal([]byte(content), &ruleGroups); err != nil {
		errs = append(errs, fmt.Errorf("%q contains invalid YAML: %v", key, err))
		return
	}

	if err := validateRuleGroupsContent(ruleGroups); err != nil {
		errs = append(errs, fmt.Errorf("%q validation failed: %v", key, err))
	}

	return
}

func validateRuleGroupsContent(ruleGroups RuleGroups) error {
	if len(ruleGroups.Groups) == 0 {
		return fmt.Errorf("at least one rule group is required")
	}

	groupNames := make(map[string]bool)

	for i, group := range ruleGroups.Groups {
		// Check group name
		if group.Name == "" {
			return fmt.Errorf("group %d: name is required", i)
		}

		if !groupRuleNameRegexp.MatchString(group.Name) {
			return fmt.Errorf("invalid Group Rule Name %s. Must match the regex %s", group.Name, groupRuleNameRegexp)
		}

		if groupNames[group.Name] {
			return fmt.Errorf("group %d: duplicate group name '%s'", i, group.Name)
		}
		groupNames[group.Name] = true

		// Validate interval if specified
		if group.Interval != "" {
			if _, err := model.ParseDuration(group.Interval); err != nil {
				return fmt.Errorf("group %d (%s): invalid interval '%s': %v", i, group.Name, group.Interval, err)
			}
		}

		// Check rules
		if len(group.Rules) == 0 {
			return fmt.Errorf("group %d (%s): at least one rule is required", i, group.Name)
		}

		for j, rule := range group.Rules {
			if err := validateRule(rule, i, j, group.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateRule(rule Rule, groupIndex, ruleIndex int, groupName string) error {
	// Expression is required
	if rule.Expr == "" {
		return fmt.Errorf("group %d (%s), rule %d: 'expr' is required", groupIndex, groupName, ruleIndex)
	}

	// Must have either alert or record, but not both
	hasAlert := rule.Alert != ""
	hasRecord := rule.Record != ""

	if !hasAlert && !hasRecord {
		return fmt.Errorf("group %d (%s), rule %d: must specify either 'alert' or 'record'", groupIndex, groupName, ruleIndex)
	}

	if hasAlert && hasRecord {
		return fmt.Errorf("group %d (%s), rule %d: cannot specify both 'alert' and 'record'", groupIndex, groupName, ruleIndex)
	}

	// Alerting rule specific validation
	if hasAlert {
		// Validate alert name
		if !groupRuleNameRegexp.MatchString(rule.Alert) {
			return fmt.Errorf("group %d (%s), rule %d: invalid alert name '%s'. Must match the regex %s", groupIndex, groupName, ruleIndex, rule.Alert, groupRuleNameRegexp)
		}

		// Validate 'for' duration if specified
		if rule.For != "" {
			if _, err := model.ParseDuration(rule.For); err != nil {
				return fmt.Errorf("group %d (%s), rule %d: invalid 'for' duration '%s': %v", groupIndex, groupName, ruleIndex, rule.For, err)
			}
		}
	}

	// Recording rule specific validation
	if hasRecord {
		if !metricNameRegexp.MatchString(rule.Record) {
			return fmt.Errorf("group %d (%s), rule %d: invalid record name '%s'. Must match the regex %s", groupIndex, groupName, ruleIndex, rule.Record, metricNameRegexp)
		}

		// Recording rules shouldn't have 'for' or 'annotations'
		if rule.For != "" {
			return fmt.Errorf("group %d (%s), rule %d: recording rules cannot have 'for' field", groupIndex, groupName, ruleIndex)
		}
		if len(rule.Annotations) > 0 {
			return fmt.Errorf("group %d (%s), rule %d: recording rules cannot have annotations", groupIndex, groupName, ruleIndex)
		}
	}

	return nil
}

func validateRuleGroupsConfiguration(diff *schema.ResourceDiff) error {
	// Ensure exactly one input method is used
	hasContent := diff.Get("content").(string) != ""
	hasContentFile := diff.Get("content_file").(string) != ""

	if !hasContent && !hasContentFile {
		return fmt.Errorf("either 'content' or 'content_file' must be specified")
	}

	if hasContent && hasContentFile {
		return fmt.Errorf("'content' and 'content_file' are mutually exclusive")
	}

	return nil
}

// Resource CRUD operations

func resourceMimirRulesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	ruleGroups, err := parseRuleGroupsConfiguration(d)
	if err != nil {
		return diag.FromErr(err)
	}

	namespace := d.Get("namespace").(string)
	orgID := d.Get("org_id").(string)
	if orgID == "" {
		orgID = client.headers["X-Scope-OrgID"]
	}

	// Determine which groups to manage
	managedGroups := determineGroupsToManage(ruleGroups, d)
	if len(managedGroups) == 0 {
		return diag.FromErr(fmt.Errorf("no rule groups selected for management"))
	}

	// Create rule groups via API
	var createdGroups []string
	for _, group := range ruleGroups.Groups {
		if !contains(managedGroups, group.Name) {
			continue // Skip groups not selected for management
		}

		if err := createRuleGroup(client, namespace, orgID, group); err != nil {
			// Clean up any groups that were already created
			for _, createdGroup := range createdGroups {
				deleteRuleGroup(client, namespace, orgID, createdGroup)
			}
			return diag.FromErr(fmt.Errorf("failed to create rule group '%s': %w", group.Name, err))
		}
		createdGroups = append(createdGroups, group.Name)
	}

	// Add time delay before read to wait for event propagation
	if ruleGroupReadDelayAfterChangeDuration > 0 {
		time.Sleep(ruleGroupReadDelayAfterChangeDuration)
	}

	// Set computed fields
	setComputedFields(d, ruleGroups, managedGroups)

	// Generate resource ID
	resourceID := fmt.Sprintf("%s/%s", orgID, namespace)
	d.SetId(resourceID)

	return resourceMimirRulesRead(ctx, d, m)
}

func resourceMimirRulesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	namespace := d.Get("namespace").(string)
	orgID := d.Get("org_id").(string)
	if orgID == "" {
		orgID = client.headers["X-Scope-OrgID"]
	}

	// Read the current configuration to get managed groups
	ruleGroups, err := parseRuleGroupsConfiguration(d)
	if err != nil {
		return diag.FromErr(err)
	}

	managedGroups := determineGroupsToManage(ruleGroups, d)

	// Verify that all managed groups still exist
	headers := make(map[string]string)
	if orgID != "" {
		headers["X-Scope-OrgID"] = orgID
	}

	var existingGroups []string
	for _, groupName := range managedGroups {
		path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, groupName)
		_, err := client.sendRequest("ruler", "GET", path, "", headers)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				// Group was deleted outside of Terraform
				continue
			}
			return diag.FromErr(fmt.Errorf("failed to read rule group '%s': %w", groupName, err))
		}
		existingGroups = append(existingGroups, groupName)
	}

	// If no groups exist, mark resource as deleted
	if len(existingGroups) == 0 {
		d.SetId("")
		return nil
	}

	// Update computed fields based on what actually exists
	setComputedFields(d, ruleGroups, existingGroups)

	return nil
}

func resourceMimirRulesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	namespace := d.Get("namespace").(string)
	orgID := d.Get("org_id").(string)
	if orgID == "" {
		orgID = client.headers["X-Scope-OrgID"]
	}

	// Get new configuration
	newRuleGroups, err := parseRuleGroupsConfiguration(d)
	if err != nil {
		return diag.FromErr(err)
	}

	oldManagedGroups := d.Get("managed_groups").([]interface{})
	newManagedGroups := determineGroupsToManage(newRuleGroups, d)

	// Convert old managed groups to string slice
	var oldGroups []string
	for _, g := range oldManagedGroups {
		oldGroups = append(oldGroups, g.(string))
	}

	// Determine what needs to be done
	groupsToDelete := difference(oldGroups, newManagedGroups)
	groupsToCreateOrUpdate := newManagedGroups

	// Delete removed groups
	for _, groupName := range groupsToDelete {
		if err := deleteRuleGroup(client, namespace, orgID, groupName); err != nil {
			return diag.FromErr(fmt.Errorf("failed to delete rule group '%s': %w", groupName, err))
		}
	}

	// Create or update groups
	for _, group := range newRuleGroups.Groups {
		if !contains(groupsToCreateOrUpdate, group.Name) {
			continue
		}

		if err := createRuleGroup(client, namespace, orgID, group); err != nil {
			return diag.FromErr(fmt.Errorf("failed to create/update rule group '%s': %w", group.Name, err))
		}
	}

	// Add time delay before read to wait for event propagation
	if ruleGroupReadDelayAfterChangeDuration > 0 {
		time.Sleep(ruleGroupReadDelayAfterChangeDuration)
	}

	// Update computed fields
	setComputedFields(d, newRuleGroups, newManagedGroups)

	return resourceMimirRulesRead(ctx, d, m)
}

func resourceMimirRulesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	namespace := d.Get("namespace").(string)
	orgID := d.Get("org_id").(string)
	if orgID == "" {
		orgID = client.headers["X-Scope-OrgID"]
	}

	// Get list of managed groups from state
	managedGroupsInterface := d.Get("managed_groups").([]interface{})
	var managedGroups []string
	for _, g := range managedGroupsInterface {
		managedGroups = append(managedGroups, g.(string))
	}

	// Delete each managed rule group
	var errors []string
	for _, groupName := range managedGroups {
		if err := deleteRuleGroup(client, namespace, orgID, groupName); err != nil {
			errors = append(errors, fmt.Sprintf("failed to delete rule group '%s': %v", groupName, err))
		}
	}

	if len(errors) > 0 {
		return diag.FromErr(fmt.Errorf("errors during deletion: %s", strings.Join(errors, "; ")))
	}

	d.SetId("")
	return nil
}

func resourceMimirRulesImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Import format: orgID/namespace
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("import ID must be in format: orgID/namespace")
	}

	orgID := parts[0]
	namespace := parts[1]

	d.Set("org_id", orgID)
	d.Set("namespace", namespace)

	// Note: For import, the user will need to provide content/content_file afterward
	// We can't automatically detect the YAML content from the API

	return []*schema.ResourceData{d}, nil
}

// Helper functions

func parseRuleGroupsConfiguration(d *schema.ResourceData) (RuleGroups, error) {
	var ruleGroups RuleGroups

	if content := d.Get("content").(string); content != "" {
		if err := yaml.Unmarshal([]byte(content), &ruleGroups); err != nil {
			return ruleGroups, fmt.Errorf("failed to parse YAML content: %w", err)
		}
	} else if contentFile := d.Get("content_file").(string); contentFile != "" {
		data, err := os.ReadFile(contentFile)
		if err != nil {
			return ruleGroups, fmt.Errorf("failed to read file %s: %w", contentFile, err)
		}
		if err := yaml.Unmarshal(data, &ruleGroups); err != nil {
			return ruleGroups, fmt.Errorf("failed to parse YAML file %s: %w", contentFile, err)
		}
	} else {
		return ruleGroups, fmt.Errorf("no rule configuration provided")
	}

	return ruleGroups, validateRuleGroupsContent(ruleGroups)
}

func determineGroupsToManage(ruleGroups RuleGroups, d *schema.ResourceData) []string {
	allGroupNames := make([]string, len(ruleGroups.Groups))
	for i, group := range ruleGroups.Groups {
		allGroupNames[i] = group.Name
	}

	// If specific groups are named, use only those
	if onlyGroups := d.Get("only_groups").(*schema.Set); onlyGroups.Len() > 0 {
		var selected []string
		for _, name := range onlyGroups.List() {
			groupName := name.(string)
			if contains(allGroupNames, groupName) {
				selected = append(selected, groupName)
			}
		}
		return selected
	}

	// If ignore_groups is set, exclude those
	if ignoreGroups := d.Get("ignore_groups").(*schema.Set); ignoreGroups.Len() > 0 {
		var ignored []string
		for _, name := range ignoreGroups.List() {
			ignored = append(ignored, name.(string))
		}

		var selected []string
		for _, groupName := range allGroupNames {
			if !contains(ignored, groupName) {
				selected = append(selected, groupName)
			}
		}
		return selected
	}

	// Default: manage all groups
	return allGroupNames
}

func setComputedFields(d *schema.ResourceData, ruleGroups RuleGroups, managedGroups []string) {
	// Set managed_groups
	d.Set("managed_groups", managedGroups)
	d.Set("groups_count", len(managedGroups))

	// Calculate total rules and other stats
	var totalRules int
	var groupDetails []map[string]interface{}

	for _, group := range ruleGroups.Groups {
		if !contains(managedGroups, group.Name) {
			continue
		}

		alertingCount := 0
		recordingCount := 0

		for _, rule := range group.Rules {
			if rule.Alert != "" {
				alertingCount++
			} else if rule.Record != "" {
				recordingCount++
			}
		}

		totalRules += len(group.Rules)

		groupDetail := map[string]interface{}{
			"name":                  group.Name,
			"interval":              group.Interval,
			"rules_count":           len(group.Rules),
			"alerting_rules_count":  alertingCount,
			"recording_rules_count": recordingCount,
		}
		groupDetails = append(groupDetails, groupDetail)
	}

	d.Set("total_rules", totalRules)
	d.Set("groups", groupDetails)

	// Calculate content hash
	contentHash := calculateContentHash(ruleGroups, managedGroups)
	d.Set("content_hash", contentHash)
}

func calculateContentHash(ruleGroups RuleGroups, managedGroups []string) string {
	// Create a subset of rule groups that are actually managed
	managedRuleGroups := RuleGroups{}
	for _, group := range ruleGroups.Groups {
		if contains(managedGroups, group.Name) {
			managedRuleGroups.Groups = append(managedRuleGroups.Groups, group)
		}
	}

	data, _ := yaml.Marshal(managedRuleGroups)
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func createRuleGroup(client *apiClient, namespace, orgID string, group RuleGroup) error {
	headers := map[string]string{"Content-Type": "application/yaml"}
	if orgID != "" {
		headers["X-Scope-OrgID"] = orgID
	}

	// Convert group back to YAML for API
	yamlData, err := yaml.Marshal(group)
	if err != nil {
		return fmt.Errorf("failed to marshal rule group to YAML: %w", err)
	}

	path := fmt.Sprintf("/config/v1/rules/%s", namespace)
	_, err = client.sendRequest("ruler", "POST", path, string(yamlData), headers)
	return err
}

func deleteRuleGroup(client *apiClient, namespace, orgID, groupName string) error {
	headers := make(map[string]string)
	if orgID != "" {
		headers["X-Scope-OrgID"] = orgID
	}

	path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, groupName)
	_, err := client.sendRequest("ruler", "DELETE", path, "", headers)
	if err != nil && strings.Contains(err.Error(), "response code '404'") {
		// Group already doesn't exist, consider this success
		return nil
	}
	return err
}

// Utility functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func difference(a, b []string) []string {
	var result []string
	for _, item := range a {
		if !contains(b, item) {
			result = append(result, item)
		}
	}
	return result
}
