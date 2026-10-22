"use client";

import React from "react";

export type TabKey = "hoje" | "quadro" | "apps" | "fin" | "con";

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
      key: "apps" as TabKey,
      label: "Meus apps",
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
      className="fixed left-0 right-0 bottom-0 bg-[var(--surf)] border-t border-[var(--line)] flex px-2 py-1.5 z-40"
      style={{ paddingBottom: "calc(6px + env(safe-area-inset-bottom, 0px))" }}
    >
      {tabs.map((tab) => {
        const isActive = currentTab === tab.key;
        return (
          <button
            key={tab.key}
            id={`nav-btn-${tab.key}`}
            onClick={() => onTabChange(tab.key)}
            className={`flex-1 flex flex-col items-center justify-center gap-0.5 py-1.5 px-0.5 rounded-xl transition-colors text-xs border-0 bg-transparent cursor-pointer ${
              isActive
                ? "text-[var(--acc)] font-semibold bg-[var(--accs)]"
                : "text-[var(--mute)] hover:text-[var(--ink)]"
            }`}
          >
            <svg
              className="w-5 h-5 stroke-current fill-none stroke-[1.8] stroke-round stroke-linejoin"
              viewBox="0 0 24 24"
            >
              {tab.icon}
            </svg>
            <span>{tab.label}</span>
          </button>
        );
      })}
    </nav>
  );
}
