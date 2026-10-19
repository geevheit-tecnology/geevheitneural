import { apiClient } from "./index";

export interface MoneyWaitingResponse {
  total_cents: number;
  count: number;
  formatted: string;
}

export interface CardItem {
  id: string;
  type: "orcamento" | "bug" | "suporte" | "financeiro";
  client_id: string;
  client_name: string;
  app_id?: string;
  app_name: string;
  title: string;
  price_cents: number;
  est_cost_cents: number;
  currency: string;
  stage: string;
  priority: number;
  due_date?: string;
  created_at: string;
  updated_at: string;
}

export interface RiskPoint {
  [key: string]: unknown;
}

export interface CardDetails {
  card: CardItem;
  events: Array<{
    id: string;
    kind: string;
    payload: Record<string, unknown>;
    actor_type: string;
    at: string;
  }>;
  risk_report?: {
    id: string;
    card_id: string;
    tests: Array<[string, string, string]>;
    secrets_scan: Array<[string, string, string]>;
    outdated_deps: Array<[string, string, string]>;
    lgpd_points: Array<[string, string, string]>;
  };
  is_gate: boolean;
  actions: string[];
  margin_cents: number;
}

export interface AppItem {
  id: string;
  client_id?: string;
  client_name: string;
  name: string;
  hosting: string;
  database: string;
  repository: string;
  domain: string;
  status: string;
  monthly_cost_cents: number;
  monthly_revenue_cents: number;
  net_margin_cents: number;
  is_costly: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateAppInput {
  name: string;
  hosting: string;
  database: string;
  repository?: string;
  domain?: string;
  client_id?: string;
}

export interface CostByService {
  service_name: string;
  cost_cents: number;
}

export interface LedgerEntry {
  id: string;
  app_id?: string;
  app_name: string;
  client_id?: string;
  client_name: string;
  kind: "receita" | "custo";
  amount_cents: number;
  currency: string;
  occurred_on: string;
  service: string;
  description: string;
  created_at: string;
}

export interface FinancesOverview {
  total_revenue_cents: number;
  total_cost_cents: number;
  net_profit_cents: number;
  costs_by_service: CostByService[];
  recent_entries: LedgerEntry[];
  suggestion: string;
}

export interface CreateFinanceEntryInput {
  kind: "receita" | "custo";
  amount_cents: number;
  service: string;
  description?: string;
  occurred_on?: string;
  app_id?: string;
}

export interface ConnectorItem {
  id: string;
  name: string;
  active: boolean;
  secret_ref: string;
  scope: string;
}

export const api = {
  getMoneyWaiting: () => apiClient<MoneyWaitingResponse>("/v1/money-waiting"),
  
  listCards: () => apiClient<{ cards: CardItem[]; count: number }>("/v1/cards"),
  
  getCardDetails: (id: string) => apiClient<CardDetails>(`/v1/cards/${id}`),
  
  decideGate: (id: string, decision: "aprovado" | "mais_info" | "recusado", note?: string) =>
    apiClient<{ success: boolean; card_id: string; decision: string }>(`/v1/cards/${id}/decision`, {
      method: "POST",
      body: JSON.stringify({ decision, note: note || "" }),
    }),

  pauseSystem: (reason?: string) =>
    apiClient<{ success: boolean; paused: boolean }>("/v1/system/pause", {
      method: "POST",
      body: JSON.stringify({ reason: reason || "Pausa de emergência pelo painel" }),
    }),

  resumeSystem: () =>
    apiClient<{ success: boolean; paused: boolean }>("/v1/system/resume", {
      method: "POST",
    }),

  listApps: () => apiClient<{ apps: AppItem[]; count: number }>("/v1/apps"),

  createApp: (data: CreateAppInput) =>
    apiClient<AppItem>("/v1/apps", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  deleteApp: (id: string) =>
    apiClient<{ success: boolean; id: string }>(`/v1/apps/${id}`, {
      method: "DELETE",
    }),

  getFinances: () => apiClient<FinancesOverview>("/v1/finances"),

  createFinanceEntry: (data: CreateFinanceEntryInput) =>
    apiClient<LedgerEntry>("/v1/finances/entries", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  listConnectors: () => apiClient<{ connectors: ConnectorItem[]; count: number }>("/v1/connectors"),

  processCommand: (input: string) =>
    apiClient<{
      intent: string;
      message: string;
      requires_confirmation: boolean;
      confirmation_token?: string;
      confirmation_prompt?: string;
      data?: unknown;
    }>("/v1/commands", {
      method: "POST",
      body: JSON.stringify({ input }),
    }),

  confirmAction: (token: string) =>
    apiClient<{ success: boolean; message: string }>("/v1/commands/confirm", {
      method: "POST",
      body: JSON.stringify({ token }),
    }),

  getRiskReport: (cardId: string) =>
    apiClient<Record<string, unknown>>(`/v1/cards/${cardId}/risk`),

  saveRiskReport: (cardId: string, data: Record<string, unknown>) =>
    apiClient<Record<string, unknown>>(`/v1/cards/${cardId}/risk`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  // Sprint 02: Fábrica Autônoma, Pipelines, Playbooks e Escopo
  startBuild: (cardId: string, spendLimitCents: number = 10000) =>
    apiClient<{
      id: string;
      card_id: string;
      stage: string;
      active_role: string;
      spend_total_cents: number;
      spend_limit_cents: number;
      status: string;
      current_task: string;
    }>(`/v1/cards/${cardId}/pipeline/start`, {
      method: "POST",
      body: JSON.stringify({ spend_limit_cents: spendLimitCents }),
    }),

  getPipeline: (cardId: string) =>
    apiClient<{
      id: string;
      card_id: string;
      stage: string;
      active_role: string;
      spend_total_cents: number;
      spend_limit_cents: number;
      status: string;
      current_task: string;
    }>(`/v1/cards/${cardId}/pipeline`),

  listPlaybooks: () =>
    apiClient<{
      playbooks: Array<{
        product_type: string;
        title: string;
        description: string;
        stack: Record<string, unknown>;
        checklist: string[];
        autonomy_level: number;
      }>;
    }>("/v1/playbooks"),

  listModules: () =>
    apiClient<{
      modules: Array<{
        name: string;
        category: string;
        description: string;
        code_templates: Record<string, unknown>;
        is_reviewed: boolean;
      }>;
    }>("/v1/modules"),

  getFactoryScoreboard: () =>
    apiClient<{
      avg_lead_time_hours: number;
      total_ai_cost_cents: number;
      total_infra_cost_cents: number;
      total_margin_cents: number;
      avg_decision_time_hours: number;
      total_projects: number;
      first_time_approval_pct: number;
      total_defects: number;
    }>("/v1/factory/scoreboard"),

  estimatePrice: (data: {
    title: string;
    description: string;
    product_type: string;
    complexity: string;
  }) =>
    apiClient<{
      suggested_price_cents: number;
      est_cost_cents: number;
      estimated_days: number;
      recommended_stack: string[];
      playbook: string;
    }>("/v1/factory/estimate", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  getPublicScope: (token: string) =>
    apiClient<{
      card_title: string;
      user_stories: Array<{
        title: string;
        given: string;
        when: string;
        then: string;
        passed: boolean;
      }>;
      tests_passed: boolean;
      client_approved: boolean;
      payment_signal_cents: number;
      payment_signal_paid: boolean;
    }>(`/v1/public/scope/${token}`),

  approvePublicScope: (token: string) =>
    apiClient<{ success: boolean; message: string }>(`/v1/public/scope/${token}/approve`, {
      method: "POST",
    }),
};

