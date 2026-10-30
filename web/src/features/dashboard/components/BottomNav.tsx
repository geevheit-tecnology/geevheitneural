"use client";

import React from "react";

export type TabKey = "hoje" | "quadro" | "lab" | "apps" | "fin" | "con";

interface BottomNavProps {
  currentTab: TabKey;
  onTabChange: (tab: TabKey) => void;
}

export function BottomNav({ currentTab, onTabChange }: BottomNavProps) {
  const tabs = [
    {
      key: "hoje" as TabKey,
      label: "Início",
      icon: (
        <>
          <path d="M3 9.5L12 3l9 6.5V20a1.5 1.5 0 0 1-1.5 1.5H4.5A1.5 1.5 0 0 1 3 20V9.5z" />
          <polyline points="9 21 9 12 15 12 15 21" />
        </>
      ),
    },
    {
      key: "quadro" as TabKey,
      label: "Quadro",
      icon: <path d="M4 4h5v16H4zM10 4h5v10h-5zM16 4h4v7h-4z" />,
    },
    {
      key: "lab" as TabKey,
      label: "Lab",
      isSpecial: true,
      icon: (
        <>
          <path d="M9 3h6M10 3v5l-4 7a3 3 0 0 0 2.6 4.5h6.8A3 3 0 0 0 18 15l-4-7V3" />
          <circle cx="12" cy="14" r="1.5" />
        </>
      ),
    },
    {
      key: "apps" as TabKey,
      label: "Apps",
      icon: (
        <>
          <rect x="4" y="4" width="7" height="7" rx="1.5" />
          <rect x="13" y="4" width="7" height="7" rx="1.5" />
          <rect x="4" y="13" width="7" height="7" rx="1.5" />
          <rect x="13" y="13" width="7" height="7" rx="1.5" />
        </>
      ),
    },
    {
      key: "fin" as TabKey,
      label: "Finanças",
      icon: <path d="M4 20V10M10 20V4M16 20v-7M21 20H3" />,
    },
    {
      key: "con" as TabKey,
      label: "Conexões",
      icon: <path d="M9 7V3m6 4V3M7 7h10v4a5 5 0 0 1-10 0zM12 16v5" />,
    },
  ];

  return (
    <nav
      id="bottom-nav"
      className="fixed left-0 right-0 bottom-0 bg-[var(--surf)] border-t border-[var(--line)] flex items-center justify-between px-1 py-1 z-40 overflow-x-hidden shadow-lg backdrop-blur-md"
      style={{ paddingBottom: "calc(6px + env(safe-area-inset-bottom, 0px))" }}
    >
      {tabs.map((tab) => {
        const isActive = currentTab === tab.key;
        const isLab = tab.key === "lab";

        let activeClasses = "text-[var(--mute)] hover:text-[var(--ink)]";
        if (isActive) {
          activeClasses = isLab
            ? "text-purple-600 dark:text-purple-400 font-bold bg-purple-500/15 border border-purple-500/20"
            : "text-[var(--acc)] font-semibold bg-[var(--accs)]";
        } else if (isLab) {
          activeClasses = "text-purple-600/80 dark:text-purple-400/80 hover:text-purple-600 font-medium";
        }

        return (
          <button
            key={tab.key}
            id={`nav-btn-${tab.key}`}
            onClick={() => onTabChange(tab.key)}
            className={`flex-1 min-w-0 max-w-full flex flex-col items-center justify-center gap-0.5 py-1 px-0.5 rounded-xl transition-all border-0 bg-transparent cursor-pointer select-none ${activeClasses}`}
          >
            <div className="relative flex items-center justify-center">
              <svg
                className="w-5 h-5 flex-shrink-0 stroke-current fill-none stroke-[1.8] stroke-round stroke-linejoin"
                viewBox="0 0 24 24"
              >
                {tab.icon}
              </svg>
              {isLab && !isActive && (
                <span className="absolute -top-0.5 -right-1 w-1.5 h-1.5 rounded-full bg-purple-500" />
              )}
            </div>
            <span className="text-[10px] sm:text-xs tracking-tight truncate max-w-full text-center leading-none mt-0.5 font-medium">
              {tab.label}
            </span>
          </button>
        );
      })}
    </nav>
  );
}
