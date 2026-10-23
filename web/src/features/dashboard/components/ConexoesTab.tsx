"use client";

import React, { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ConnectorItem,
  api,
  WebhookEndpoint,
  NotificationPreferences,
} from "@/shared/api-client/endpoints";

interface ConexoesTabProps {
  connectors: ConnectorItem[];
  onShowToast: (msg: string) => void;
}

export function ConexoesTab({ connectors, onShowToast }: ConexoesTabProps) {
  const queryClient = useQueryClient();
  const [activeSection, setActiveSection] = useState<"notificacoes" | "webhooks" | "conectores">("webhooks");

  // Estados do modal de novo webhook
  const [showWebhookModal, setShowWebhookModal] = useState(false);
  const [webhookURL, setWebhookURL] = useState("");
  const [webhookDesc, setWebhookDesc] = useState("");
  const [selectedEvents, setSelectedEvents] = useState<string[]>(["*"]);
  const [createdSecret, setCreatedSecret] = useState<string | null>(null);

  // Queries
  const { data: webhooks = [], isLoading: loadingWebhooks } = useQuery({
    queryKey: ["webhooks"],
    queryFn: () => api.listWebhooks(),
  });

  const { data: deliveries = [] } = useQuery({
    queryKey: ["webhook-deliveries"],
    queryFn: () => api.listWebhookDeliveries(),
  });

  const { data: preferences } = useQuery({
    queryKey: ["notification-preferences"],
    queryFn: () => api.getNotificationPreferences(),
  });

  // Mutations
  const createWebhookMutation = useMutation({
    mutationFn: (data: { url: string; description: string; events: string[] }) =>
      api.createWebhook(data),
    onSuccess: (ep) => {
      queryClient.invalidateQueries({ queryKey: ["webhooks"] });
      setCreatedSecret(ep.secret_token || null);
      setWebhookURL("");
      setWebhookDesc("");
      onShowToast("Webhook criado com sucesso! Guarde o segredo HMAC.");
    },
    onError: (err: unknown) => {
      const msg = err instanceof Error ? err.message : "Erro ao criar webhook";
      alert(msg);
    },
  });

  const deleteWebhookMutation = useMutation({
    mutationFn: (id: string) => api.deleteWebhook(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks"] });
      onShowToast("Webhook removido");
    },
  });

  const redeliverMutation = useMutation({
    mutationFn: ({ id, event, payload }: { id: string; event: string; payload: Record<string, unknown> }) =>
      api.redeliverWebhook(id, event, payload),
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ["webhook-deliveries"] });
      alert(`Reenvio concluído!\nStatus: ${res.status_code}\nSucesso: ${res.success}`);
    },
  });

  const savePreferencesMutation = useMutation({
    mutationFn: (prefs: NotificationPreferences) => api.saveNotificationPreferences(prefs),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notification-preferences"] });
      onShowToast("Preferências de notificação salvas com sucesso!");
    },
  });

  const [customPrefs, setCustomPrefs] = useState<NotificationPreferences | null>(null);
  const activePrefs: NotificationPreferences = customPrefs ?? preferences ?? {
    channels: {
      telegram_enabled: false,
      whatsapp_enabled: false,
      email_enabled: false,
    },
    events: {
      intake_enabled: true,
      pix_enabled: true,
      portal_accept_enabled: true,
      subscription_overdue_enabled: true,
    },
  };

  const fallbackList = [
    "GitHub",
    "Vercel",
    "Render",
    "Neon",
    "Supabase",
    "Netlify",
    "Firebase",
    "Docker Hub",
    "Provedor de IA (agentes)",
  ];

  const items = connectors.length > 0
    ? connectors
    : fallbackList.map((name, i) => ({
        id: `con-${i}`,
        name,
        active: false,
        secret_ref: "",
        scope: "read_only",
      }));

  const handleToggle = () => {
    onShowToast("Continua desligado. Ligue só depois de configurar a chave nas variáveis do servidor.");
  };

  const handleEventToggle = (evt: string) => {
    if (evt === "*") {
      setSelectedEvents(["*"]);
      return;
    }
    const filtered = selectedEvents.filter((e) => e !== "*");
    if (filtered.includes(evt)) {
      const next = filtered.filter((e) => e !== evt);
      setSelectedEvents(next.length === 0 ? ["*"] : next);
    } else {
      setSelectedEvents([...filtered, evt]);
    }
  };

  return (
    <div id="tab-conexoes" className="space-y-4">
      {/* Navegação entre Sub-Seções */}
      <div className="flex bg-[var(--surf)] p-1 rounded-2xl border border-[var(--line)]">
        <button
          type="button"
          onClick={() => setActiveSection("webhooks")}
          className={`flex-1 py-2 text-xs font-bold rounded-xl transition-all cursor-pointer ${
            activeSection === "webhooks"
              ? "bg-[var(--acc)] text-white shadow"
              : "text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          Webhooks de Saída (n8n / CRM)
        </button>
        <button
          type="button"
          onClick={() => setActiveSection("notificacoes")}
          className={`flex-1 py-2 text-xs font-bold rounded-xl transition-all cursor-pointer ${
            activeSection === "notificacoes"
              ? "bg-[var(--acc)] text-white shadow"
              : "text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          Notificações do Dono
        </button>
        <button
          type="button"
          onClick={() => setActiveSection("conectores")}
          className={`flex-1 py-2 text-xs font-bold rounded-xl transition-all cursor-pointer ${
            activeSection === "conectores"
              ? "bg-[var(--acc)] text-white shadow"
              : "text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          Conectores Nuvem
        </button>
      </div>

      {/* 1. SEÇÃO DE WEBHOOKS (Fase 15) */}
      {activeSection === "webhooks" && (
        <div className="space-y-4">
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 flex items-center justify-between gap-3 shadow-sm">
            <div>
              <h3 className="text-sm font-bold text-[var(--ink)] m-0">Webhooks de Saída</h3>
              <p className="text-xs text-[var(--mute)] m-0 mt-0.5">
                Despache eventos em tempo real para n8n, Make, CRM ou sistemas externos com assinatura HMAC-SHA256 e proteção anti-SSRF.
              </p>
            </div>
            <button
              onClick={() => {
                setCreatedSecret(null);
                setShowWebhookModal(true);
              }}
              className="px-3 py-1.5 rounded-xl bg-[var(--acc)] text-white text-xs font-bold hover:opacity-90 cursor-pointer flex-none"
            >
              + Novo Webhook
            </button>
          </div>

          {/* Lista de Endpoints */}
          <div className="space-y-3">
            {loadingWebhooks ? (
              <p className="text-xs text-[var(--mute)]">Carregando webhooks...</p>
            ) : webhooks.length === 0 ? (
              <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-6 text-center text-xs text-[var(--mute)]">
                Nenhum webhook cadastrado. Clique em &quot;+ Novo Webhook&quot; para conectar seu n8n ou CRM.
              </div>
            ) : (
              webhooks.map((ep: WebhookEndpoint) => (
                <div
                  key={ep.id}
                  className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 space-y-2 shadow-sm"
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-[var(--ok)]" />
                      <b className="text-xs font-bold text-[var(--ink)] font-mono">{ep.url}</b>
                    </div>
                    <button
                      onClick={() => {
                        if (confirm(`Remover webhook ${ep.url}?`)) {
                          deleteWebhookMutation.mutate(ep.id);
                        }
                      }}
                      className="text-xs text-red-500 hover:underline cursor-pointer"
                    >
                      Excluir
                    </button>
                  </div>
                  {ep.description && (
                    <div className="text-xs text-[var(--mute)]">{ep.description}</div>
                  )}
                  <div className="flex flex-wrap gap-1 pt-1">
                    {ep.events.map((ev) => (
                      <span
                        key={ev}
                        className="px-2 py-0.5 rounded-md bg-[var(--bg)] border border-[var(--line)] text-[10px] text-[var(--ink)] font-mono"
                      >
                        {ev}
                      </span>
                    ))}
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Histórico Recente de Entregas */}
          {deliveries.length > 0 && (
            <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 space-y-3 shadow-sm">
              <h3 className="text-xs font-bold text-[var(--ink)] uppercase tracking-wider m-0">
                Histórico Recente de Entregas
              </h3>
              <div className="space-y-2 max-h-60 overflow-y-auto">
                {deliveries.map((del) => (
                  <div
                    key={del.id}
                    className="flex items-center justify-between p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-xs"
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <span
                          className={`px-1.5 py-0.5 rounded text-[10px] font-bold ${
                            del.success ? "bg-emerald-500/20 text-emerald-500" : "bg-red-500/20 text-red-500"
                          }`}
                        >
                          {del.status_code > 0 ? `${del.status_code} HTTP` : "Falha"}
                        </span>
                        <b className="text-[var(--ink)] font-mono">{del.event}</b>
                      </div>
                      <div className="text-[10px] text-[var(--mute)] mt-0.5">
                        Tentativas: {del.attempts} · {new Date(del.created_at).toLocaleTimeString("pt-BR")}
                      </div>
                    </div>
                    <button
                      onClick={() =>
                        redeliverMutation.mutate({
                          id: del.webhook_id,
                          event: del.event,
                          payload: del.payload,
                        })
                      }
                      disabled={redeliverMutation.isPending}
                      className="px-2 py-1 text-[11px] rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] hover:border-[var(--acc)] cursor-pointer"
                    >
                      Reenviar
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* 2. SEÇÃO DE NOTIFICAÇÕES (Fase 15) */}
      {activeSection === "notificacoes" && (
        <div className="space-y-4">
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm space-y-4">
            <div>
              <h3 className="text-sm font-bold text-[var(--ink)] m-0">Canais de Notificação do Dono</h3>
              <p className="text-xs text-[var(--mute)] m-0 mt-0.5">
                Escolha por onde receber alertas críticos instantâneos pelo celular.
              </p>
            </div>

            {/* Toggles de Canais */}
            <div className="space-y-3 border-t border-[var(--line)] pt-3">
              <div className="flex items-center justify-between text-xs">
                <div>
                  <b className="text-[var(--ink)] block">Telegram Bot</b>
                  <span className="text-[var(--mute)]">Receba mensagens diretas via bot configurado no servidor.</span>
                </div>
                <input
                  type="checkbox"
                  checked={activePrefs.channels.telegram_enabled}
                  onChange={(e) =>
                    setCustomPrefs({
                      ...activePrefs,
                      channels: { ...activePrefs.channels, telegram_enabled: e.target.checked },
                    })
                  }
                  className="w-5 h-5 rounded cursor-pointer"
                />
              </div>

              <div className="flex items-center justify-between text-xs">
                <div>
                  <b className="text-[var(--ink)] block">WhatsApp</b>
                  <span className="text-[var(--mute)]">Alertas via API/Gateway de WhatsApp para seu número pessoal.</span>
                </div>
                <input
                  type="checkbox"
                  checked={activePrefs.channels.whatsapp_enabled}
                  onChange={(e) =>
                    setCustomPrefs({
                      ...activePrefs,
                      channels: { ...activePrefs.channels, whatsapp_enabled: e.target.checked },
                    })
                  }
                  className="w-5 h-5 rounded cursor-pointer"
                />
              </div>

              <div className="flex items-center justify-between text-xs">
                <div>
                  <b className="text-[var(--ink)] block">E-mail Operacional</b>
                  <span className="text-[var(--mute)]">Relatórios e avisos em formato texto para seu e-mail de dono.</span>
                </div>
                <input
                  type="checkbox"
                  checked={activePrefs.channels.email_enabled}
                  onChange={(e) =>
                    setCustomPrefs({
                      ...activePrefs,
                      channels: { ...activePrefs.channels, email_enabled: e.target.checked },
                    })
                  }
                  className="w-5 h-5 rounded cursor-pointer"
                />
              </div>
            </div>

            {/* Eventos que disparam alertas */}
            <div className="space-y-3 border-t border-[var(--line)] pt-3">
              <h4 className="text-xs font-bold text-[var(--ink)] uppercase tracking-wider m-0">
                Eventos Críticos Monitorados
              </h4>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
                <label className="flex items-center gap-2 p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] cursor-pointer">
                  <input
                    type="checkbox"
                    checked={activePrefs.events.intake_enabled}
                    onChange={(e) =>
                      setCustomPrefs({
                        ...activePrefs,
                        events: { ...activePrefs.events, intake_enabled: e.target.checked },
                      })
                    }
                  />
                  <span>Novo pedido em <code>/intake</code></span>
                </label>

                <label className="flex items-center gap-2 p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] cursor-pointer">
                  <input
                    type="checkbox"
                    checked={activePrefs.events.pix_enabled}
                    onChange={(e) =>
                      setCustomPrefs({
                        ...activePrefs,
                        events: { ...activePrefs.events, pix_enabled: e.target.checked },
                      })
                    }
                  />
                  <span>Pagamento PIX confirmado</span>
                </label>

                <label className="flex items-center gap-2 p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] cursor-pointer">
                  <input
                    type="checkbox"
                    checked={activePrefs.events.portal_accept_enabled}
                    onChange={(e) =>
                      setCustomPrefs({
                        ...activePrefs,
                        events: { ...activePrefs.events, portal_accept_enabled: e.target.checked },
                      })
                    }
                  />
                  <span>Aceite formal no Portal</span>
                </label>

                <label className="flex items-center gap-2 p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] cursor-pointer">
                  <input
                    type="checkbox"
                    checked={activePrefs.events.subscription_overdue_enabled}
                    onChange={(e) =>
                      setCustomPrefs({
                        ...activePrefs,
                        events: {
                          ...activePrefs.events,
                          subscription_overdue_enabled: e.target.checked,
                        },
                      })
                    }
                  />
                  <span>Mensalidade em atraso</span>
                </label>
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <button
                onClick={() => savePreferencesMutation.mutate(activePrefs)}
                disabled={savePreferencesMutation.isPending}
                className="px-4 py-2 rounded-xl bg-[var(--acc)] text-white text-xs font-bold hover:opacity-90 cursor-pointer"
              >
                {savePreferencesMutation.isPending ? "Salvando..." : "Salvar Preferências"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 3. SEÇÃO DE CONECTORES NUVEM */}
      {activeSection === "conectores" && (
        <div className="space-y-3">
          <div className="bg-[var(--accs)] text-[var(--ink)] rounded-xl p-3 text-xs sm:text-sm font-medium leading-relaxed">
            Os conectores e agentes ficam desligados. Quando decidir sobre os agentes, basta configurar a chave e ligar. Comece sempre com chaves somente leitura.
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {items.map((conn) => (
              <div
                key={conn.id}
                className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 flex items-center justify-between gap-3 shadow-sm"
              >
                <div>
                  <h3 className="text-sm font-bold text-[var(--ink)] m-0 leading-tight">
                    {conn.name}
                  </h3>
                  <div className="text-xs text-[var(--mute)] mt-0.5">
                    {conn.active ? "Ativo" : "Desligado · sem chave"}
                  </div>
                </div>

                <button
                  id={`switch-conn-${conn.id}`}
                  onClick={() => handleToggle()}
                  className="w-11 h-6 rounded-full bg-[var(--line)] border-0 relative cursor-pointer flex-none transition-colors"
                  aria-label={`Ligar conector ${conn.name}`}
                >
                  <span className="block w-5 h-5 rounded-full bg-[var(--surf)] shadow absolute top-0.5 left-0.5 transition-transform" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* MODAL: CRIAR NOVO WEBHOOK */}
      {showWebhookModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4">
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 max-w-md w-full shadow-2xl space-y-4">
            <h3 className="text-base font-bold text-[var(--ink)] m-0">Novo Webhook de Saída</h3>

            {createdSecret ? (
              <div className="space-y-3">
                <div className="bg-emerald-500/10 border border-emerald-500/30 rounded-xl p-3 text-xs text-emerald-500">
                  ✅ <b>Webhook cadastrado com sucesso!</b>
                </div>

                <div>
                  <label className="text-xs text-[var(--mute)] block mb-1">
                    Segredo de Assinatura HMAC-SHA256 (Exibido apenas uma vez):
                  </label>
                  <div className="p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] font-mono text-[11px] break-all select-all text-[var(--ink)]">
                    {createdSecret}
                  </div>
                </div>

                <p className="text-[11px] text-[var(--mute)] leading-relaxed">
                  Utilize este segredo no seu n8n / servidor para validar o cabeçalho <code>X-Geevheit-Signature</code>.
                </p>

                <div className="flex justify-end pt-2">
                  <button
                    onClick={() => setShowWebhookModal(false)}
                    className="px-4 py-2 rounded-xl bg-[var(--acc)] text-white text-xs font-bold cursor-pointer"
                  >
                    Concluído
                  </button>
                </div>
              </div>
            ) : (
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  createWebhookMutation.mutate({
                    url: webhookURL,
                    description: webhookDesc,
                    events: selectedEvents,
                  });
                }}
                className="space-y-3"
              >
                <div>
                  <label className="text-xs text-[var(--mute)] block mb-1">URL de Destino (HTTPS)</label>
                  <input
                    type="url"
                    required
                    placeholder="https://seu-n8n.com/webhook/geevheit"
                    value={webhookURL}
                    onChange={(e) => setWebhookURL(e.target.value)}
                    className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
                  />
                </div>

                <div>
                  <label className="text-xs text-[var(--mute)] block mb-1">Descrição / Finalidade</label>
                  <input
                    type="text"
                    placeholder="Ex.: n8n CRM Pipeline, Notificações Externas..."
                    value={webhookDesc}
                    onChange={(e) => setWebhookDesc(e.target.value)}
                    className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
                  />
                </div>

                <div>
                  <label className="text-xs text-[var(--mute)] block mb-1">Eventos Assinados</label>
                  <div className="grid grid-cols-2 gap-2 text-xs">
                    <label className="flex items-center gap-1.5 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={selectedEvents.includes("*")}
                        onChange={() => handleEventToggle("*")}
                      />
                      <span>Todos (*)</span>
                    </label>
                    <label className="flex items-center gap-1.5 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={selectedEvents.includes("intake.received")}
                        onChange={() => handleEventToggle("intake.received")}
                      />
                      <span>intake.received</span>
                    </label>
                    <label className="flex items-center gap-1.5 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={selectedEvents.includes("payment.pix_confirmed")}
                        onChange={() => handleEventToggle("payment.pix_confirmed")}
                      />
                      <span>payment.pix_confirmed</span>
                    </label>
                    <label className="flex items-center gap-1.5 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={selectedEvents.includes("portal.delivery_accepted")}
                        onChange={() => handleEventToggle("portal.delivery_accepted")}
                      />
                      <span>portal.delivery_accepted</span>
                    </label>
                    <label className="flex items-center gap-1.5 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={selectedEvents.includes("subscription.overdue")}
                        onChange={() => handleEventToggle("subscription.overdue")}
                      />
                      <span>subscription.overdue</span>
                    </label>
                  </div>
                </div>

                <div className="flex justify-end gap-2 pt-2">
                  <button
                    type="button"
                    onClick={() => setShowWebhookModal(false)}
                    className="px-3 py-2 text-xs font-semibold rounded-xl border border-[var(--line)] bg-transparent text-[var(--ink)] cursor-pointer"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    disabled={createWebhookMutation.isPending}
                    className="px-3 py-2 text-xs font-semibold rounded-xl bg-[var(--acc)] text-white border-0 cursor-pointer"
                  >
                    {createWebhookMutation.isPending ? "Cadastrando..." : "Cadastrar Webhook"}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
