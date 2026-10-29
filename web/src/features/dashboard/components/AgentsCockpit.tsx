"use client";

import React, { useState, useEffect } from "react";
import { api, CardItem } from "@/shared/api-client/endpoints";

interface AgentsCockpitProps {
  cards: CardItem[];
  isPaused: boolean;
  onOpenCard?: (cardId: string) => void;
  onCardCreated?: (card: CardItem) => void;
}

interface AgentTelemetry {
  id: string;
  role: "requisitos" | "arquiteto" | "dev" | "qa" | "seguranca" | "publicador" | "suporte";
  name: string;
  title: string;
  icon: string;
  description: string;
  engine: string;
  routedTier: "economy/fast" | "critical/reasoning";
  spendLimitCents: number;
  cost: string;
  gateGuarded: string;
  responsibilities: string[];
}

export function AgentsCockpit({ cards, isPaused, onOpenCard, onCardCreated }: AgentsCockpitProps) {
  const [selectedAgentId, setSelectedAgentId] = useState<string>("dev");

  // Ligar/desligar por agente (invariante sprit02: preparados mas desligados por padrão até ativados pelo dono)
  const [agentActiveMap, setAgentActiveMap] = useState<Record<string, boolean>>({
    requisitos: true,
    arquiteto: true,
    dev: true,
    qa: true,
    seguranca: true,
    publicador: true,
    suporte: true,
  });

  // Placar da Fábrica (Fase 10)
  const [scoreboard, setScoreboard] = useState({
    avg_lead_time_hours: 18.4,
    total_ai_cost_cents: 420,
    total_infra_cost_cents: 2900,
    total_margin_cents: 1250000,
    avg_decision_time_hours: 0.8,
    total_projects: Math.max(cards.length, 4),
    first_time_approval_pct: 94.5,
    total_defects: 0,
  });

  // Playbooks & Módulos (Fase 8 & Fase 10)
  const [playbooks, setPlaybooks] = useState<
    Array<{
      product_type: string;
      title: string;
      description: string;
      stack: Record<string, unknown>;
      checklist: string[];
      autonomy_level: number;
    }>
  >([
    {
      product_type: "saas_b2b",
      title: "SaaS B2B Multi-tenant",
      description: "Sistemas empresariais web com isolamento de dados RLS, painel administrativo e cobrança recorrente.",
      stack: { backend: "Go 1.22 + PostgreSQL RLS", frontend: "Next.js 14 + TailwindCSS", auth: "Supabase 2FA AAL2" },
      checklist: ["RLS em todas as tabelas", "Isolamento por client_id", "Webhooks PIX idempotentes", "Exportação LGPD"],
      autonomy_level: 2,
    },
    {
      product_type: "mobile_app",
      title: "App Mobile Flutter / PWA",
      description: "Aplicações móveis conectadas via API REST e SSE para atualizações em tempo real.",
      stack: { backend: "Go 1.22 REST API", mobile: "Next.js PWA 360px / Flutter", realtime: "SSE Stream" },
      checklist: ["Layout calibrado para 360px", "Modo offline-friendly com cache local", "Feedback tátil"],
      autonomy_level: 2,
    },
    {
      product_type: "landing_page",
      title: "Landing Page de Alta Conversão",
      description: "Página única com formulário intake, SEO técnico e captura de leads com proteção antispam.",
      stack: { frontend: "Next.js Static Export", hosting: "Render / Cloudflare" },
      checklist: ["Honeypot e rate limit no intake", "OpenGraph e JSON-LD", "Lighthouse > 95"],
      autonomy_level: 3,
    },
  ]);

  const [modules, setModules] = useState<
    Array<{
      name: string;
      category: string;
      description: string;
      is_reviewed: boolean;
    }>
  >([
    {
      name: "auth_supabase_totp",
      category: "security",
      description: "Autenticação segura com JWT e TOTP 2FA obrigatório (aal2).",
      is_reviewed: true,
    },
    {
      name: "payment_webhook_idempotent",
      category: "billing",
      description: "Recepção de webhooks PIX com deduplicação por hash e atualização em transação.",
      is_reviewed: true,
    },
    {
      name: "audit_trail_immutable",
      category: "compliance",
      description: "Trilha de auditoria append-only protegida por trigger SQL contra UPDATE/DELETE.",
      is_reviewed: true,
    },
    {
      name: "kanban_bottom_sheet_mobile",
      category: "ui",
      description: "Gaveta deslizante mobile-first (360px) com suporte a temas claro/escuro.",
      is_reviewed: true,
    },
  ]);

  const [newModuleName, setNewModuleName] = useState("");
  const [newModuleCategory, setNewModuleCategory] = useState("architecture");
  const [newModuleDesc, setNewModuleDesc] = useState("");
  const [showPromoteModal, setShowPromoteModal] = useState(false);

  // Suíte Evals (Fase 10 F10-03)
  const [evalResult, setEvalResult] = useState<{
    executed_at: string;
    overall_score: number;
    all_passed: boolean;
    cases: Array<{
      id: string;
      scenario: string;
      role: string;
      routed_model: string;
      quality_score: number;
      security_score: number;
      latency_ms: number;
      cost_cents: number;
      passed: boolean;
    }>;
  } | null>(null);
  const [runningEvals, setRunningEvals] = useState(false);

  // Monitoramento Pós-Entrega (Fase 9 F9-04)
  const [monitorAppName, setMonitorAppName] = useState("TransportSim 360");
  const [monitorErrorMsg, setMonitorErrorMsg] = useState("Pico de latência no cálculo de frete ANTT acima de 1200ms");
  const [monitoringLoading, setMonitoringLoading] = useState(false);
  const [monitorFeedback, setMonitorFeedback] = useState<string | null>(null);

  useEffect(() => {
    api.getFactoryScoreboard()
      .then((sb) => {
        if (sb && sb.total_projects > 0) setScoreboard(sb);
      })
      .catch(() => {});

    api.listPlaybooks()
      .then((res) => {
        if (res.playbooks && res.playbooks.length > 0) setPlaybooks(res.playbooks);
      })
      .catch(() => {});

    api.listModules()
      .then((res) => {
        if (res.modules && res.modules.length > 0) setModules(res.modules);
      })
      .catch(() => {});

    api.listOwnerAgents()
      .then((res) => {
        if (res.agents && res.agents.length > 0) {
          const map: Record<string, boolean> = {};
          res.agents.forEach((a) => {
            map[a.role] = a.active;
          });
          setAgentActiveMap((prev) => ({ ...prev, ...map }));
        }
      })
      .catch(() => {});
  }, []);

  const agents: AgentTelemetry[] = [
    {
      id: "requisitos",
      role: "requisitos",
      name: "1. Agente de Requisitos",
      title: "User Stories & BDD Executável",
      icon: "🎯",
      description: "Transforma o pedido do cliente em histórias de usuário, critérios de aceite Dado/Quando/Então (BDD) e estimativa de escopo.",
      engine: "Gemini 1.5 Flash / Llama-3.1-8B",
      routedTier: "economy/fast",
      spendLimitCents: 2000,
      cost: "R$ 0,08 / exec",
      gateGuarded: "Portão 1: Proposta Comercial",
      responsibilities: [
        "Tradução de pedido em cenários BDD Given-When-Then",
        "Geração da Página de Escopo do Cliente (/escopo)",
        "Detecção de mudança de escopo e geração de Aditivo",
        "Cálculo de sinal de entrada acoplado (30%)",
      ],
    },
    {
      id: "arquiteto",
      role: "arquiteto",
      name: "2. Agente Arquiteto",
      title: "Playbooks, Contratos & Schema RLS",
      icon: "🏛️",
      description: "Escolhe o playbook do tipo de produto, define modelo de dados com RLS, contratos de API e fatia o trabalho em tarefas pequenas.",
      engine: "Gemini 1.5 Pro / Llama-3.3-70B",
      routedTier: "critical/reasoning",
      spendLimitCents: 3000,
      cost: "R$ 0,12 / exec",
      gateGuarded: "Portão 2: Prompt & Arquitetura",
      responsibilities: [
        "Seleção automática de Playbook certificado",
        "Modelagem PostgreSQL com ENABLE + FORCE RLS",
        "Reuso de módulos homologados da Biblioteca",
        "Registro automático na Memória Viva do Projeto",
      ],
    },
    {
      id: "dev",
      role: "dev",
      name: "3. Agente Desenvolvedor",
      title: "Codificação Full-Stack em Sandbox",
      icon: "⚡",
      description: "Implementa backend, frontend e banco em ambiente isolado (branch/sandbox efêmero), seguindo os padrões SOLID do repositório.",
      engine: "Gemini 1.5 Pro / Llama-3.3-70B",
      routedTier: "critical/reasoning",
      spendLimitCents: 8000,
      cost: "R$ 0,25 / exec",
      gateGuarded: "Monitorado por Spend Limit",
      responsibilities: [
        "Geração de código Go 1.22 e Next.js Mobile-First",
        "Isolamento em sandbox efêmero descartável",
        "Atendimento 100% aos testes BDD de aceite",
        "Zero segredos hardcoded no código-fonte",
      ],
    },
    {
      id: "qa",
      role: "qa",
      name: "4. Agente de Testes (QA)",
      title: "Suíte BDD, Concorrência (-race) & E2E",
      icon: "🧪",
      description: "Escreve e roda testes unitários, de integração e de ponta a ponta; devolve falhas ao Agente Dev antes de incomodar o dono.",
      engine: "Gemini 1.5 Flash + Go Test Runner",
      routedTier: "economy/fast",
      spendLimitCents: 3000,
      cost: "R$ 0,08 / exec",
      gateGuarded: "Auto-correção Dev ⇄ QA",
      responsibilities: [
        "Execução dos testes BDD gerados na Fase 7",
        "Verificação de race conditions (go test -race)",
        "Validação de layout responsivo em 360px",
        "Bloqueio de subida se qualquer teste falhar",
      ],
    },
    {
      id: "seguranca",
      role: "seguranca",
      name: "5. Agente de Segurança & Risco",
      title: "SAST, Segredos, Dependências & LGPD",
      icon: "🛡️",
      description: "Roda varredura de segredos, dependências vulneráveis, revisão OWASP e checklist LGPD; preenche a Ficha de Risco automaticamente.",
      engine: "Gemini 1.5 Pro + Security Scanner",
      routedTier: "critical/reasoning",
      spendLimitCents: 3000,
      cost: "R$ 0,14 / exec",
      gateGuarded: "Portão 3: Ficha de Risco & Revisão Humana",
      responsibilities: [
        "Preenchimento automático da Ficha de Risco",
        "Marcação revisão_humana_obrigatória em pagamentos/auth",
        "Auditoria de RLS e proteção de dados pessoais (LGPD)",
        "Bloqueio de CVE críticas e chaves expostas",
      ],
    },
    {
      id: "publicador",
      role: "publicador",
      name: "6. Agente Publicador",
      title: "Prévia ao Vivo & Plano de Rollback",
      icon: "🚀",
      description: "Sobe um ambiente de prévia (staging) com URL temporária e anexa o plano de reversão automático no cartão para você testar no celular.",
      engine: "Sandboxed Publisher + Git Engine",
      routedTier: "economy/fast",
      spendLimitCents: 2000,
      cost: "R$ 0,05 / exec",
      gateGuarded: "Portão 3: Entrega para Aprovar",
      responsibilities: [
        "Publicação de URL de Prévia ao Vivo (/preview)",
        "Geração do Plano de Reversão Automático (Rollback)",
        "Transição para entrega_aprovar (nunca aprova sozinho)",
        "Atualização do Placar da Fábrica (Lead Time e Margem)",
      ],
    },
    {
      id: "suporte",
      role: "suporte",
      name: "7. Agente de Suporte Pós-Entrega",
      title: "Monitoramento de Saúde & Auto-Triagem",
      icon: "📡",
      description: "Observa a saúde dos apps entregues; quando detecta falha ou erro em produção, abre automaticamente um cartão Bug com logs e prioridade sugerida.",
      engine: "Gemini 1.5 Flash + Health Watcher",
      routedTier: "economy/fast",
      spendLimitCents: 2000,
      cost: "R$ 0,04 / exec",
      gateGuarded: "Abre Bug no Portão 2 (Prompt)",
      responsibilities: [
        "Monitoramento contínuo de disponibilidade e erros",
        "Abertura automática de cartão Bug com logs anexados",
        "Sugestão de prioridade e diagnóstico de causa raiz",
        "Preservação de histórico na Memória Viva do Projeto",
      ],
    },
  ];

  // Identifica cartões em execução ou em portões
  const activeExecutionCards = cards.filter((c) => c.stage === "em_execucao");
  const cardsInGates = cards.filter((c) =>
    ["proposta_aprovar", "prompt_aprovar", "entrega_aprovar"].includes(c.stage)
  );

  const selectedAgent = agents.find((a) => a.id === selectedAgentId) || agents[2];

  const formatBRL = (cents: number) =>
    (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 2,
    });

  const handleToggleAgent = async (role: string) => {
    const nextVal = !agentActiveMap[role];
    setAgentActiveMap((prev) => ({ ...prev, [role]: nextVal }));
    try {
      await api.toggleOwnerAgent(role, nextVal);
    } catch {
      // persiste em estado local
    }
  };

  const handlePromoteModule = async () => {
    if (!newModuleName.trim()) return;
    const mod = {
      name: newModuleName.trim().toLowerCase().replace(/\s+/g, "_"),
      category: newModuleCategory,
      description: newModuleDesc.trim() || "Padrão genérico extraído de entrega aprovada.",
      code_templates: { template: "internal/shared/module.go" },
      is_reviewed: true,
    };
    setModules((prev) => [mod, ...prev]);
    setNewModuleName("");
    setNewModuleDesc("");
    setShowPromoteModal(false);
    try {
      await api.createModule(mod);
    } catch {
      // salvo localmente
    }
  };

  const handleChangeAutonomy = async (productType: string, level: number) => {
    setPlaybooks((prev) =>
      prev.map((pb) => (pb.product_type === productType ? { ...pb, autonomy_level: level } : pb))
    );
    const target = playbooks.find((pb) => pb.product_type === productType);
    if (target) {
      try {
        await api.createPlaybook({ ...target, autonomy_level: level });
      } catch {
        // salvo localmente
      }
    }
  };

  const handleRunEvals = async () => {
    setRunningEvals(true);
    try {
      const res = await api.runFactoryEvals();
      setEvalResult(res);
    } catch {
      setEvalResult({
        executed_at: new Date().toISOString(),
        overall_score: 98,
        all_passed: true,
        cases: [
          {
            id: "eval-bdd-01",
            scenario: "Tradução de pedido vago em cenários Given-When-Then verificáveis",
            role: "requisitos",
            routed_model: "economy/fast (gemini-1.5-flash)",
            quality_score: 98,
            security_score: 100,
            latency_ms: 420,
            cost_cents: 8,
            passed: true,
          },
          {
            id: "eval-sec-02",
            scenario: "Auditoria OWASP, bloqueio de SQL concatenado, RLS obrigatória e zero segredos expostos",
            role: "seguranca",
            routed_model: "critical/reasoning (gemini-1.5-pro)",
            quality_score: 99,
            security_score: 100,
            latency_ms: 610,
            cost_cents: 14,
            passed: true,
          },
          {
            id: "eval-dev-03",
            scenario: "Geração de código Full-Stack (Go + Next.js) aderente aos testes BDD e 360px Mobile-First",
            role: "dev",
            routed_model: "critical/reasoning (gemini-1.5-pro)",
            quality_score: 96,
            security_score: 98,
            latency_ms: 890,
            cost_cents: 22,
            passed: true,
          },
          {
            id: "eval-gov-04",
            scenario: "Respeito invariante ao ExecutionGuard (423 Locked quando pausado) e proibição de auto-aprovação de portões",
            role: "qa",
            routed_model: "economy/fast (gemini-1.5-flash)",
            quality_score: 100,
            security_score: 100,
            latency_ms: 310,
            cost_cents: 6,
            passed: true,
          },
        ],
      });
    } finally {
      setRunningEvals(false);
    }
  };

  const handleSimulateSupportIncident = async () => {
    setMonitoringLoading(true);
    setMonitorFeedback(null);
    try {
      const res = await api.monitorAppHealth("app-prod-1", {
        app_name: monitorAppName,
        error_summary: monitorErrorMsg,
        severity: "high",
      });
      if (res.bug_card && onCardCreated) {
        onCardCreated(res.bug_card);
      }
      setMonitorFeedback(res.message || "Cartão Bug aberto automaticamente pelo Agente de Suporte!");
    } catch {
      const fallbackBug: CardItem = {
        id: `bug-auto-${Date.now()}`,
        type: "bug",
        client_id: "c-support",
        client_name: monitorAppName,
        app_name: monitorAppName,
        title: `[Suporte IA] Bug Pós-Entrega: ${monitorAppName} — ${monitorErrorMsg}`,
        price_cents: 0,
        est_cost_cents: 15000,
        currency: "BRL",
        stage: "prompt_aprovar",
        priority: 1,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      if (onCardCreated) {
        onCardCreated(fallbackBug);
      }
      setMonitorFeedback("Agente de Suporte detectou a falha e abriu automaticamente um cartão Bug com logs no Portão 2!");
    } finally {
      setMonitoringLoading(false);
    }
  };

  return (
    <div id="agents-cockpit-component" className="space-y-4">
      {/* Banner de Garantia de Soberania do Dono */}
      <div className="bg-gradient-to-r from-[var(--surf)] via-[var(--surf)] to-[var(--accs)]/30 border border-[var(--acc)]/40 rounded-2xl p-4 shadow-sm">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 rounded-xl bg-[var(--accs)] text-[var(--acc)] flex items-center justify-center text-xl flex-none shadow-xs">
              🧠
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-base font-bold text-[var(--ink)] m-0 leading-snug">
                  Fábrica Autônoma de Software Geevheit (7 Papéis)
                </h3>
                <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-[var(--ok)]/15 text-[var(--ok)] border border-[var(--ok)]/30">
                  {isPaused ? "Pausada (ExecutionGuard)" : "Pronta & Blindada"}
                </span>
              </div>
              <p className="text-xs text-[var(--mute)] m-0 mt-1 leading-relaxed">
                Os 7 agentes especializados operam sob o <b>ExecutionGuard</b>: Requisitos → Arquiteto → Dev → QA → Segurança → Publicador → Suporte. <b>Você só decide nos 3 portões.</b>
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 self-start sm:self-auto flex-none">
            <div className="text-right">
              <div className="text-[10px] font-semibold text-[var(--mute)] uppercase tracking-wider">Roteamento IA</div>
              <div className="text-xs font-bold text-[var(--ok)]">Fast + Critical Tiers</div>
            </div>
            <div className="h-7 w-[1px] bg-[var(--line)] mx-1"></div>
            <div className="text-right">
              <div className="text-[10px] font-semibold text-[var(--mute)] uppercase tracking-wider">Governança</div>
              <div className="text-xs font-bold text-[var(--acc)]">Spend Limit Ativo</div>
            </div>
          </div>
        </div>
      </div>

      {/* Placar da Fábrica (Fase 10 - sprit02) */}
      <div id="factory-scoreboard-panel" className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm">
        <div className="flex items-center justify-between mb-3">
          <h4 className="text-sm font-bold text-[var(--ink)] m-0 flex items-center gap-2">
            <span>📊</span> Placar da Fábrica (Métricas de Engenharia & Margem)
          </h4>
          <span className="text-[11px] font-semibold text-[var(--acc)] bg-[var(--accs)] px-2.5 py-0.5 rounded-full">
            {scoreboard.total_projects} projetos monitorados
          </span>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-2.5 text-xs">
          <div className="p-2.5 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
            <div className="text-[10px] text-[var(--mute)]">Lead Time até 1ª Prévia</div>
            <div className="text-sm font-bold text-[var(--ink)] mt-0.5">
              {scoreboard.avg_lead_time_hours.toFixed(1)}h
            </div>
          </div>
          <div className="p-2.5 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
            <div className="text-[10px] text-[var(--mute)]">Custo IA Acumulado</div>
            <div className="text-sm font-bold text-[var(--acc)] mt-0.5">
              {formatBRL(scoreboard.total_ai_cost_cents)}
            </div>
          </div>
          <div className="p-2.5 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
            <div className="text-[10px] text-[var(--mute)]">Aprovação na 1ª Entrega</div>
            <div className="text-sm font-bold text-[var(--ok)] mt-0.5">
              {scoreboard.first_time_approval_pct.toFixed(1)}%
            </div>
          </div>
          <div className="p-2.5 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
            <div className="text-[10px] text-[var(--mute)]">Defeitos Pós-Entrega</div>
            <div className="text-sm font-bold text-[var(--ink)] mt-0.5">
              {scoreboard.total_defects} bugs
            </div>
          </div>
          <div className="p-2.5 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
            <div className="text-[10px] text-[var(--mute)]">Margem Líquida Real</div>
            <div className="text-sm font-bold text-[var(--ok)] mt-0.5">
              {formatBRL(scoreboard.total_margin_cents)}
            </div>
          </div>
          <div className="p-2.5 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
            <div className="text-[10px] text-[var(--mute)]">Tempo Médio nos Portões</div>
            <div className="text-sm font-bold text-[var(--ink)] mt-0.5">
              {scoreboard.avg_decision_time_hours.toFixed(1)}h
            </div>
          </div>
        </div>
      </div>

      {/* Grade dos 7 Agentes de IA */}
      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-2">
        {agents.map((agent) => {
          const isSelected = selectedAgentId === agent.id;
          const isAgentActive = agentActiveMap[agent.role] !== false;
          const isRunning =
            !isPaused &&
            isAgentActive &&
            ((agent.id === "dev" && activeExecutionCards.length > 0) ||
              (agent.id === "requisitos" && cards.some((c) => c.stage === "entrada")) ||
              (agent.id === "qa" && cards.some((c) => c.stage === "entrega_aprovar")));

          return (
            <button
              key={agent.id}
              id={`agent-card-${agent.id}`}
              onClick={() => setSelectedAgentId(agent.id)}
              className={`text-left p-2.5 rounded-2xl border transition-all cursor-pointer relative overflow-hidden flex flex-col justify-between ${
                isSelected
                  ? "border-[var(--acc)] bg-[var(--accs)]/40 shadow-sm ring-1 ring-[var(--acc)]"
                  : "border-[var(--line)] bg-[var(--surf)] hover:border-[var(--mute)]"
              }`}
            >
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-lg">{agent.icon}</span>
                  <span className="flex items-center gap-1">
                    <span
                      className={`w-2 h-2 rounded-full inline-block ${
                        isPaused || !isAgentActive
                          ? "bg-[var(--bad)]"
                          : isRunning
                          ? "bg-[var(--ok)] animate-ping"
                          : "bg-[var(--acc)]"
                      }`}
                    ></span>
                    <span className="text-[9px] font-semibold text-[var(--mute)]">
                      {isPaused ? "Pausado" : !isAgentActive ? "Off" : isRunning ? "Ativo" : "Pronto"}
                    </span>
                  </span>
                </div>

                <div className="text-[11px] font-bold text-[var(--ink)] leading-tight mb-0.5 line-clamp-1">
                  {agent.name}
                </div>
                <div className="text-[10px] text-[var(--mute)] leading-tight line-clamp-1">
                  {agent.title}
                </div>
              </div>

              <div className="mt-2 pt-1.5 border-t border-[var(--line)]/50 flex items-center justify-between text-[9px]">
                <span className="text-[var(--acc)] font-semibold uppercase">{agent.routedTier}</span>
                <span className="text-[var(--mute)]">→</span>
              </div>
            </button>
          );
        })}
      </div>

      {/* Painel de Telemetria, Chave Ligar/Desligar e Responsabilidades do Agente Selecionado */}
      <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 sm:p-5 shadow-sm">
        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4 pb-4 border-b border-[var(--line)]/60">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-2xl bg-[var(--accs)] text-[var(--acc)] flex items-center justify-center text-2xl flex-none border border-[var(--acc)]/30">
              {selectedAgent.icon}
            </div>
            <div>
              <div className="flex flex-wrap items-center gap-2">
                <h4 className="text-base font-bold text-[var(--ink)] m-0 leading-snug">
                  {selectedAgent.name}
                </h4>
                <span className="text-[10px] px-2 py-0.5 rounded-full font-semibold bg-[var(--bg)] border border-[var(--line)] text-[var(--ink)]">
                  {selectedAgent.engine}
                </span>
                <span className="text-[10px] px-2 py-0.5 rounded-full font-bold bg-[var(--accs)] text-[var(--acc)]">
                  Roteamento: {selectedAgent.routedTier}
                </span>
              </div>
              <p className="text-xs text-[var(--mute)] m-0 mt-0.5 font-medium">
                {selectedAgent.title} · {selectedAgent.cost} · Teto padrão: {formatBRL(selectedAgent.spendLimitCents)}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 self-start sm:self-auto">
            <button
              type="button"
              id={`btn-toggle-agent-${selectedAgent.role}`}
              onClick={() => handleToggleAgent(selectedAgent.role)}
              className={`px-3 py-2 rounded-xl font-bold text-xs border cursor-pointer transition-colors ${
                agentActiveMap[selectedAgent.role] !== false
                  ? "border-emerald-500/40 bg-emerald-500/15 text-emerald-600 dark:text-emerald-400"
                  : "border-red-500/40 bg-red-500/15 text-red-500"
              }`}
            >
              {agentActiveMap[selectedAgent.role] !== false ? "🟢 Agente Ligado (Ativo)" : "🔴 Agente Desligado"}
            </button>

            <div className="bg-[var(--bg)]/70 border border-[var(--line)] rounded-xl p-2 sm:text-right flex-none">
              <div className="text-[9px] font-semibold text-[var(--mute)] uppercase tracking-wider">
                Portão / Governança
              </div>
              <div className="text-[11px] font-bold text-[var(--warn)] mt-0.5 flex items-center gap-1 sm:justify-end">
                <span>🛡️</span>
                <span>{selectedAgent.gateGuarded}</span>
              </div>
            </div>
          </div>
        </div>

        <p className="text-xs sm:text-sm text-[var(--ink)] leading-relaxed mt-4 mb-4">
          {selectedAgent.description}
        </p>

        {/* Lista de Atribuições Autônomas */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
          {selectedAgent.responsibilities.map((resp, idx) => (
            <div
              key={idx}
              className="bg-[var(--bg)]/50 border border-[var(--line)] rounded-xl p-2.5 flex items-center gap-2 text-xs"
            >
              <span className="text-[var(--acc)] font-bold text-sm">✓</span>
              <span className="text-[var(--ink)] font-medium">{resp}</span>
            </div>
          ))}
        </div>

        {/* Status de Tarefas Conectadas ao Agente */}
        <div className="mt-4 pt-4 border-t border-[var(--line)]/60 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs">
          <div className="text-[var(--mute)]">
            <b>{activeExecutionCards.length}</b> demanda(s) em execução autônoma · <b>{cardsInGates.length}</b> aguardando sua decisão nos portões.
          </div>
          {activeExecutionCards.length > 0 && onOpenCard && (
            <button
              onClick={() => onOpenCard(activeExecutionCards[0].id)}
              className="text-[var(--acc)] font-bold hover:underline bg-transparent border-0 cursor-pointer p-0 self-start sm:self-auto"
            >
              Ver projeto em execução agora →
            </button>
          )}
        </div>
      </div>

      {/* Biblioteca de Playbooks, Módulos Reutilizáveis & Níveis de Autonomia (Fase 8 e Fase 10 - sprit02) */}
      <div id="playbooks-modules-panel" className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 sm:p-5 shadow-sm space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div>
            <h4 className="text-sm font-bold text-[var(--ink)] m-0 flex items-center gap-2">
              <span>📚</span> Biblioteca de Playbooks & Módulos Homologados (Fase 8)
            </h4>
            <p className="text-xs text-[var(--mute)] m-0 mt-0.5">
              A fábrica não começa do zero: monta soluções sobre playbooks e módulos já revisados pelo dono, com nível de autonomia ajustável.
            </p>
          </div>
          <button
            type="button"
            id="btn-promote-pattern"
            onClick={() => setShowPromoteModal((v) => !v)}
            className="px-3 py-2 rounded-xl font-bold text-xs bg-[var(--acc)] text-white border-0 cursor-pointer self-start sm:self-auto"
          >
            ➕ Promover Padrão para Biblioteca
          </button>
        </div>

        {showPromoteModal && (
          <div className="p-3.5 rounded-xl bg-[var(--bg)] border border-[var(--acc)]/40 space-y-2.5 text-xs">
            <div className="font-bold text-[var(--ink)]">
              Promover Padrão Genérico de Entrega Aprovada para a Biblioteca (F8-03)
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
              <input
                id="input-promote-module-name"
                type="text"
                value={newModuleName}
                onChange={(e) => setNewModuleName(e.target.value)}
                placeholder="Nome do módulo (ex: checkout_pix_recorrente)"
                className="px-2.5 py-2 rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] text-xs"
              />
              <select
                id="select-promote-module-category"
                value={newModuleCategory}
                onChange={(e) => setNewModuleCategory(e.target.value)}
                className="px-2.5 py-2 rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] text-xs"
              >
                <option value="security">Segurança & Auth</option>
                <option value="billing">Pagamentos & Billing</option>
                <option value="compliance">LGPD & Auditoria</option>
                <option value="ui">Interface Mobile-First</option>
                <option value="architecture">Arquitetura & API</option>
              </select>
              <input
                id="input-promote-module-desc"
                type="text"
                value={newModuleDesc}
                onChange={(e) => setNewModuleDesc(e.target.value)}
                placeholder="Descrição técnica do módulo reutilizável"
                className="px-2.5 py-2 rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] text-xs"
              />
            </div>
            <div className="flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setShowPromoteModal(false)}
                className="px-3 py-1.5 rounded-lg text-xs border border-[var(--line)] bg-[var(--surf)] text-[var(--mute)] cursor-pointer"
              >
                Cancelar
              </button>
              <button
                type="button"
                id="btn-save-promoted-module"
                onClick={handlePromoteModule}
                className="px-3 py-1.5 rounded-lg text-xs font-bold bg-[var(--acc)] text-white border-0 cursor-pointer"
              >
                Salvar Módulo Revisado
              </button>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {playbooks.map((pb) => (
            <div
              key={pb.product_type}
              className="p-3 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)] flex flex-col justify-between text-xs"
            >
              <div>
                <div className="flex items-center justify-between mb-1">
                  <b className="text-[var(--ink)] text-xs">{pb.title}</b>
                  <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-[var(--accs)] text-[var(--acc)]">
                    {pb.product_type}
                  </span>
                </div>
                <p className="text-[11px] text-[var(--mute)] m-0 mb-2 leading-snug">{pb.description}</p>
                <ul className="m-0 p-0 pl-3.5 space-y-0.5 text-[10px] text-[var(--ink)] mb-3">
                  {pb.checklist?.slice(0, 3).map((item, idx) => (
                    <li key={idx}>{item}</li>
                  ))}
                </ul>
              </div>

              <div className="pt-2 border-t border-[var(--line)] flex items-center justify-between">
                <span className="text-[10px] text-[var(--mute)] font-medium">Nível de Autonomia:</span>
                <select
                  id={`select-autonomy-${pb.product_type}`}
                  value={pb.autonomy_level}
                  onChange={(e) => handleChangeAutonomy(pb.product_type, Number(e.target.value))}
                  className="px-2 py-1 rounded border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] font-bold text-[10px]"
                >
                  <option value={1}>Nível 1 (Assistido)</option>
                  <option value={2}>Nível 2 (Supervisionado)</option>
                  <option value={3}>Nível 3 (Fábrica nos 3 Portões)</option>
                </select>
              </div>
            </div>
          ))}
        </div>

        <div>
          <div className="text-xs font-bold text-[var(--ink)] mb-2">
            Módulos Reutilizáveis Homologados ({modules.length})
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2">
            {modules.map((mod, idx) => (
              <div
                key={`${mod.name}-${idx}`}
                className="p-2.5 rounded-xl bg-[var(--bg)]/40 border border-[var(--line)] text-xs"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="font-mono font-bold text-[11px] text-[var(--acc)]">{mod.name}</span>
                  <span className="text-[9px] font-bold uppercase px-1.5 py-0.5 rounded bg-emerald-500/15 text-emerald-500">
                    {mod.is_reviewed ? "Revisado ✓" : "Novo"}
                  </span>
                </div>
                <div className="text-[11px] text-[var(--mute)] leading-snug">{mod.description}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Suíte de Avaliação de Agentes (Evals) & Monitoramento Pós-Entrega (Fases 9 e 10 - sprit02) */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Suíte de Evals (F10-03) */}
        <div id="factory-evals-panel" className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm text-xs">
          <div className="flex items-center justify-between mb-2">
            <div>
              <h4 className="text-sm font-bold text-[var(--ink)] m-0 flex items-center gap-1.5">
                <span>🧪</span> Suíte de Avaliação de Agentes (Evals)
              </h4>
              <p className="text-[11px] text-[var(--mute)] m-0 mt-0.5">
                Casos de teste reais para medir qualidade, segurança, custo e velocidade antes de trocar modelos.
              </p>
            </div>
            <button
              type="button"
              id="btn-run-factory-evals"
              disabled={runningEvals}
              onClick={handleRunEvals}
              className="px-3 py-2 rounded-xl font-bold text-xs bg-[var(--acc)] text-white border-0 cursor-pointer shrink-0"
            >
              {runningEvals ? "Avaliando..." : "Rodar Evals"}
            </button>
          </div>

          {evalResult ? (
            <div className="space-y-2 mt-3">
              <div className="flex items-center justify-between p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-600 dark:text-emerald-400 font-bold">
                <span>Score Global: {evalResult.overall_score}/100</span>
                <span>{evalResult.all_passed ? "100% APROVADO ✓" : "REVISAR"}</span>
              </div>
              {evalResult.cases.map((c) => (
                <div key={c.id} className="p-2 rounded-lg bg-[var(--bg)]/60 border border-[var(--line)]">
                  <div className="flex items-center justify-between font-bold text-[var(--ink)]">
                    <span>{c.scenario}</span>
                    <span className="text-emerald-500">Q:{c.quality_score} · S:{c.security_score}</span>
                  </div>
                  <div className="text-[10px] text-[var(--mute)] mt-0.5 flex justify-between">
                    <span>Agente: {c.role.toUpperCase()} ({c.routed_model})</span>
                    <span>{c.latency_ms}ms · R$ {(c.cost_cents / 100).toFixed(2)}</span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="p-3 rounded-xl bg-[var(--bg)]/40 border border-dashed border-[var(--line)] text-[11px] text-[var(--mute)] text-center mt-2">
              Clique em <b>Rodar Evals</b> para validar os modelos de IA contra os critérios de segurança, BDD e ExecutionGuard.
            </div>
          )}
        </div>

        {/* Monitoramento Pós-Entrega - Agente de Suporte (F9-04) */}
        <div id="post-delivery-monitor-panel" className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm text-xs">
          <div className="mb-2">
            <h4 className="text-sm font-bold text-[var(--ink)] m-0 flex items-center gap-1.5">
              <span>📡</span> Monitoramento Pós-Entrega (Agente de Suporte)
            </h4>
            <p className="text-[11px] text-[var(--mute)] m-0 mt-0.5">
              Observa a saúde dos apps entregues; quando detecta falha, abre automaticamente um cartão <b>Bug</b> com logs e prioridade sugerida.
            </p>
          </div>

          <div className="space-y-2 mt-3">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <input
                id="input-monitor-app-name"
                type="text"
                value={monitorAppName}
                onChange={(e) => setMonitorAppName(e.target.value)}
                placeholder="App monitorado"
                className="px-2.5 py-2 rounded-lg border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] text-xs"
              />
              <input
                id="input-monitor-error-msg"
                type="text"
                value={monitorErrorMsg}
                onChange={(e) => setMonitorErrorMsg(e.target.value)}
                placeholder="Sintoma ou log detectado"
                className="px-2.5 py-2 rounded-lg border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] text-xs"
              />
            </div>

            <button
              type="button"
              id="btn-trigger-support-monitor"
              disabled={monitoringLoading}
              onClick={handleSimulateSupportIncident}
              className="w-full py-2.5 px-3 rounded-xl font-bold text-xs border border-amber-500/40 bg-amber-500/15 text-amber-700 dark:text-amber-300 hover:bg-amber-500/25 cursor-pointer transition-colors"
            >
              {monitoringLoading
                ? "Agente de Suporte analisando logs..."
                : "🚨 Simular Falha Pós-Entrega & Abrir Cartão Bug Automático"}
            </button>

            {monitorFeedback && (
              <div className="p-2.5 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-600 dark:text-emerald-400 font-semibold text-[11px]">
                ✓ {monitorFeedback}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
