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

  // Sprint 03: Motor de Cobrança PIX (Fase 11)
  generatePix: (cardId: string, pixKey?: string) =>
    apiClient<{
      id: string;
      card_id: string;
      client_id: string;
      amount_cents: number;
      currency: string;
      pix_emv: string;
      pix_txid: string;
      pix_key: string;
      status: string;
      paid_at?: string;
      created_at: string;
    }>(`/v1/cards/${cardId}/pix`, {
      method: "POST",
      body: JSON.stringify({ pix_key: pixKey }),
    }),

  getPix: (cardId: string) =>
    apiClient<{
      id: string;
      card_id: string;
      client_id: string;
      amount_cents: number;
      currency: string;
      pix_emv: string;
      pix_txid: string;
      pix_key: string;
      status: string;
      paid_at?: string;
      created_at: string;
    }>(`/v1/cards/${cardId}/pix`),

  // Sprint 03: Propostas Comerciais & Contratos (Fase 12)
  getProposalDocument: (cardId: string) =>
    apiClient<ProposalDocument>(`/v1/cards/${cardId}/proposal-document`),

  generateProposalDocument: (cardId: string) =>
    apiClient<ProposalDocument>(`/v1/cards/${cardId}/proposal-document`, {
      method: "POST",
    }),

  // Sprint 03: Portal Exclusivo do Cliente (Fase 13)
  generatePortalLink: (cardId: string, daysValid: number = 30) =>
    apiClient<{
      token: string;
      portal_url: string;
      card_id: string;
      days_valid: number;
    }>(`/v1/cards/${cardId}/portal-link`, {
      method: "POST",
      body: JSON.stringify({ days_valid: daysValid }),
    }),

  getPortalStatus: (cardId: string) =>
    apiClient<PortalStatusResponse>(`/v1/cards/${cardId}/portal-link`),

  revokePortalLink: (cardId: string) =>
    apiClient<{ message: string }>(`/v1/cards/${cardId}/portal-link`, {
      method: "DELETE",
    }),

  // Subscriptions & MRR (Fase 14)
  getMRRMetrics: () => apiClient<MRRMetrics>("/v1/finances/mrr"),
  getSubscriptions: (params?: { client_id?: string; status?: string }) => {
    const q = new URLSearchParams();
    if (params?.client_id) q.set("client_id", params.client_id);
    if (params?.status) q.set("status", params.status);
    const qs = q.toString() ? `?${q.toString()}` : "";
    return apiClient<SubscriptionItem[]>(`/v1/subscriptions${qs}`);
  },
  createSubscription: (data: CreateSubscriptionInput) =>
    apiClient<SubscriptionItem>("/v1/subscriptions", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateSubscriptionStatus: (id: string, status: string) =>
    apiClient<SubscriptionItem>(`/v1/subscriptions/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    }),
  getSubscriptionCharges: (params?: { status?: string; due_before?: string }) => {
    const q = new URLSearchParams();
    if (params?.status) q.set("status", params.status);
    if (params?.due_before) q.set("due_before", params.due_before);
    const qs = q.toString() ? `?${q.toString()}` : "";
    return apiClient<SubscriptionChargeItem[]>(`/v1/finances/charges${qs}`);
  },
  paySubscriptionCharge: (id: string, paymentMethod = "pix") =>
    apiClient<SubscriptionChargeItem>(`/v1/finances/charges/${id}/pay`, {
      method: "POST",
      body: JSON.stringify({ payment_method: paymentMethod }),
    }),
  processDailySubscriptions: () =>
    apiClient<{ overdue_count: number; generated_count: number; processed_at: string }>(
      "/v1/subscriptions/process-daily",
      { method: "POST" }
    ),

  // Webhooks & Notificações (Fase 15)
  listWebhooks: () => apiClient<WebhookEndpoint[]>("/v1/webhooks"),
  createWebhook: (data: { url: string; description: string; events: string[] }) =>
    apiClient<WebhookEndpoint>("/v1/webhooks", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  deleteWebhook: (id: string) =>
    apiClient<{ message: string }>(`/v1/webhooks/${id}`, { method: "DELETE" }),
  listWebhookDeliveries: (webhookId?: string) => {
    const qs = webhookId ? `?webhook_id=${webhookId}` : "";
    return apiClient<WebhookDelivery[]>(`/v1/webhooks/deliveries${qs}`);
  },
  redeliverWebhook: (id: string, event: string, payload: Record<string, unknown>) =>
    apiClient<WebhookDelivery>(`/v1/webhooks/${id}/redeliver`, {
      method: "POST",
      body: JSON.stringify({ event, payload }),
    }),
  getNotificationPreferences: () =>
    apiClient<NotificationPreferences>("/v1/notifications/preferences"),
  saveNotificationPreferences: (prefs: NotificationPreferences) =>
    apiClient<NotificationPreferences>("/v1/notifications/preferences", {
      method: "PUT",
      body: JSON.stringify(prefs),
    }),
};

export interface ProposalDocument {
  id: string;
  card_id: string;
  title: string;
  client_name: string;
  app_name: string;
  scope_summary: string;
  deliverables: string[];
  timeline_days: number;
  warranty_days: number;
  price_cents: number;
  currency: string;
  payment_terms: string;
  accept_url?: string;
  pix_emv?: string;
  pix_txid?: string;
  formatted_whatsapp: string;
  terms_and_conditions: string;
  generated_at: string;
}

export interface PortalStageItem {
  key: string;
  label: string;
  description: string;
  status: "concluido" | "em_andamento" | "pendente";
  order: number;
  updated_at?: string;
}

export interface ClientPortalData {
  project_title: string;
  client_name: string;
  app_name: string;
  current_stage_label: string;
  progress_percentage: number;
  stages: PortalStageItem[];
  can_accept: boolean;
  is_accepted: boolean;
  accepted_at?: string;
  accepted_by?: string;
}

export interface PortalStatusResponse {
  card_id: string;
  is_active: boolean;
  portal_url?: string;
  expires_at?: string;
  last_accessed_at?: string;
  accepted_at?: string;
  accepted_by?: string;
}

export interface SubscriptionItem {
  id: string;
  owner_id: string;
  client_id: string;
  app_id?: string;
  plan: string;
  description: string;
  amount_cents: number;
  currency: string;
  interval: "mensal" | "trimestral" | "semestral" | "anual";
  due_day: number;
  status: "ativa" | "pausada" | "cancelada" | "inadimplente";
  start_date: string;
  end_date?: string;
  next_due_date: string;
  created_at: string;
  updated_at: string;
}

export interface SubscriptionChargeItem {
  id: string;
  owner_id: string;
  subscription_id: string;
  amount_cents: number;
  currency: string;
  due_date: string;
  status: "pendente" | "paga" | "atrasada";
  paid_at?: string;
  payment_method: string;
  created_at: string;
  updated_at: string;
}

export interface MRRMetrics {
  current_mrr_cents: number;
  projected_mrr_cents: number;
  active_count: number;
  paused_count: number;
  canceled_count: number;
  delinquent_count: number;
  delinquent_amount_cents: number;
}

export interface CreateSubscriptionInput {
  client_id: string;
  app_id?: string;
  plan: string;
  description?: string;
  amount_cents: number;
  currency?: string;
  interval: "mensal" | "trimestral" | "semestral" | "anual";
  due_day: number;
}

export interface WebhookEndpoint {
  id: string;
  owner_id: string;
  url: string;
  description: string;
  secret_token?: string;
  events: string[];
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface WebhookDelivery {
  id: string;
  webhook_id: string;
  event: string;
  payload: Record<string, unknown>;
  status_code: number;
  attempts: number;
  response_body: string;
  success: boolean;
  created_at: string;
}

export interface NotificationPreferences {
  owner_id?: string;
  channels: {
    telegram_enabled: boolean;
    whatsapp_enabled: boolean;
    email_enabled: boolean;
  };
  events: {
    intake_enabled: boolean;
    pix_enabled: boolean;
    portal_accept_enabled: boolean;
    subscription_overdue_enabled: boolean;
  };
}




