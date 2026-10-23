"use client";

import React from "react";
import { api, CardDetails, ProposalDocument, PortalStatusResponse } from "@/shared/api-client/endpoints";
import { ProposalModal } from "./ProposalModal";

interface CardDetailSheetProps {
  cardDetails: CardDetails | null;
  isOpen: boolean;
  onClose: () => void;
  onApprove: (cardId: string) => void;
  onAskInfo: (cardId: string) => void;
  onReject: (cardId: string) => void;
  onBuild?: (cardId: string) => void;
  isProcessing: boolean;
}

export function CardDetailSheet({
  cardDetails,
  isOpen,
  onClose,
  onApprove,
  onAskInfo,
  onReject,
  onBuild,
  isProcessing,
}: CardDetailSheetProps) {
  const [riskAccepted, setRiskAccepted] = React.useState(false);
  const [pipelineData, setPipelineData] = React.useState<{
    active_role: string;
    spend_total_cents: number;
    status: string;
    current_task: string;
  } | null>(null);

  const [pixData, setPixData] = React.useState<{
    id: string;
    pix_emv: string;
    pix_txid: string;
    status: string;
  } | null>(null);
  const [loadingPix, setLoadingPix] = React.useState(false);
  const [copiedPix, setCopiedPix] = React.useState(false);

  const [portalStatus, setPortalStatus] = React.useState<PortalStatusResponse | null>(null);
  const [loadingPortal, setLoadingPortal] = React.useState(false);
  const [copiedPortal, setCopiedPortal] = React.useState(false);

  const [proposalData, setProposalData] = React.useState<ProposalDocument | null>(null);
  const [isProposalOpen, setIsProposalOpen] = React.useState(false);
  const [loadingProposal, setLoadingProposal] = React.useState(false);

  const handleOpenProposal = async () => {
    if (!cardDetails?.card?.id) return;
    setLoadingProposal(true);
    try {
      const doc = await api.getProposalDocument(cardDetails.card.id);
      setProposalData(doc);
      setIsProposalOpen(true);
    } catch {
      try {
        const doc = await api.generateProposalDocument(cardDetails.card.id);
        setProposalData(doc);
        setIsProposalOpen(true);
      } catch (err) {
        console.error("Falha ao abrir proposta comercial:", err);
      }
    } finally {
      setLoadingProposal(false);
    }
  };

  React.useEffect(() => {
    if (cardDetails?.card?.id) {
      api.getPipeline(cardDetails.card.id)
        .then((p) => {
          setPipelineData(p);
        })
        .catch(() => {
          setPipelineData(null);
        });

      if (cardDetails.card.price_cents > 0) {
        api.getPix(cardDetails.card.id)
          .then((p) => setPixData(p))
          .catch(() => setPixData(null));
      } else {
        Promise.resolve().then(() => setPixData(null));
      }

      api.getPortalStatus(cardDetails.card.id)
        .then((st) => setPortalStatus(st))
        .catch(() => setPortalStatus(null));
    }
  }, [cardDetails?.card?.id, cardDetails?.card?.price_cents]);

  const handleGeneratePix = async () => {
    if (!cardDetails?.card?.id) return;
    setLoadingPix(true);
    try {
      const p = await api.generatePix(cardDetails.card.id);
      setPixData(p);
    } catch {
      // fallback
    } finally {
      setLoadingPix(false);
    }
  };

  const handleCopyPix = () => {
    if (pixData?.pix_emv) {
      navigator.clipboard.writeText(pixData.pix_emv);
      setCopiedPix(true);
      setTimeout(() => setCopiedPix(false), 3000);
    }
  };

  const handleGeneratePortal = async () => {
    if (!cardDetails?.card?.id) return;
    setLoadingPortal(true);
    try {
      const res = await api.generatePortalLink(cardDetails.card.id, 30);
      setPortalStatus({
        card_id: res.card_id,
        is_active: true,
        portal_url: res.portal_url,
        expires_at: new Date(Date.now() + 30 * 24 * 3600 * 1000).toISOString(),
      });
      if (res.portal_url) {
        navigator.clipboard.writeText(res.portal_url);
        setCopiedPortal(true);
        setTimeout(() => setCopiedPortal(false), 3000);
      }
    } catch (err) {
      console.error("Erro ao gerar link do portal:", err);
    } finally {
      setLoadingPortal(false);
    }
  };

  const handleRevokePortal = async () => {
    if (!cardDetails?.card?.id) return;
    if (!confirm("Tem certeza que deseja revogar o link do portal do cliente? O cliente perderá o acesso imediatamente.")) return;
    setLoadingPortal(true);
    try {
      await api.revokePortalLink(cardDetails.card.id);
      setPortalStatus(null);
    } catch (err) {
      console.error("Erro ao revogar portal:", err);
    } finally {
      setLoadingPortal(false);
    }
  };

  const handleCopyPortal = () => {
    if (portalStatus?.portal_url) {
      navigator.clipboard.writeText(portalStatus.portal_url);
      setCopiedPortal(true);
      setTimeout(() => setCopiedPortal(false), 3000);
    }
  };

  if (!isOpen || !cardDetails) return null;

  const { card, events, risk_report, actions, margin_cents } = cardDetails;
  const isDeliveryGate = card.stage === "entrega_aprovar";

  const typeConfig: Record<string, { label: string; color: string }> = {
    orcamento: { label: "Orçamento", color: "#0e6b73" },
    bug: { label: "Bug", color: "#b42318" },
    suporte: { label: "Suporte", color: "#b45309" },
    financeiro: { label: "Financeiro", color: "#6d4bb5" },
  };

  const currentType = typeConfig[card.type] || { label: card.type, color: "var(--mute)" };

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  return (
    <div
      id="card-detail-backdrop"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      className="fixed inset-0 bg-black/70 z-50 flex items-end justify-center sm:items-center sm:p-4"
    >
      <div
        id="card-detail-sheet"
        className="bg-[var(--surf)] border-t sm:border border-[var(--line)] w-full max-w-[560px] rounded-t-3xl sm:rounded-2xl p-5 max-h-[85vh] overflow-y-auto animate-in slide-in-from-bottom-5 duration-200"
        style={{ paddingBottom: "calc(20px + env(safe-area-inset-bottom, 0px))" }}
      >
        {/* Header do Cartão */}
        <div className="flex items-center justify-between mb-2">
          <span
            className="text-xs px-2.5 py-0.5 rounded-full font-semibold"
            style={{ backgroundColor: "var(--accs)", color: currentType.color }}
          >
            {currentType.label}
          </span>
          <button
            onClick={onClose}
            className="text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0 text-xl cursor-pointer p-1"
          >
            ✕
          </button>
        </div>

        <h3 className="text-lg font-bold text-[var(--ink)] m-0 mb-1 leading-snug">
          {card.title}
        </h3>
        <div className="text-xs text-[var(--mute)] mb-4">
          {card.client_name} · {card.app_name !== "Novo" ? card.app_name : "Novo Projeto"}
        </div>

        {/* Linhas de Valores Financeiros */}
        <div className="bg-[var(--bg)]/50 border border-[var(--line)] rounded-xl p-3 space-y-2 mb-4 text-sm">
          <div className="flex justify-between">
            <span className="text-[var(--mute)]">Valor cobrado</span>
            <b className="text-[var(--ink)]">
              {card.price_cents > 0 ? formatBRL(card.price_cents) : "Sem cobrança"}
            </b>
          </div>
          <div className="flex justify-between">
            <span className="text-[var(--mute)]">Custo estimado</span>
            <b className="text-[var(--ink)]">{formatBRL(card.est_cost_cents)}</b>
          </div>
          <div className="flex justify-between border-t border-[var(--line)] pt-2 font-medium">
            <span className="text-[var(--mute)]">Margem prevista</span>
            <b className={margin_cents >= 0 ? "text-[var(--ok)]" : "text-[var(--bad)]"}>
              {formatBRL(margin_cents)}
            </b>
          </div>
        </div>

        {/* Cobrança PIX Instantânea (Fase 11) */}
        {card.price_cents > 0 && (
          <div className="bg-[var(--surf)] border border-emerald-500/30 rounded-xl p-3.5 mb-4 text-xs">
            <div className="flex items-center justify-between mb-2">
              <b className="text-[var(--ink)] flex items-center gap-1.5 text-sm font-bold">
                <span className="text-emerald-500 font-mono">❖</span> Cobrança via PIX
              </b>
              {pixData && (
                <span className={`text-[10px] font-bold uppercase px-2 py-0.5 rounded ${
                  pixData.status === "paid" 
                    ? "bg-emerald-500/20 text-emerald-500" 
                    : "bg-amber-500/20 text-amber-500"
                }`}>
                  {pixData.status === "paid" ? "Liquidado ✓" : "Pendente"}
                </span>
              )}
            </div>

            {!pixData ? (
              <button
                type="button"
                id="btn-generate-pix"
                disabled={loadingPix}
                onClick={handleGeneratePix}
                className="w-full py-2.5 px-3 rounded-lg font-semibold text-xs border border-emerald-500/40 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-500/20 cursor-pointer flex items-center justify-center gap-2 transition-colors"
              >
                <span>💳</span> {loadingPix ? "Gerando cobrança..." : "Gerar QR Code PIX de " + formatBRL(card.price_cents)}
              </button>
            ) : (
              <div className="space-y-3 pt-1">
                <div className="flex flex-col items-center justify-center bg-white p-3 rounded-xl border border-[var(--line)]">
                  <img
                    src={`https://api.qrserver.com/v1/create-qr-code/?size=160x160&data=${encodeURIComponent(pixData.pix_emv)}`}
                    alt="QR Code PIX"
                    className="w-36 h-36"
                  />
                  <span className="text-[10px] text-gray-500 mt-1 font-mono">TXID: {pixData.pix_txid}</span>
                </div>

                <div className="space-y-1.5">
                  <span className="text-[11px] text-[var(--mute)] block">Código PIX Copia e Cola:</span>
                  <div className="p-2 bg-[var(--bg)] border border-[var(--line)] rounded-lg font-mono text-[10px] break-all select-all max-h-16 overflow-y-auto">
                    {pixData.pix_emv}
                  </div>
                  <button
                    type="button"
                    id="btn-copy-pix"
                    onClick={handleCopyPix}
                    className="w-full py-2 px-3 rounded-lg font-bold text-xs bg-emerald-600 hover:bg-emerald-700 text-white cursor-pointer transition-colors flex items-center justify-center gap-1.5 shadow-sm"
                  >
                    <span>{copiedPix ? "✓" : "📋"}</span>
                    {copiedPix ? "Código PIX Copiado!" : "Copiar Código PIX"}
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Portal Exclusivo do Cliente (Fase 13) */}
        <div className="bg-[var(--surf)] border border-sky-500/30 rounded-xl p-3.5 mb-4 text-xs">
          <div className="flex items-center justify-between mb-2">
            <b className="text-[var(--ink)] flex items-center gap-1.5 text-sm font-bold">
              <span className="text-sky-500">🌐</span> Portal do Cliente
            </b>
            {portalStatus?.is_active && (
              <span className="text-[10px] font-bold uppercase px-2 py-0.5 rounded bg-sky-500/20 text-sky-500">
                Link Ativo
              </span>
            )}
          </div>

          {!portalStatus?.is_active ? (
            <div>
              <p className="text-[11px] text-[var(--mute)] mb-2.5">
                Compartilhe com o cliente um link seguro e criptografado para acompanhamento das etapas e aceite formal do projeto.
              </p>
              <button
                type="button"
                id="btn-generate-portal"
                disabled={loadingPortal}
                onClick={handleGeneratePortal}
                className="w-full py-2.5 px-3 rounded-lg font-semibold text-xs border border-sky-500/40 bg-sky-500/10 text-sky-600 dark:text-sky-400 hover:bg-sky-500/20 cursor-pointer flex items-center justify-center gap-2 transition-colors"
              >
                <span>🔗</span> {loadingPortal ? "Gerando link..." : "Gerar Link do Portal do Cliente"}
              </button>
            </div>
          ) : (
            <div className="space-y-2.5 pt-1">
              {portalStatus.accepted_at ? (
                <div className="p-2.5 bg-emerald-500/10 border border-emerald-500/30 rounded-lg flex items-center gap-2 text-emerald-600 dark:text-emerald-400">
                  <span className="text-base">✓</span>
                  <div>
                    <div className="font-bold text-[11px]">Aceite Formal Registrado pelo Cliente!</div>
                    <div className="text-[10px] opacity-80">
                      Assinado por {portalStatus.accepted_by || "Cliente"} em {new Date(portalStatus.accepted_at).toLocaleString("pt-BR")}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="text-[11px] text-[var(--mute)]">
                  O cliente visualiza o progresso em tempo real e pode realizar o aceite formal na etapa de entrega.
                </div>
              )}

              {portalStatus.last_accessed_at && (
                <div className="text-[10px] text-[var(--mute)]">
                  Último acesso do cliente: {new Date(portalStatus.last_accessed_at).toLocaleString("pt-BR")}
                </div>
              )}

              <div className="flex gap-2">
                <button
                  type="button"
                  id="btn-copy-portal"
                  onClick={handleCopyPortal}
                  className="flex-1 py-2 px-3 rounded-lg font-bold text-xs bg-sky-600 hover:bg-sky-700 text-white cursor-pointer transition-colors flex items-center justify-center gap-1.5 shadow-sm"
                >
                  <span>{copiedPortal ? "✓" : "📋"}</span>
                  {copiedPortal ? "Link do Portal Copiado!" : "Copiar Link do Portal"}
                </button>
                <button
                  type="button"
                  id="btn-revoke-portal"
                  disabled={loadingPortal}
                  onClick={handleRevokePortal}
                  className="py-2 px-3 rounded-lg font-semibold text-xs border border-red-500/30 bg-red-500/10 text-red-500 hover:bg-red-500/20 cursor-pointer transition-colors"
                  title="Revogar link"
                >
                  Revogar
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Pipeline de Fábrica Autônoma (Fases 6 a 9) */}
        {pipelineData && (
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-xl p-3 mb-4 text-xs">
            <div className="flex items-center justify-between mb-2">
              <b className="text-[var(--ink)] flex items-center gap-1.5 text-sm">
                <span className="w-2 h-2 rounded-full bg-[var(--acc)] animate-pulse" />
                Pipeline da Fábrica Autônoma
              </b>
              <span className="text-[10px] font-semibold uppercase px-2 py-0.5 rounded bg-[var(--accs)] text-[var(--acc)]">
                {pipelineData.status}
              </span>
            </div>
            <div className="space-y-1 text-[var(--mute)]">
              <div className="flex justify-between">
                <span>Agente Responsável:</span>
                <b className="text-[var(--ink)] uppercase">{pipelineData.active_role}</b>
              </div>
              <div className="flex justify-between">
                <span>Custo de IA Acumulado:</span>
                <b className="text-[var(--ok)]">{formatBRL(pipelineData.spend_total_cents)}</b>
              </div>
              <div className="pt-1 text-[11px] text-[var(--ink)] italic">
                &quot;{pipelineData.current_task}&quot;
              </div>
            </div>
          </div>
        )}

        {/* Ficha de Risco (quando presente na entrega) */}
        {risk_report && (
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-xl p-3 mb-4 text-xs">
            <b className="text-[var(--ink)] block mb-2 text-sm">Ficha de Risco de Entrega</b>
            <div className="space-y-1.5">
              {risk_report.tests?.map((t, idx) => (
                <div key={idx} className="flex justify-between items-center">
                  <span className="text-[var(--mute)]">{t[0]}</span>
                  <b className={t[2] === "ok" ? "text-[var(--ok)]" : "text-[var(--warn)]"}>
                    {t[1]}
                  </b>
                </div>
              ))}
              {risk_report.secrets_scan?.map((s, idx) => (
                <div key={idx} className="flex justify-between items-center">
                  <span className="text-[var(--mute)]">{s[0]}</span>
                  <b className={s[2] === "ok" ? "text-[var(--ok)]" : "text-[var(--warn)]"}>
                    {s[1]}
                  </b>
                </div>
              ))}
              {risk_report.outdated_deps?.map((d, idx) => (
                <div key={idx} className="flex justify-between items-center">
                  <span className="text-[var(--mute)]">{d[0]}</span>
                  <b className={d[2] === "ok" ? "text-[var(--ok)]" : "text-[var(--warn)]"}>
                    {d[1]}
                  </b>
                </div>
              ))}
              {risk_report.lgpd_points?.map((l, idx) => (
                <div key={idx} className="flex justify-between items-center">
                  <span className="text-[var(--mute)]">{l[0]}</span>
                  <b className={l[2] === "ok" ? "text-[var(--ok)]" : "text-[var(--warn)]"}>
                    {l[1]}
                  </b>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Histórico Cronológico */}
        <div className="mb-5">
          <span className="text-xs font-semibold text-[var(--mute)] uppercase tracking-wider block mb-2">
            Histórico
          </span>
          <ul className="m-0 p-0 pl-3.5 border-l-2 border-[var(--line)] space-y-2 list-none text-xs text-[var(--mute)]">
            {events.length > 0 ? (
              events.map((ev, i) => (
                <li key={i} className="leading-relaxed">
                  <span className="text-[var(--ink)] font-medium">
                    {String(ev.payload?.title || ev.payload?.output || ev.kind)}
                  </span>
                  {Boolean(ev.payload?.note) && <span className="block italic text-[11px]">{String(ev.payload?.note)}</span>}
                </li>
              ))
            ) : (
              <li>Pedido registrado no sistema</li>
            )}
          </ul>
        </div>

        {/* Botões de Ação */}
        <div className="grid gap-2">
          {/* Botão de Construção Autônoma (Fase 6) */}
          {onBuild && card.stage !== "em_execucao" && card.stage !== "entregue" && card.stage !== "recusado" && (
            <button
              id="sheet-build-btn"
              disabled={isProcessing}
              onClick={() => onBuild(card.id)}
              className="w-full py-3 px-4 rounded-xl font-semibold text-sm border-0 bg-gradient-to-r from-[var(--acc)] to-teal-600 text-white cursor-pointer hover:opacity-95 shadow-sm flex items-center justify-center gap-2"
            >
              <span>⚡</span> Construir Projeto com Agentes (Fábrica Autônoma)
            </button>
          )}

          {isDeliveryGate && (
            <label className="flex items-start gap-2 bg-[var(--warns)]/30 border border-[var(--warn)]/40 p-2.5 rounded-xl text-xs cursor-pointer mb-1 text-[var(--ink)]">
              <input
                id="risk-accepted-checkbox"
                type="checkbox"
                checked={riskAccepted}
                onChange={(e) => setRiskAccepted(e.target.checked)}
                className="mt-0.5 accent-[var(--acc)] cursor-pointer"
              />
              <span className="leading-tight">
                Declaro que avaliei a <b>Ficha de Risco</b> acima (testes, segredos, dependências e LGPD) antes de autorizar a entrega ao cliente.
              </span>
            </label>
          )}

          {actions.length > 0 && (
            <button
              id="sheet-approve-btn"
              disabled={isProcessing || (isDeliveryGate && !riskAccepted)}
              onClick={() => onApprove(card.id)}
              className={`w-full py-3 px-4 rounded-xl font-semibold text-sm border-0 transition-opacity ${
                isDeliveryGate && !riskAccepted
                  ? "bg-[var(--line)] text-[var(--mute)] cursor-not-allowed"
                  : "bg-[var(--acc)] text-white cursor-pointer hover:opacity-90"
              }`}
            >
              {actions[0]}
            </button>
          )}

          {card.stage !== "entregue" && card.stage !== "recusado" && (
            <div className="grid grid-cols-2 gap-2">
              <button
                id="sheet-ask-info-btn"
                disabled={isProcessing}
                onClick={() => onAskInfo(card.id)}
                className="py-2.5 px-3 rounded-xl font-semibold text-xs border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] cursor-pointer hover:border-[var(--acc)]"
              >
                Pedir mais informações
              </button>
              <button
                id="sheet-reject-btn"
                disabled={isProcessing}
                onClick={() => onReject(card.id)}
                className="py-2.5 px-3 rounded-xl font-semibold text-xs border border-[var(--line)] bg-[var(--surf)] text-[var(--bad)] cursor-pointer hover:border-[var(--bad)]"
              >
                Recusar
              </button>
            </div>
          )}

          {/* Botão de Proposta Comercial & Contrato (Fase 12) */}
          <button
            type="button"
            id="sheet-proposal-btn"
            disabled={loadingProposal}
            onClick={handleOpenProposal}
            className="w-full py-2.5 px-3 rounded-xl font-bold text-xs border border-indigo-500/40 bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-500/20 cursor-pointer flex items-center justify-center gap-2 transition-colors shadow-sm"
          >
            <span>📄</span> {loadingProposal ? "Abrindo proposta..." : "Visualizar Proposta Comercial & Contrato"}
          </button>

          <button
            id="sheet-close-btn"
            onClick={onClose}
            className="w-full py-2.5 px-4 rounded-xl font-semibold text-xs border border-[var(--line)] bg-transparent text-[var(--mute)] cursor-pointer hover:text-[var(--ink)]"
          >
            Fechar
          </button>
        </div>
      </div>

      {/* Modal de Proposta Comercial e Contrato (Fase 12) */}
      <ProposalModal
        proposal={proposalData}
        isOpen={isProposalOpen}
        onClose={() => setIsProposalOpen(false)}
      />
    </div>
  );
}
