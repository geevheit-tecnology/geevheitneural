"use client";

import React from "react";

interface PausedBannerProps {
  isPaused: boolean;
  onResume: () => void;
}

export function PausedBanner({ isPaused, onResume }: PausedBannerProps) {
  if (!isPaused) return null;

  return (
    <div
      id="paused-system-banner"
      className="bg-[var(--warns)] text-[var(--warn)] border border-[var(--warn)]/20 rounded-xl p-3 mx-4 mb-3 text-xs sm:text-sm font-semibold flex items-center justify-between gap-2 shadow-sm"
    >
      <span>
        ⚠️ Tudo pausado: nenhum agente nem conexão pode agir. Toque em Retomar para liberar.
      </span>
      <button
        id="resume-banner-btn"
        onClick={onResume}
        className="px-2.5 py-1 bg-[var(--warn)] text-white font-bold rounded-lg text-xs border-0 cursor-pointer hover:opacity-90 flex-none"
      >
        Retomar
      </button>
    </div>
  );
}
