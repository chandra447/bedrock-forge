package generator

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateAgentAliases creates agent alias resources for an agent
func (g *HCLGenerator) generateAgentAliases(body *hclwrite.Body, agentName string, aliases []models.AgentAlias) error {
	if len(aliases) == 0 {
		return nil
	}

	agentResourceName := g.sanitizeResourceName(agentName)

	for _, alias := range aliases {
		aliasResourceName := fmt.Sprintf("%s_%s_alias", agentResourceName, g.sanitizeResourceName(alias.Name))

		g.logger.WithField("agent", agentName).WithField("alias", alias.Name).Debug("Generating agent alias")

		// Create module block for agent alias
		moduleBlock := body.AppendNewBlock("module", []string{aliasResourceName})
		moduleBody := moduleBlock.Body()

		// Set module source
		moduleSource := fmt.Sprintf("%s//modules/bedrock-agent-alias", g.config.ModuleRegistry)
		if g.config.ModuleVersion != "" {
			moduleSource += fmt.Sprintf("?ref=%s", g.config.ModuleVersion)
		}
		moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

		// Set required attributes
		moduleBody.SetAttributeValue("agent_alias_name", cty.StringVal(alias.Name))
		moduleBody.SetAttributeValue("agent_id", cty.StringVal(fmt.Sprintf("${module.%s.agent_id}", agentResourceName)))

		// Optional description
		if alias.Description != "" {
			moduleBody.SetAttributeValue("description", cty.StringVal(alias.Description))
		}

		// Tags
		if len(alias.Tags) > 0 {
			tagValues := make(map[string]cty.Value)
			for key, value := range alias.Tags {
				tagValues[key] = cty.StringVal(value)
			}
			moduleBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
		}

		body.AppendNewline()

		g.logger.WithField("agent", agentName).WithField("alias", alias.Name).Info("Generated agent alias module")
	}

	return nil
}
