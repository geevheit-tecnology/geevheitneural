"use client";

import React, { useState } from "react";
import { CardItem } from "@/shared/api-client/endpoints";

interface AgentsCockpitProps {
  cards: CardItem[];
  isPaused: boolean;
  onOpenCard?: (cardId: string) => void;
}

interface AgentTelemetry {
  id: string;
  role: "requisitos" | "arquiteto" | "dev" | "qa" | "publicador";
  name: string;
  title: string;
  icon: string;
  description: string;
  engine: string;
  cost: string;
  gateGuarded: string;
  responsibilities: string[];
}

export function AgentsCockpit({ cards, isPaused, onOpenCard }: AgentsCockpitProps) {
  const [selectedAgentId, setSelectedAgentId] = useState<string>("dev");

  const agents: AgentTelemetry[] = [
    {
      id: "requisitos",
      role: "requisitos",
      name: "Agente Comercial & PO",
      title: "Especificação & Escopo de Negócio",
      icon: "🎯",
      description: "Interpreta solicitações de clientes, gera user stories em formato BDD e calcula orçamento ideal com margem de lucro prevista.",
      engine: "Gemini 1.5 Flash",
      cost: "R$ 0,00 (Free Tier)",
      gateGuarded: "Portão 1: Proposta Comercial",
      responsibilities: [
        "Captação via intake inteligente",
        "Estimativa de complexidade e prazo",
        "Geração de proposta formal com link de aceite",
        "Detecção de requisitos ambíguos",
      ],
    },
    {
      id: "arquiteto",
      role: "arquiteto",
      name: "Agente Arquiteto de Software",
      title: "Clean Architecture & Design System",
      icon: "🏛️",
      description: "Desenha a estrutura de pastas, contratos de API (OpenAPI), schema de banco com RLS forçada e seleciona os módulos reutilizáveis.",
      engine: "Gemini 1.5 Flash",
      cost: "R$ 0,00 (Free Tier)",
      gateGuarded: "Portão 2: Prompt de Execução",
      responsibilities: [
        "Definição de modelo de dados relacional",
        "Garantia dos princípios SOLID e Clean Arch",
        "Design de rotas e políticas de segurança RLS",
        "Definição de dependências sem custos fixos",
      ],
    },
    {
      id: "dev",
      role: "dev",
      name: "Agente Engenheiro Fullstack",
      title: "Codificação Autônoma no Sandbox",
      icon: "⚡",
      description: "Codifica endpoints em Go e telas em Next.js dentro de um diretório isolado (sandbox efêmero anti-path traversal).",
      engine: "Gemini 1.5 Flash",
      cost: "R$ 0,00 (Free Tier)",
      gateGuarded: "Monitorado por Spend Limit",
      responsibilities: [
        "Construção de código fonte Go e TypeScript",
        "Isolamento em sandbox local seguro",
        "Atendimento estrito às user stories do PO",
        "Execução orientada por playbooks certificados",
      ],
    },
    {
      id: "qa",
      role: "qa",
      name: "Agente QA & Auditor de Risco",
      title: "Testes Automáticos, Segredos & LGPD",
      icon: "🛡️",
      description: "Executa testes unitários com detecção de concorrência (-race), varredura de credenciais e conformidade com privacidade LGPD.",
      engine: "Gemini 1.5 Flash + Go Test Suite",
      cost: "R$ 0,00 (Free Tier)",
      gateGuarded: "Portão 3: Ficha de Risco para Entrega",
      responsibilities: [
        "Execução de suíte de testes com flag -race",
        "Detecção de chaves/senhas expostas em código",
        "Validação de conformidade com LGPD",
        "Preenchimento automático da Ficha de Risco",
      ],
    },
    {
      id: "publicador",
      role: "publicador",
      name: "Agente DevOps & Publisher",
      title: "Git Atômico & Sandbox de Prévia",
      icon: "🚀",
      description: "Realiza commits versionados com Conventional Commits, cria repositório e disponibiliza a prévia ao vivo para teste.",
      engine: "Git Local Engine + Sandboxed Server",
      cost: "R$ 0,00 (Free Tier)",
      gateGuarded: "Liberação exclusiva sob 2FA do Dono",
      responsibilities: [
        "Versionamento local atômico no Git",
        "Geração de endpoint de preview (/v1/cards/{id}/preview)",
        "Isolamento seguro de portas e processos",
        "Empacotamento estático para Vercel/Render",
      ],
    },
  ];

  // Identifica cartões em execução ou em portões
  const activeExecutionCards = cards.filter((c) => c.stage === "em_execucao");
  const cardsInGates = cards.filter((c) =>
    ["proposta_aprovar", "prompt_aprovar", "entrega_aprovar"].includes(c.stage)
  );

  const selectedAgent = agents.find((a) => a.id === selectedAgentId) || agents[2];

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
                  Fábrica Autônoma de Software Geevheit
                </h3>
                <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-[var(--ok)]/15 text-[var(--ok)] border border-[var(--ok)]/30">
                  {isPaused ? "Pausada" : "Pronta & Blindada"}
                </span>
              </div>
              <p className="text-xs text-[var(--mute)] m-0 mt-1 leading-relaxed">
                Os 5 agentes de IA operam sob o <b>ExecutionGuard</b>: concebem, codificam e testam. <b>Você só decide e aprova nos portões.</b> Nenhuma linha vai para produção sem seu comando.
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 self-start sm:self-auto flex-none">
            <div className="text-right">
              <div className="text-[10px] font-semibold text-[var(--mute)] uppercase tracking-wider">Custo de Operação</div>
              <div className="text-xs font-bold text-[var(--ok)]">R$ 0,00 / Free Tier</div>
            </div>
            <div className="h-7 w-[1px] bg-[var(--line)] mx-1"></div>
            <div className="text-right">
              <div className="text-[10px] font-semibold text-[var(--mute)] uppercase tracking-wider">Taxa de Execução</div>
              <div className="text-xs font-bold text-[var(--acc)]">15 req/min (Protegido)</div>
            </div>
          </div>
        </div>
      </div>

      {/* Grade dos 5 Agentes de IA */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-2.5">
        {agents.map((agent) => {
          const isSelected = selectedAgentId === agent.id;
          const isRunning = !isPaused && (
            (agent.id === "dev" && activeExecutionCards.length > 0) ||
            (agent.id === "requisitos" && cards.some((c) => c.stage === "entrada")) ||
            (agent.id === "qa" && cards.some((c) => c.stage === "entrega_aprovar"))
          );

          return (
            <button
              key={agent.id}
              id={`agent-card-${agent.id}`}
              onClick={() => setSelectedAgentId(agent.id)}
              className={`text-left p-3 rounded-2xl border transition-all cursor-pointer relative overflow-hidden flex flex-col justify-between ${
                isSelected
                  ? "border-[var(--acc)] bg-[var(--accs)]/40 shadow-sm ring-1 ring-[var(--acc)]"
                  : "border-[var(--line)] bg-[var(--surf)] hover:border-[var(--mute)]"
              }`}
            >
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xl">{agent.icon}</span>
                  <span className="flex items-center gap-1">
                    <span
                      className={`w-2 h-2 rounded-full inline-block ${
                        isPaused
                          ? "bg-[var(--bad)]"
                          : isRunning
                          ? "bg-[var(--ok)] animate-ping"
                          : "bg-[var(--acc)]"
                      }`}
                    ></span>
                    <span className="text-[10px] font-semibold text-[var(--mute)]">
                      {isPaused ? "Parado" : isRunning ? "Ativo" : "Standby"}
                    </span>
                  </span>
                </div>

                <div className="text-xs font-bold text-[var(--ink)] leading-tight mb-0.5 line-clamp-1">
                  {agent.name}
                </div>
                <div className="text-[10px] text-[var(--mute)] leading-tight line-clamp-1">
                  {agent.title}
                </div>
              </div>

              <div className="mt-3 pt-2 border-t border-[var(--line)]/50 flex items-center justify-between text-[10px]">
                <span className="text-[var(--acc)] font-medium">Ver detalhes</span>
                <span className="text-[var(--mute)]">→</span>
              </div>
            </button>
          );
        })}
      </div>

      {/* Painel de Telemetria e Responsabilidades do Agente Selecionado */}
      <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 sm:p-5 shadow-sm">
        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4 pb-4 border-b border-[var(--line)]/60">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-2xl bg-[var(--accs)] text-[var(--acc)] flex items-center justify-center text-2xl flex-none border border-[var(--acc)]/30">
              {selectedAgent.icon}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h4 className="text-base font-bold text-[var(--ink)] m-0 leading-snug">
                  {selectedAgent.name}
                </h4>
                <span className="text-[11px] px-2.5 py-0.5 rounded-full font-semibold bg-[var(--bg)] border border-[var(--line)] text-[var(--ink)]">
                  {selectedAgent.engine}
                </span>
              </div>
              <p className="text-xs text-[var(--mute)] m-0 mt-0.5 font-medium">
                {selectedAgent.title} · {selectedAgent.cost}
              </p>
            </div>
          </div>

          <div className="bg-[var(--bg)]/70 border border-[var(--line)] rounded-xl p-2.5 sm:text-right flex-none">
            <div className="text-[10px] font-semibold text-[var(--mute)] uppercase tracking-wider">
              Blindagem de Segurança
            </div>
            <div className="text-xs font-bold text-[var(--warn)] mt-0.5 flex items-center gap-1 sm:justify-end">
              <span>🛡️</span>
              <span>{selectedAgent.gateGuarded}</span>
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
    </div>
  );
}
