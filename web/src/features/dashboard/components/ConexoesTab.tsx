"use client";

import React from "react";
import { ConnectorItem } from "@/shared/api-client/endpoints";

interface ConexoesTabProps {
  connectors: ConnectorItem[];
  onShowToast: (msg: string) => void;
}

export function ConexoesTab({ connectors, onShowToast }: ConexoesTabProps) {
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

  return (
    <div id="tab-conexoes" className="space-y-3">
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

            {/* Switch Visual */}
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
  );
}
