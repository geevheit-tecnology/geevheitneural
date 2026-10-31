"use client";

import React, { useState, useRef, useEffect } from "react";

interface HeaderProps {
  title: string;
  moneyWaitingFormatted: string;
  pendingDecisionsCount: number;
  isPaused: boolean;
  onTogglePause: () => void;
  isTogglingPause: boolean;
  userEmail?: string;
  onLogout?: () => void;
  onNavigateHome?: () => void;
  onNavigateTab?: (tab: "hoje" | "quadro" | "lab" | "apps" | "fin" | "con") => void;
  currentTab?: string;
  isDemoMode?: boolean;
}

export function Header({
  title,
  moneyWaitingFormatted,
  pendingDecisionsCount,
  isPaused,
  onTogglePause,
  isTogglingPause,
  userEmail,
  onLogout,
  onNavigateHome,
  onNavigateTab,
  currentTab,
  isDemoMode = false,
}: HeaderProps) {
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [showLogoutModal, setShowLogoutModal] = useState(false);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  const [isDarkMode, setIsDarkMode] = useState<boolean>(() => {
    if (typeof window !== "undefined") {
      return document.documentElement.getAttribute("data-theme") === "dark";
    }
    return false;
  });

  // Fecha o dropdown ao clicar fora ou ao pressionar ESC
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setShowUserMenu(false);
      }
    }
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setShowUserMenu(false);
        setShowLogoutModal(false);
        setShowConfirmModal(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

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
      <header className="sticky top-0 z-30 bg-[var(--bg)] px-4 pt-3 pb-2 flex items-center justify-between gap-2.5 border-b border-[var(--line)]/50 backdrop-blur-md">
        <div className="flex items-center gap-2.5">
          <img
            src="/icon-192.png"
            alt="GeevheitNeural"
            className="w-9 h-9 rounded-xl border border-[var(--line)] shadow-sm object-cover flex-shrink-0"
          />
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold tracking-tight text-[var(--ink)] m-0 leading-tight">
                {title}
              </h1>
              {currentTab && currentTab !== "hoje" && onNavigateHome && (
                <button
                  id="btn-back-to-home"
                  onClick={onNavigateHome}
                  title="Voltar para a tela inicial"
                  className="text-[11px] font-semibold text-[var(--acc)] bg-[var(--accs)] hover:opacity-80 px-2 py-0.5 rounded-lg border-0 cursor-pointer flex items-center gap-1 transition-all"
                >
                  <span>← Início</span>
                </button>
              )}
            </div>
            <div className="text-xs text-[var(--mute)]">Central de Operação</div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {/* Pílula de Dinheiro Parado */}
          <div
            id="header-money-pill"
            className="bg-[var(--warns)] text-[var(--warn)] rounded-xl px-2.5 py-1 text-xs font-semibold leading-tight text-center max-w-[170px] hidden xs:block"
          >
            {moneyWaitingFormatted} parados esperando você · {pendingDecisionsCount} decisões
          </div>

          {/* Botão de Atalho Rápido para o Laboratório IA */}
          {onNavigateTab && (
            <button
              id="header-lab-btn"
              onClick={() => onNavigateTab("lab")}
              title="Abrir Laboratório de IA"
              className={`px-2.5 py-1.5 rounded-xl text-xs font-bold flex items-center gap-1.5 transition-all cursor-pointer border ${
                currentTab === "lab"
                  ? "bg-purple-600 text-white border-purple-500 shadow-sm"
                  : "bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/30 hover:bg-purple-500/20"
              }`}
            >
              <span className="text-sm">🧪</span>
              <span className="font-semibold">Lab</span>
            </button>
          )}

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

          {/* Menu do Usuário / Dono (Logout) */}
          <div className="relative" ref={menuRef}>
            <button
              id="user-menu-btn"
              onClick={() => setShowUserMenu(!showUserMenu)}
              aria-label="Menu do dono da conta"
              aria-expanded={showUserMenu}
              className="w-8 h-8 rounded-full bg-[var(--acc)] text-white font-bold text-xs flex items-center justify-center cursor-pointer border border-[var(--line)] shadow-sm hover:opacity-90 transition-opacity"
            >
              DN
            </button>

            {/* Dropdown do Menu do Usuário */}
            {showUserMenu && (
              <div
                id="user-dropdown-menu"
                className="absolute right-0 top-10 w-56 bg-[var(--surf)] border border-[var(--line)] rounded-2xl shadow-2xl p-2 z-50 text-xs"
              >
                <div className="px-3 py-2 border-b border-[var(--line)]/60 mb-1">
                  <div className="font-bold text-[var(--ink)] text-sm">Dono Geevheit</div>
                  <div className="text-[11px] text-[var(--mute)] truncate">{userEmail || "dono@geevheit.com"}</div>
                  <div className="mt-1 flex items-center gap-1 text-[10px] text-[var(--ok)] font-medium">
                    <span className="w-1.5 h-1.5 rounded-full bg-[var(--ok)] inline-block"></span>
                    {isDemoMode ? "Demonstração · sem sessão real" : "2FA (AAL2) Ativo"}
                  </div>
                </div>

                {onNavigateHome && (
                  <button
                    id="dropdown-home-btn"
                    onClick={() => {
                      setShowUserMenu(false);
                      onNavigateHome();
                    }}
                    className="w-full text-left px-3 py-2 rounded-xl text-[var(--ink)] hover:bg-[var(--accs)] flex items-center gap-2 font-medium cursor-pointer border-0 bg-transparent"
                  >
                    <span>🏠</span> Início (Hub de Quadros)
                  </button>
                )}

                {onNavigateTab && (
                  <button
                    id="dropdown-lab-btn"
                    onClick={() => {
                      setShowUserMenu(false);
                      onNavigateTab("lab");
                    }}
                    className="w-full text-left px-3 py-2 rounded-xl text-purple-600 dark:text-purple-400 hover:bg-purple-500/10 flex items-center gap-2 font-semibold cursor-pointer border-0 bg-transparent transition-colors"
                  >
                    <span>🧪</span> Laboratório de IA
                  </button>
                )}

                <button
                  id="dropdown-clear-cache-btn"
                  onClick={async () => {
                    try {
                      if ("caches" in window) {
                        const keys = await caches.keys();
                        await Promise.all(keys.map((k) => caches.delete(k)));
                      }
                      if ("serviceWorker" in navigator) {
                        const regs = await navigator.serviceWorker.getRegistrations();
                        for (const reg of regs) {
                          await reg.unregister();
                        }
                      }
                    } catch {}
                    window.location.reload();
                  }}
                  className="w-full text-left px-3 py-2 rounded-xl text-[var(--mute)] hover:text-[var(--ink)] hover:bg-[var(--line)]/30 flex items-center gap-2 font-medium cursor-pointer border-0 bg-transparent transition-colors"
                >
                  <span>🔄</span> Atualizar App (Limpar Cache)
                </button>

                <div className="border-t border-[var(--line)]/60 my-1"></div>

                <button
                  id="logout-btn"
                  onClick={() => {
                    setShowUserMenu(false);
                    setShowLogoutModal(true);
                  }}
                  className="w-full text-left px-3 py-2 rounded-xl text-[var(--bad)] hover:bg-[var(--bad)]/10 flex items-center gap-2 font-semibold cursor-pointer border-0 bg-transparent transition-colors"
                >
                  <svg className="w-4 h-4 stroke-current fill-none stroke-[2] stroke-round stroke-linejoin" viewBox="0 0 24 24">
                    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                    <polyline points="16 17 21 12 16 7" />
                    <line x1="21" y1="12" x2="9" y2="12" />
                  </svg>
                  <span>Sair da conta</span>
                </button>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* Modal de Confirmação de Logout Seguro */}
      {showLogoutModal && (
        <div
          id="confirm-logout-modal"
          className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4 backdrop-blur-sm"
        >
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 max-w-sm w-full shadow-2xl">
            <div className="w-10 h-10 rounded-2xl bg-[var(--bad)]/10 text-[var(--bad)] flex items-center justify-center mb-3">
              <svg className="w-5 h-5 stroke-current fill-none stroke-[2] stroke-round stroke-linejoin" viewBox="0 0 24 24">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                <polyline points="16 17 21 12 16 7" />
                <line x1="21" y1="12" x2="9" y2="12" />
              </svg>
            </div>
            <h3 className="text-base font-bold text-[var(--ink)] mb-1">
              Deseja realmente sair?
            </h3>
            <p className="text-xs text-[var(--mute)] mb-4 leading-relaxed">
              Sua sessão segura será encerrada. Todas as alterações e registros já estão sincronizados com segurança no banco de dados.
            </p>
            <div className="flex justify-end gap-2">
              <button
                id="cancel-logout-btn"
                onClick={() => setShowLogoutModal(false)}
                className="px-4 py-2 rounded-xl text-xs font-semibold border border-[var(--line)] bg-transparent text-[var(--ink)] cursor-pointer hover:bg-[var(--line)]/30"
              >
                Cancelar
              </button>
              <button
                id="confirm-logout-btn"
                onClick={() => {
                  setShowLogoutModal(false);
                  if (onLogout) onLogout();
                }}
                className="px-4 py-2 rounded-xl text-xs font-semibold border-0 text-white bg-[var(--bad)] cursor-pointer hover:opacity-90"
              >
                Sim, Sair
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal de Confirmação de Parada / Retomada */}
      {showConfirmModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4 backdrop-blur-sm">
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
