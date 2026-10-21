"use client";

import React, { useState } from "react";
import { ProposalDocument } from "@/shared/api-client/endpoints";

interface ProposalModalProps {
  proposal: ProposalDocument | null;
  isOpen: boolean;
  onClose: () => void;
}

export function ProposalModal({ proposal, isOpen, onClose }: ProposalModalProps) {
  const [copiedWhatsApp, setCopiedWhatsApp] = useState(false);
  const [copiedPix, setCopiedPix] = useState(false);

  if (!isOpen || !proposal) return null;

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 2,
    });
  };

  const handleCopyWhatsApp = () => {
    if (proposal.formatted_whatsapp) {
      navigator.clipboard.writeText(proposal.formatted_whatsapp);
      setCopiedWhatsApp(true);
      setTimeout(() => setCopiedWhatsApp(false), 3000);
    }
  };

  const handleCopyPix = () => {
    if (proposal.pix_emv) {
      navigator.clipboard.writeText(proposal.pix_emv);
      setCopiedPix(true);
      setTimeout(() => setCopiedPix(false), 3000);
    }
  };

  const handlePrint = () => {
    window.print();
  };

  return (
    <div
      id="proposal-modal-backdrop"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-3 sm:p-6 overflow-y-auto"
    >
      <div
        id="proposal-modal-content"
        className="bg-[var(--surf)] border border-[var(--line)] w-full max-w-2xl rounded-3xl p-6 sm:p-8 shadow-2xl relative max-h-[90vh] overflow-y-auto print:max-h-none print:shadow-none print:border-none print:p-0 print:bg-white print:text-black"
      >
        {/* Barra superior de ações (escondida na impressão) */}
        <div className="flex items-center justify-between pb-4 border-b border-[var(--line)] mb-6 print:hidden">
          <div className="flex items-center gap-2">
            <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-[var(--accs)] text-[var(--acc)]">
              Proposta Comercial Oficial
            </span>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              id="btn-print-proposal"
              onClick={handlePrint}
              className="px-3 py-1.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] hover:border-[var(--acc)] text-xs font-semibold cursor-pointer transition-colors flex items-center gap-1.5"
              title="Imprimir ou Salvar em PDF"
            >
              <span>🖨️</span> PDF
            </button>
            <button
              type="button"
              onClick={onClose}
              className="text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0 text-xl cursor-pointer p-1"
            >
              ✕
            </button>
          </div>
        </div>

        {/* Timbrado da Software House Geevheit */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-[var(--line)]">
          <div className="flex items-center gap-3">
            <img
              src="/logo.png"
              alt="Geevheit"
              className="w-12 h-12 rounded-2xl border border-[var(--line)] shadow-sm object-cover"
            />
            <div>
              <h2 className="text-lg font-bold text-[var(--ink)] m-0 leading-tight print:text-black">
                Geevheit Software House
              </h2>
              <div className="text-xs text-[var(--mute)] print:text-gray-600">
                Engenharia de Software, IA & Soluções Digitais
              </div>
            </div>
          </div>

          <div className="text-left sm:text-right text-xs text-[var(--mute)] print:text-gray-600 space-y-0.5">
            <div>Data: {new Date(proposal.generated_at).toLocaleDateString("pt-BR")}</div>
            <div className="font-mono text-[10px]">ID: {proposal.id.substring(0, 8).toUpperCase()}</div>
          </div>
        </div>

        {/* Cabeçalho do Projeto & Cliente */}
        <div className="my-6 space-y-1">
          <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--acc)]">
            Proposta de Desenvolvimento
          </span>
          <h1 className="text-2xl font-extrabold text-[var(--ink)] leading-snug print:text-black">
            {proposal.title}
          </h1>
          <div className="text-sm text-[var(--mute)] print:text-gray-700">
            Cliente: <b className="text-[var(--ink)] print:text-black">{proposal.client_name}</b> · Aplicação: {proposal.app_name}
          </div>
        </div>

        {/* Resumo do Escopo */}
        <div className="bg-[var(--bg)]/70 border border-[var(--line)] rounded-2xl p-4 mb-6 print:border-gray-300 print:bg-gray-50">
          <h4 className="text-xs font-bold uppercase tracking-wider text-[var(--mute)] mb-2 print:text-gray-600">
            Escopo Técnico
          </h4>
          <p className="text-sm text-[var(--ink)] leading-relaxed m-0 print:text-black">
            {proposal.scope_summary}
          </p>
        </div>

        {/* Entregáveis Estruturados */}
        <div className="mb-6 space-y-3">
          <h4 className="text-xs font-bold uppercase tracking-wider text-[var(--mute)] print:text-gray-600">
            Principais Entregas & Especificações
          </h4>
          <div className="grid gap-2">
            {proposal.deliverables.map((item, idx) => (
              <div
                key={idx}
                className="flex items-start gap-2.5 text-xs text-[var(--ink)] p-2.5 rounded-xl border border-[var(--line)] bg-[var(--surf)] print:border-gray-200 print:text-black"
              >
                <span className="text-emerald-500 font-bold">✓</span>
                <span className="leading-tight">{item}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Prazos & Garantia Técnica */}
        <div className="grid grid-cols-2 gap-3 mb-6">
          <div className="p-3 rounded-2xl border border-[var(--line)] bg-[var(--bg)]/50 print:border-gray-300">
            <span className="text-[11px] text-[var(--mute)] block print:text-gray-600">Prazo de Execução</span>
            <b className="text-base text-[var(--ink)] print:text-black">
              {proposal.timeline_days} dias úteis
            </b>
            <span className="text-[10px] text-[var(--mute)] block mt-0.5">Após aprovação e sinal</span>
          </div>
          <div className="p-3 rounded-2xl border border-[var(--line)] bg-[var(--bg)]/50 print:border-gray-300">
            <span className="text-[11px] text-[var(--mute)] block print:text-gray-600">Garantia Técnica</span>
            <b className="text-base text-emerald-500 print:text-emerald-700">
              {proposal.warranty_days} dias corridos
            </b>
            <span className="text-[10px] text-[var(--mute)] block mt-0.5">Suporte e correções inclusos</span>
          </div>
        </div>

        {/* Condições Financeiras e Investimento */}
        <div className="bg-gradient-to-br from-[var(--accs)]/40 to-transparent border border-[var(--acc)]/30 rounded-2xl p-5 mb-6 print:border-gray-300 print:bg-gray-50">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3">
            <div>
              <span className="text-xs font-bold text-[var(--acc)] uppercase tracking-wider">
                Investimento Total
              </span>
              <div className="text-2xl font-black text-[var(--ink)] print:text-black">
                {proposal.price_cents > 0 ? formatBRL(proposal.price_cents) : "Sob Consulta"}
              </div>
            </div>
            <div className="text-xs text-[var(--mute)] sm:text-right print:text-gray-600">
              Condição Comercial:
              <div className="font-semibold text-[var(--ink)] print:text-black">{proposal.payment_terms}</div>
            </div>
          </div>

          {/* Dados de Pagamento PIX (se disponível) */}
          {proposal.pix_emv && (
            <div className="pt-3 border-t border-[var(--line)] space-y-2">
              <span className="text-[11px] font-semibold text-[var(--ink)] flex items-center gap-1.5 print:text-black">
                <span>⚡</span> Pagamento Facilitado via PIX
              </span>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  readOnly
                  value={proposal.pix_emv}
                  className="w-full text-[10px] font-mono px-3 py-1.5 rounded-lg border border-[var(--line)] bg-[var(--bg)] text-[var(--mute)] select-all truncate print:hidden"
                />
                <button
                  type="button"
                  id="btn-copy-proposal-pix"
                  onClick={handleCopyPix}
                  className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs cursor-pointer whitespace-nowrap transition-colors print:hidden"
                >
                  {copiedPix ? "✓ Copiado" : "Copiar PIX"}
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Termos e Condições */}
        <div className="mb-6 text-[11px] text-[var(--mute)] space-y-1.5 leading-relaxed border-t border-[var(--line)] pt-4 print:text-gray-600">
          <b className="text-xs text-[var(--ink)] block mb-1 print:text-black">Termos Contratuais & Aceite</b>
          <div className="whitespace-pre-line font-normal">{proposal.terms_and_conditions}</div>
        </div>

        {/* Ações de Rodapé (1 Toque) */}
        <div className="pt-4 border-t border-[var(--line)] flex flex-col sm:flex-row gap-2.5 print:hidden">
          <button
            type="button"
            id="btn-copy-proposal-whatsapp"
            onClick={handleCopyWhatsApp}
            className="flex-1 py-3 px-4 rounded-xl font-bold text-xs bg-emerald-600 hover:bg-emerald-700 text-white cursor-pointer transition-all shadow-md flex items-center justify-center gap-2"
          >
            <span>📱</span> {copiedWhatsApp ? "Copiado para WhatsApp! ✓" : "Copiar Proposta para WhatsApp"}
          </button>

          <button
            type="button"
            onClick={handlePrint}
            className="py-3 px-4 rounded-xl font-semibold text-xs border border-[var(--line)] bg-[var(--surf)] hover:border-[var(--acc)] text-[var(--ink)] cursor-pointer transition-colors flex items-center justify-center gap-1.5"
          >
            <span>🖨️</span> Salvar PDF
          </button>

          <button
            type="button"
            id="btn-close-proposal-modal"
            onClick={onClose}
            className="py-3 px-4 rounded-xl font-semibold text-xs border border-[var(--line)] bg-transparent text-[var(--mute)] hover:text-[var(--ink)] cursor-pointer"
          >
            Fechar
          </button>
        </div>
      </div>
    </div>
  );
}
