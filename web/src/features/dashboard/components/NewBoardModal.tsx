"use client";

import React, { useState } from "react";
import { CreateAppInput } from "@/shared/api-client/endpoints";

interface NewBoardModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreateBoard: (data: CreateAppInput) => void;
  isProcessing?: boolean;
}

export function NewBoardModal({
  isOpen,
  onClose,
  onCreateBoard,
  isProcessing = false,
}: NewBoardModalProps) {
  const [name, setName] = useState("");
  const [hosting, setHosting] = useState("Render");
  const [database, setDatabase] = useState("Supabase");
  const [domain, setDomain] = useState("");
  const [repository, setRepository] = useState("");

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    onCreateBoard({
      name: name.trim(),
      hosting,
      database,
      domain: domain.trim() || undefined,
      repository: repository.trim() || undefined,
    });

    setName("");
    setDomain("");
    setRepository("");
    onClose();
  };

  return (
    <div
      id="modal-new-board"
      className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4 backdrop-blur-sm"
    >
      <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 max-w-md w-full shadow-2xl">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <span className="text-xl">📋</span>
            <h3 className="text-lg font-bold text-[var(--ink)] m-0">Criar Novo Quadro</h3>
          </div>
          <button
            onClick={onClose}
            aria-label="Fechar modal"
            className="text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0 cursor-pointer text-lg leading-none p-1"
          >
            ✕
          </button>
        </div>

        <p className="text-xs text-[var(--mute)] mb-4 leading-relaxed">
          Cada quadro representa um projeto ou aplicativo operado pelos seus agentes e fluxos de desenvolvimento.
        </p>

        <form onSubmit={handleSubmit} className="space-y-3.5">
          <div>
            <label
              htmlFor="input-board-name"
              className="block text-xs font-semibold text-[var(--mute)] mb-1"
            >
              Nome do Quadro / Projeto *
            </label>
            <input
              id="input-board-name"
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Ex: portal-cliente, crm-imoveis, api-pagamentos"
              className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)] transition-colors"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="select-board-hosting"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Hospedagem
              </label>
              <select
                id="select-board-hosting"
                value={hosting}
                onChange={(e) => setHosting(e.target.value)}
                className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              >
                <option value="Render">Render</option>
                <option value="Vercel">Vercel</option>
                <option value="Netlify">Netlify</option>
                <option value="Firebase">Firebase</option>
                <option value="Docker (VPS)">Docker (VPS)</option>
                <option value="Cloudflare">Cloudflare</option>
              </select>
            </div>

            <div>
              <label
                htmlFor="select-board-database"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Banco de Dados
              </label>
              <select
                id="select-board-database"
                value={database}
                onChange={(e) => setDatabase(e.target.value)}
                className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              >
                <option value="Supabase">Supabase</option>
                <option value="Neon">Neon</option>
                <option value="Firebase">Firebase</option>
                <option value="Postgres local">Postgres local</option>
                <option value="SQLite">SQLite</option>
                <option value="—">Sem banco (Estático)</option>
              </select>
            </div>
          </div>

          <div>
            <label
              htmlFor="input-board-domain"
              className="block text-xs font-semibold text-[var(--mute)] mb-1"
            >
              Domínio (opcional)
            </label>
            <input
              id="input-board-domain"
              type="text"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              placeholder="Ex: app.cliente.com.br"
              className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
            />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button
              id="btn-create-board-cancel"
              type="button"
              onClick={onClose}
              disabled={isProcessing}
              className="px-4 py-2 rounded-xl text-xs font-semibold border border-[var(--line)] bg-transparent text-[var(--ink)] cursor-pointer hover:bg-[var(--line)]/30"
            >
              Cancelar
            </button>
            <button
              id="btn-create-board-submit"
              type="submit"
              disabled={isProcessing || !name.trim()}
              className="px-4 py-2 rounded-xl text-xs font-semibold border-0 text-white bg-[var(--acc)] cursor-pointer hover:opacity-90 disabled:opacity-50 transition-opacity flex items-center gap-1.5"
            >
              {isProcessing ? "Criando..." : "Criar Quadro"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
