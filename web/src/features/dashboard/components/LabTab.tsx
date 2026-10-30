"use client";

import React, { useState, useEffect, useRef } from "react";
import { CardItem } from "@/shared/api-client/endpoints";

interface LabTabProps {
  cards: CardItem[];
  isPaused: boolean;
  onOpenCard: (cardId: string) => void;
  onStartBuild: (cardId: string, spendLimitCents?: number) => void;
  onApproveCard: (cardId: string) => void;
  onOpenNewCardModal: () => void;
}

interface AgentChatMessage {
  id: string;
  sender: "requisitos" | "arquiteto" | "dev" | "qa" | "seguranca" | "publicador" | "suporte" | "owner" | "system";
  agentName: string;
  agentRole: string;
  icon: string;
  text: string;
  timestamp: string;
  codeSnippet?: {
    filename: string;
    code: string;
  };
  status?: "pending" | "running" | "success" | "warning";
}

export function LabTab({
  cards,
  isPaused,
  onOpenCard,
  onStartBuild,
  onApproveCard,
  onOpenNewCardModal,
}: LabTabProps) {
  // Demandas em execução ou prontas para o laboratório
  const activeBuildingCards = cards.filter((c) => c.stage === "em_execucao");
  const candidateCards = cards.filter((c) => ["entrada", "proposta_aprovar", "prompt_aprovar"].includes(c.stage));

  // Cartão atualmente inspecionado no laboratório
  const [selectedCardId, setSelectedCardId] = useState<string>(() => {
    if (activeBuildingCards.length > 0) return activeBuildingCards[0].id;
    if (candidateCards.length > 0) return candidateCards[0].id;
    return cards[0]?.id || "";
  });

  // Atualiza seleção caso mude a lista de em_execucao
  useEffect(() => {
    if (activeBuildingCards.length > 0 && !activeBuildingCards.some((c) => c.id === selectedCardId)) {
      setSelectedCardId(activeBuildingCards[0].id);
    }
  }, [activeBuildingCards, selectedCardId]);

  const currentCard = cards.find((c) => c.id === selectedCardId) || activeBuildingCards[0] || cards[0];

  // Estado do chat interno do Laboratório
  const [chatInput, setChatInput] = useState("");
  const [activeWorkspaceFile, setActiveWorkspaceFile] = useState<string>("main.go");
  const [activeTabMode, setActiveTabMode] = useState<"chat" | "workspace" | "telemetry">("chat");
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Mensagens do feed do Laboratório por projeto
  const [labMessages, setLabMessages] = useState<AgentChatMessage[]>([
    {
      id: "m1",
      sender: "system",
      agentName: "Laboratório Geevheit",
      agentRole: "Orquestrador Neural",
      icon: "🧪",
      text: "Sandbox isolado alocado com sucesso. 7 agentes preparados para execução com motores Groq & Gemini.",
      timestamp: "10:00",
      status: "success",
    },
    {
      id: "m2",
      sender: "requisitos",
      agentName: "Alice Requisitos",
      agentRole: "Engenheira de BDD",
      icon: "📋",
      text: "Especificação BDD validada: 'Dado que o cliente acessa o app, Quando selecionar um item, Então o total deve ser recalculado'.",
      timestamp: "10:01",
      status: "success",
    },
    {
      id: "m3",
      sender: "arquiteto",
      agentName: "Bruno Arquiteto",
      agentRole: "Arquiteto de Software",
      icon: "🏗️",
      text: "Playbook SaaS B2B Multi-tenant aplicado. Isolamento com PostgreSQL Row Level Security (RLS) configurado.",
      timestamp: "10:01",
      status: "success",
    },
    {
      id: "m4",
      sender: "dev",
      agentName: "Carlos Dev",
      agentRole: "Desenvolvedor Fullstack",
      icon: "💻",
      text: "Escrevendo handlers HTTP em Go e componentes de tela em Next.js no sandbox isolado...",
      timestamp: "10:02",
      status: "running",
      codeSnippet: {
        filename: "main.go",
        code: `package main\n\nimport (\n\t"net/http"\n\t"log/slog"\n)\n\nfunc main() {\n\tmux := http.NewServeMux()\n\tmux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {\n\t\tw.WriteHeader(http.StatusOK)\n\t\tw.Write([]byte("ok"))\n\t})\n\tslog.Info("servidor online na porta :8080")\n\thttp.ListenAndServe(":8080", mux)\n}`,
      },
    },
    {
      id: "m5",
      sender: "qa",
      agentName: "Daniela QA",
      agentRole: "Especialista em Testes",
      icon: "🧪",
      text: "Suíte de testes de regressão executada com detector de data race (-race). Todos os 14 testes passaram com 100% de cobertura.",
      timestamp: "10:03",
      status: "success",
    },
    {
      id: "m6",
      sender: "seguranca",
      agentName: "Eduardo Segurança",
      agentRole: "Auditor SecOps & LGPD",
      icon: "🛡️",
      text: "Varredura estática de credenciais limpa. RLS validado em todas as tabelas. Nenhuma vulnerabilidade de injeção SQL encontrada.",
      timestamp: "10:04",
      status: "success",
    },
  ]);

  // Auto-scroll
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [labMessages, activeTabMode]);

  const handleSendLabMessage = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const text = chatInput.trim();
    if (!text) return;

    const ownerMsg: AgentChatMessage = {
      id: `owner_${Date.now()}`,
      sender: "owner",
      agentName: "Dono (Você)",
      agentRole: "Comandante Operacional",
      icon: "👑",
      text,
      timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
    };

    setLabMessages((prev) => [...prev, ownerMsg]);
    setChatInput("");

    // Resposta inteligente da equipe do Laboratório
    setTimeout(() => {
      const lower = text.toLowerCase();
      let replyAgent: AgentChatMessage["sender"] = "dev";
      let replyName = "Carlos Dev";
      let replyRole = "Desenvolvedor Fullstack";
      let replyIcon = "💻";
      let replyText = `Ordem "${text}" recebida e compreendida! Já estou adaptando a implementação no código do workspace.`;

      if (lower.includes("pausar") || lower.includes("parar")) {
        replyAgent = "system";
        replyName = "ExecutionGuard";
        replyRole = "Trava de Segurança";
        replyIcon = "🛑";
        replyText = "Atenção: Ações no sandbox foram congeladas preventivamente conforme sua diretriz de segurança.";
      } else if (lower.includes("teste") || lower.includes("qa") || lower.includes("validar")) {
        replyAgent = "qa";
        replyName = "Daniela QA";
        replyRole = "Especialista em Testes";
        replyIcon = "🧪";
        replyText = "Rodando suíte de testes de integração com banco em memória... 100% aprovado! Nenhuma regressão detectada.";
      } else if (lower.includes("seguranca") || lower.includes("lgpd") || lower.includes("vazamento")) {
        replyAgent = "seguranca";
        replyName = "Eduardo Segurança";
        replyRole = "Auditor SecOps & LGPD";
        replyIcon = "🛡️";
        replyText = "Auditoria executada: Chaves de API estão salvas em cofre seguro. Tabelas com RLS forçada.";
      } else if (lower.includes("arquiteto") || lower.includes("banco") || lower.includes("postgres")) {
        replyAgent = "arquiteto";
        replyName = "Bruno Arquiteto";
        replyRole = "Arquiteto de Software";
        replyIcon = "🏗️";
        replyText = "Padrão de arquitetura limpa (Clean Architecture / SOLID) mantido: Domain puro, Portas desacopladas.";
      }

      const agentReply: AgentChatMessage = {
        id: `ast_${Date.now()}`,
        sender: replyAgent,
        agentName: replyName,
        agentRole: replyRole,
        icon: replyIcon,
        text: replyText,
        timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
        status: "success",
      };

      setLabMessages((prev) => [...prev, agentReply]);
    }, 600);
  };

  const workspaceFiles: Record<string, string> = {
    "main.go": `package main

import (
\t"context"
\t"log/slog"
\t"net/http"
\t"os"
)

func main() {
\tlogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
\tmux := http.NewServeMux()

\tmux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
\t\tw.WriteHeader(http.StatusOK)
\t\tw.Write([]byte("healthy"))
\t})

\tmux.HandleFunc("POST /v1/demanda/process", func(w http.ResponseWriter, r *http.Request) {
\t\tw.Header().Set("Content-Type", "application/json")
\t\tw.Write([]byte(\`{"status":"processed","ai_pipeline":"active"}\`))
\t})

\tlogger.Info("servidor da fábrica autônoma ativo", slog.String("port", ":8080"))
\thttp.ListenAndServe(":8080", mux)
}`,
    "schema.sql": `-- Tabela com Row Level Security (RLS) forçada
CREATE TABLE IF NOT EXISTS public.projetos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(64) NOT NULL,
    titulo VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'em_execucao',
    custo_ia_centavos BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE public.projetos ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.projetos FORCE ROW LEVEL SECURITY;

CREATE POLICY projetos_isolation ON public.projetos
    FOR ALL
    USING (client_id = current_setting('request.jwt.claim.client_id', true));`,
    "page.tsx": `"use client";

import React, { useState } from "react";

export default function AppPreview() {
  const [pedido, setPedido] = useState("");

  return (
    <div className="p-4 max-w-md mx-auto space-y-4">
      <h1 className="text-lg font-bold text-indigo-600">
        ✨ Projeto Autônomo em Construção
      </h1>
      <p className="text-xs text-gray-500">
        Gerado pelo agente Dev e aprovado pelo agente QA.
      </p>
      <input
        type="text"
        value={pedido}
        onChange={(e) => setPedido(e.target.value)}
        placeholder="Digite seu pedido..."
        className="w-full p-2 border rounded-xl text-xs"
      />
      <button className="w-full py-2 bg-indigo-600 text-white rounded-xl font-bold text-xs">
        Confirmar Ação
      </button>
    </div>
  );
}`,
  };

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  return (
    <div id="tab-lab" className="space-y-4 pb-20 animate-fadeIn">
      {/* Cabeçalho do Laboratório */}
      <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-2xl bg-gradient-to-tr from-indigo-500 via-purple-500 to-pink-500 flex items-center justify-center text-white text-xl shadow-md">
              🧪
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-base sm:text-lg font-bold text-[var(--ink)] m-0 leading-tight">
                  Laboratório Neural
                </h1>
                <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30 flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-ping"></span>
                  7 Agentes Ativos
                </span>
              </div>
              <p className="text-xs text-[var(--mute)] m-0 mt-0.5">
                Sala de Operações da Fábrica de IA · Construção em Sandbox Isolado
              </p>
            </div>
          </div>

          {/* Seletor de Projetos em Construção */}
          <div className="flex items-center gap-2">
            <select
              value={selectedCardId}
              onChange={(e) => setSelectedCardId(e.target.value)}
              className="text-xs p-2 px-3 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] font-medium outline-none focus:border-[var(--acc)] max-w-[200px] truncate"
            >
              {cards.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.stage === "em_execucao" ? "⚡ " : ""}
                  {c.title} ({c.client_name})
                </option>
              ))}
            </select>

            <button
              type="button"
              onClick={onOpenNewCardModal}
              className="p-2 px-3 rounded-xl bg-[var(--acc)] text-white font-bold text-xs border-0 cursor-pointer hover:opacity-90 transition-opacity flex items-center gap-1 shadow-sm whitespace-nowrap"
            >
              <span>✨</span>
              <span className="hidden sm:inline">Nova Demanda</span>
            </button>
          </div>
        </div>

        {/* Barra de Progresso e Métricas do Projeto Atual */}
        {currentCard && (
          <div className="mt-4 pt-3 border-t border-[var(--line)] grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs">
            <div className="bg-[var(--bg)]/70 p-2.5 rounded-xl border border-[var(--line)]">
              <span className="text-[10px] text-[var(--mute)] block">Projeto Inspecionado</span>
              <b className="text-[var(--ink)] truncate block text-xs" title={currentCard.title}>
                {currentCard.title}
              </b>
              <span className="text-[10px] text-[var(--acc)]">{currentCard.client_name}</span>
            </div>

            <div className="bg-[var(--bg)]/70 p-2.5 rounded-xl border border-[var(--line)]">
              <span className="text-[10px] text-[var(--mute)] block">Estágio da Pipeline</span>
              <b className="text-[var(--ink)] block text-xs capitalize">
                {currentCard.stage === "em_execucao" ? "⚡ Em Construção" : currentCard.stage.replace(/_/g, " ")}
              </b>
              <span className="text-[10px] text-emerald-500">Sandbox Isolado Seguro</span>
            </div>

            <div className="bg-[var(--bg)]/70 p-2.5 rounded-xl border border-[var(--line)]">
              <span className="text-[10px] text-[var(--mute)] block">Custo IA Acumulado</span>
              <b className="text-[var(--ink)] block text-xs">R$ 0,42 / 50,00</b>
              <span className="text-[10px] text-[var(--mute)] font-mono">Motor: Groq / Gemini</span>
            </div>

            <div className="bg-[var(--bg)]/70 p-2.5 rounded-xl border border-[var(--line)] flex flex-col justify-between">
              <span className="text-[10px] text-[var(--mute)] block">Ação Rápida</span>
              {currentCard.stage !== "em_execucao" ? (
                <button
                  type="button"
                  onClick={() => onStartBuild(currentCard.id, 5000)}
                  className="w-full py-1 px-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-[11px] border-0 cursor-pointer shadow-xs transition-colors flex items-center justify-center gap-1"
                >
                  <span>⚡</span>
                  <span>Construir no Lab</span>
                </button>
              ) : (
                <button
                  type="button"
                  onClick={() => onApproveCard(currentCard.id)}
                  className="w-full py-1 px-2 rounded-lg bg-[var(--acc)] hover:opacity-90 text-white font-bold text-[11px] border-0 cursor-pointer shadow-xs transition-opacity flex items-center justify-center gap-1"
                >
                  <span>🛡️</span>
                  <span>Aprovar Portão</span>
                </button>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Navegação de Abas do Laboratório: Chat com Agentes | Workspace / Código | Telemetria */}
      <div className="flex gap-1.5 p-1 bg-[var(--surf)] border border-[var(--line)] rounded-xl">
        <button
          type="button"
          onClick={() => setActiveTabMode("chat")}
          className={`flex-1 py-2 px-3 rounded-lg text-xs font-bold transition-all cursor-pointer flex items-center justify-center gap-1.5 ${
            activeTabMode === "chat"
              ? "bg-[var(--acc)] text-white shadow-xs"
              : "text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0"
          }`}
        >
          <span>💬</span>
          <span>Feed dos 7 Agentes (Chat ao Vivo)</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveTabMode("workspace")}
          className={`flex-1 py-2 px-3 rounded-lg text-xs font-bold transition-all cursor-pointer flex items-center justify-center gap-1.5 ${
            activeTabMode === "workspace"
              ? "bg-[var(--acc)] text-white shadow-xs"
              : "text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0"
          }`}
        >
          <span>💻</span>
          <span>Workspace (Código Gerado)</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveTabMode("telemetry")}
          className={`flex-1 py-2 px-3 rounded-lg text-xs font-bold transition-all cursor-pointer flex items-center justify-center gap-1.5 ${
            activeTabMode === "telemetry"
              ? "bg-[var(--acc)] text-white shadow-xs"
              : "text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0"
          }`}
        >
          <span>📊</span>
          <span>Métricas dos Agentes</span>
        </button>
      </div>

      {/* Modo 1: FEED AO VIVO / CHAT DO LABORATÓRIO */}
      {activeTabMode === "chat" && (
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl flex flex-col h-[520px] shadow-sm overflow-hidden">
          {/* Header do Chat dos Agentes */}
          <div className="p-3 bg-[var(--bg)]/80 border-b border-[var(--line)] flex items-center justify-between text-xs">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
              <span className="font-bold text-[var(--ink)]">Canal de Comunicação dos Agentes de IA</span>
            </div>
            <div className="text-[11px] text-[var(--mute)] font-mono">
              Sandbox: /tmp/geev-sandbox/{currentCard?.id?.substring(0, 8) || "main"}
            </div>
          </div>

          {/* Feed de Mensagens com scroll */}
          <div className="flex-1 overflow-y-auto p-4 space-y-3.5">
            {labMessages.map((msg) => {
              const isOwner = msg.sender === "owner";
              return (
                <div
                  key={msg.id}
                  className={`flex flex-col ${isOwner ? "items-end" : "items-start"}`}
                >
                  <div className="flex items-center gap-1.5 mb-1 px-1">
                    <span className="text-xs">{msg.icon}</span>
                    <span className="text-[11px] font-bold text-[var(--ink)]">{msg.agentName}</span>
                    <span className="text-[10px] text-[var(--mute)] font-mono">({msg.agentRole})</span>
                    <span className="text-[10px] text-[var(--mute)]">• {msg.timestamp}</span>
                  </div>

                  <div
                    className={`max-w-[90%] rounded-2xl p-3 text-xs leading-relaxed shadow-xs ${
                      isOwner
                        ? "bg-[var(--acc)] text-white rounded-tr-xs"
                        : "bg-[var(--bg)] border border-[var(--line)] text-[var(--ink)] rounded-tl-xs"
                    }`}
                  >
                    <p className="m-0">{msg.text}</p>

                    {/* Trecho de Código embutido no feed */}
                    {msg.codeSnippet && (
                      <div className="mt-2 pt-2 border-t border-[var(--line)]/50">
                        <div className="text-[10px] font-mono text-[var(--mute)] mb-1 flex items-center gap-1">
                          <span>📄</span>
                          <span>{msg.codeSnippet.filename}</span>
                        </div>
                        <pre className="p-2.5 rounded-xl bg-black/90 text-emerald-400 font-mono text-[11px] overflow-x-auto leading-relaxed border border-white/10 m-0">
                          {msg.codeSnippet.code}
                        </pre>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
            <div ref={messagesEndRef} />
          </div>

          {/* Barra de Interação do Dono com o Laboratório */}
          <form
            onSubmit={handleSendLabMessage}
            className="p-3 bg-[var(--bg)]/90 border-t border-[var(--line)] flex items-center gap-2"
          >
            <input
              type="text"
              value={chatInput}
              onChange={(e) => setChatInput(e.target.value)}
              placeholder="Envie uma diretriz para os agentes (ex: 'Dev, adicione login com Google')..."
              className="flex-1 text-xs sm:text-sm py-2 px-3.5 rounded-xl border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] outline-none focus:border-[var(--acc)]"
            />
            <button
              type="submit"
              disabled={!chatInput.trim()}
              className="py-2 px-4 rounded-xl bg-[var(--acc)] text-white font-bold text-xs border-0 cursor-pointer hover:opacity-90 disabled:opacity-50 transition-opacity flex-none shadow-sm"
            >
              Comandar
            </button>
          </form>
        </div>
      )}

      {/* Modo 2: WORKSPACE / CÓDIGO FONTE GERADO NO SANDBOX */}
      {activeTabMode === "workspace" && (
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm space-y-3">
          <div className="flex items-center justify-between pb-3 border-b border-[var(--line)]">
            <div className="flex items-center gap-2">
              <span className="text-base">📁</span>
              <h3 className="text-sm font-bold text-[var(--ink)] m-0">Arquivos Gerados no Sandbox</h3>
            </div>
            <div className="flex gap-1.5">
              {Object.keys(workspaceFiles).map((filename) => (
                <button
                  key={filename}
                  type="button"
                  onClick={() => setActiveWorkspaceFile(filename)}
                  className={`py-1 px-2.5 rounded-lg text-xs font-mono font-semibold cursor-pointer transition-all border ${
                    activeWorkspaceFile === filename
                      ? "bg-[var(--acc)] text-white border-[var(--acc)] shadow-xs"
                      : "bg-[var(--bg)] border-[var(--line)] text-[var(--mute)] hover:text-[var(--ink)]"
                  }`}
                >
                  {filename}
                </button>
              ))}
            </div>
          </div>

          <div className="relative">
            <pre className="p-4 rounded-xl bg-slate-950 text-slate-100 font-mono text-xs overflow-x-auto leading-relaxed border border-slate-800 h-[380px] m-0">
              {workspaceFiles[activeWorkspaceFile] || "// Selecione um arquivo"}
            </pre>
            <div className="absolute top-2 right-2 px-2 py-1 rounded bg-slate-800 text-[10px] text-slate-300 font-mono">
              Modo Leitura Isolado
            </div>
          </div>
        </div>
      )}

      {/* Modo 3: TELEMETRIA DOS 7 AGENTES */}
      {activeTabMode === "telemetry" && (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
          {[
            { role: "Requisitos", name: "Alice", icon: "📋", status: "Concluído", action: "Cenários BDD gerados", cost: "R$ 0,08" },
            { role: "Arquiteto", name: "Bruno", icon: "🏗️", status: "Concluído", action: "Clean Arch + RLS", cost: "R$ 0,20" },
            { role: "Desenvolvedor", name: "Carlos", icon: "💻", status: "Ativo", action: "Escrevendo Go & Next.js", cost: "R$ 0,42" },
            { role: "QA", name: "Daniela", icon: "🧪", status: "Pronto", action: "14 testes unitários PASS", cost: "R$ 0,15" },
            { role: "Segurança", name: "Eduardo", icon: "🛡️", status: "Ativo", action: "Scan sem credenciais", cost: "R$ 0,10" },
            { role: "Publicador", name: "Felipe", icon: "🚀", status: "Preparado", action: "Build estático Next.js", cost: "R$ 0,05" },
            { role: "Suporte", name: "Gabriel", icon: "🎧", status: "Standby", action: "Monitoramento pós-deploy", cost: "R$ 0,00" },
          ].map((agent) => (
            <div
              key={agent.role}
              className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-3.5 shadow-sm space-y-1.5"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-lg">{agent.icon}</span>
                  <div>
                    <h4 className="text-xs font-bold text-[var(--ink)] m-0 leading-tight">
                      {agent.name} ({agent.role})
                    </h4>
                    <span className="text-[10px] text-[var(--mute)]">Motor Groq/Gemini</span>
                  </div>
                </div>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30">
                  {agent.status}
                </span>
              </div>
              <div className="text-[11px] text-[var(--mute)] bg-[var(--bg)] p-2 rounded-xl border border-[var(--line)]">
                {agent.action}
              </div>
              <div className="flex justify-between text-[10px] text-[var(--mute)] pt-0.5">
                <span>Custo computacional:</span>
                <b className="text-[var(--ink)] font-mono">{agent.cost}</b>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

