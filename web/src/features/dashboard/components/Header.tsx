"use client";

import React, { useState } from "react";

interface HeaderProps {
  title: string;
  moneyWaitingFormatted: string;
  pendingDecisionsCount: number;
  isPaused: boolean;
  onTogglePause: () => void;
  isTogglingPause: boolean;
}

export function Header({
  title,
  moneyWaitingFormatted,
  pendingDecisionsCount,
  isPaused,
  onTogglePause,
  isTogglingPause,
}: HeaderProps) {
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [isDarkMode, setIsDarkMode] = useState<boolean>(() => {
    if (typeof window !== "undefined") {
      return document.documentElement.getAttribute("data-theme") === "dark";
    }
    return false;
  });

  const toggleTheme = () => {
    const nextDark = !isDarkMode;
    setIsDarkMode(nextDark);
    const themeName = nextDark ? "dark" : "light";
    document.documentElement.setAttribute("data-theme", themeName);
    try {
      localStorage.setItem("geev_theme", themeName);
    } catch {}
  };

  return (
    <>
      <header className="sticky top-0 z-30 bg-[var(--bg)] px-4 pt-3 pb-2 flex items-center justify-between gap-2.5">
        <div className="flex items-center gap-2.5">
          <img
            src="/icon-192.png"
            alt="GeevheitNeural"
            className="w-9 h-9 rounded-xl border border-[var(--line)] shadow-sm object-cover flex-shrink-0"
          />
          <div>
            <h1 className="text-xl font-bold tracking-tight text-[var(--ink)] m-0 leading-tight">
              {title}
            </h1>
            <div className="text-xs text-[var(--mute)]">Central de Operação</div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {/* Pílula de Dinheiro Parado */}
          <div
            id="header-money-pill"
            className="bg-[var(--warns)] text-[var(--warn)] rounded-xl px-2.5 py-1 text-xs font-semibold leading-tight text-center max-w-[170px]"
          >
            {moneyWaitingFormatted} parados esperando você · {pendingDecisionsCount} decisões
          </div>

          {/* Alternador de Tema */}
          <button
            id="theme-toggle-btn"
            onClick={toggleTheme}
            aria-label="Alternar tema claro e escuro"
            className="w-8 h-8 rounded-full border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] flex items-center justify-center cursor-pointer text-sm transition-colors hover:border-[var(--acc)]"
          >
            {isDarkMode ? "☀️" : "🌙"}
          </button>

          {/* Botão SOS de Parada / Retomada */}
          <button
            id="sos-action-btn"
            onClick={() => setShowConfirmModal(true)}
            disabled={isTogglingPause}
            className={`font-bold text-xs border-0 rounded-full py-2 px-3 text-white cursor-pointer transition-all ${
              isPaused
                ? "bg-[var(--ok)] hover:opacity-90"
                : "bg-[var(--bad)] hover:opacity-90"
            }`}
          >
            {isPaused ? "Retomar" : "Parar"}
          </button>
        </div>
      </header>

      {/* Modal de Confirmação de Parada / Retomada */}
      {showConfirmModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4">
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 max-w-sm w-full shadow-2xl">
            <h3 className="text-lg font-bold text-[var(--ink)] mb-2">
              {isPaused ? "Retomar o sistema?" : "Parar todo o sistema de emergência?"}
            </h3>
            <p className="text-sm text-[var(--mute)] mb-5 leading-relaxed">
              {isPaused
                ? "Ao retomar, execuções de cartões aprovados voltam a ser permitidas. Os conectores seguem desligados por padrão."
                : "Ao parar, nenhuma IA ou conector poderá agir ou executar tarefas no sistema até você retomar. Deseja confirmar?"}
            </p>
            <div className="flex justify-end gap-2">
              <button
                id="cancel-sos-btn"
                onClick={() => setShowConfirmModal(false)}
                className="px-4 py-2 rounded-xl text-sm font-semibold border border-[var(--line)] bg-transparent text-[var(--ink)] cursor-pointer"
              >
                Cancelar
              </button>
              <button
                id="confirm-sos-btn"
                onClick={() => {
                  setShowConfirmModal(false);
                  onTogglePause();
                }}
                className={`px-4 py-2 rounded-xl text-sm font-semibold border-0 text-white cursor-pointer ${
                  isPaused ? "bg-[var(--ok)]" : "bg-[var(--bad)]"
                }`}
              >
                {isPaused ? "Sim, Retomar" : "Sim, Parar Agora"}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
