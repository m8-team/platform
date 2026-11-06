import {approvalKpis, approvals} from './approvals.mock'
import {competitorHeatmap, competitorKpis, competitorMatches, marketLadder, priceIndexTrend} from './competitors.mock'
import {atRiskSkus, categoryAccuracy, forecastInsights, forecastKpis, forecastVsActual, inventoryRiskMatrix, modelDrivers} from './forecasts.mock'
import {integrations} from './integrations.mock'
import {markdownCandidates, markdownGuardrails, markdownKpis, markdownWindows} from './markdown.mock'
import {
  approvalSummary,
  markdownAlerts,
  marketHeatmap,
  overviewForecast,
  overviewInsights,
  overviewKpis,
  overviewRecommendations,
} from './overview.mock'
import {appliedVsActual, priceActionKpis, priceActions} from './priceActions.mock'
import {portfolioDistribution, productKpis, productOpportunities, productSegments, products} from './products.mock'
import {ruleGroups, rules} from './rules.mock'
import {markdownPlanner, priceImpact, scenarioComparison, scenarioSummary, simulationGuardrails, simulationKpis} from './simulation.mock'

const delay = 160

function resolveMock<T>(data: T): Promise<T> {
  return new Promise((resolve) => {
    window.setTimeout(() => resolve(data), delay)
  })
}

export function getOverviewDashboard() {
  return resolveMock({
    kpis: overviewKpis,
    forecast: overviewForecast,
    heatmap: marketHeatmap,
    insights: overviewInsights,
    markdownAlerts,
    recommendations: overviewRecommendations,
    approvalSummary,
  })
}

export function getPriceActions() {
  return resolveMock({kpis: priceActionKpis, actions: priceActions, appliedVsActual})
}

export function getProducts() {
  return resolveMock({kpis: productKpis, products, segments: productSegments, portfolioDistribution, opportunities: productOpportunities})
}

export function getCompetitors() {
  return resolveMock({kpis: competitorKpis, trend: priceIndexTrend, heatmap: competitorHeatmap, ladder: marketLadder, matches: competitorMatches})
}

export function getForecasts() {
  return resolveMock({kpis: forecastKpis, forecastVsActual, categoryAccuracy, inventoryRiskMatrix, insights: forecastInsights, atRiskSkus, modelDrivers})
}

export function getMarkdownCandidates() {
  return resolveMock({kpis: markdownKpis, candidates: markdownCandidates, windows: markdownWindows, guardrails: markdownGuardrails})
}

export function getSimulation() {
  return resolveMock({kpis: simulationKpis, scenarioComparison, priceImpact, guardrails: simulationGuardrails, planner: markdownPlanner, scenarioSummary})
}

export function getRules() {
  return resolveMock({groups: ruleGroups, rules})
}

export function getApprovals() {
  return resolveMock({kpis: approvalKpis, approvals})
}

export function getIntegrations() {
  return resolveMock({integrations})
}
