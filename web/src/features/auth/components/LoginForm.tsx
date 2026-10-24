"use client";

import { useState } from "react";
import { supabase } from "@/shared/auth/supabase";

interface LoginFormProps {
  onSuccess?: () => void;
  onDemoLogin?: () => void;
}

export function LoginForm({ onSuccess, onDemoLogin }: LoginFormProps) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [totpCode, setTotpCode] = useState("");
  const [factorId, setFactorId] = useState<string | null>(null);
  const [step, setStep] = useState<"credentials" | "totp">("credentials");
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  async function handlePasswordSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    try {
      const { error } = await supabase.auth.signInWithPassword({
        email,
        password,
      });

      if (error) {
        throw error;
      }

      // Verifica se o usuário possui fatores MFA TOTP cadastrados
      const factorsRes = await supabase.auth.mfa.listFactors();
      const totpFactors = factorsRes.data?.totp || [];

      if (totpFactors.length > 0) {
        // Exige o segundo fator (AAL2)
        setFactorId(totpFactors[0].id);
        setStep("totp");
      } else {
        // Sessão iniciada (caso TOTP ainda não esteja cadastrado na Fase 1)
        if (onSuccess) onSuccess();
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Credenciais incorretas";
      setErrorMsg(msg);
    } finally {
      setLoading(false);
    }
  }

  async function handleTOTPSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!factorId) return;

    setErrorMsg(null);
    setLoading(true);

    try {
      const challenge = await supabase.auth.mfa.challenge({ factorId });
      if (challenge.error) throw challenge.error;

      const verify = await supabase.auth.mfa.verify({
        factorId,
        challengeId: challenge.data.id,
        code: totpCode,
      });

      if (verify.error) throw verify.error;

      // Autenticação com AAL2 concluída com sucesso
      if (onSuccess) onSuccess();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Código TOTP inválido ou expirado";
      setErrorMsg(msg);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="w-full max-w-[360px] mx-auto p-4">
      <div className="text-center mb-6">
        <img
          src="/logo.png"
          alt="GeevheitNeural"
          className="w-16 h-16 mx-auto mb-3 rounded-2xl shadow-md border border-[var(--line)] object-cover"
        />
        <h1 className="text-xl font-bold tracking-tight">GeevheitNeural</h1>
        <p className="text-xs text-[var(--mute)] mt-1">
          Acesso exclusivo do dono (2FA obrigatório)
        </p>
      </div>

      <div className="box">
        {errorMsg && (
          <div
            id="login-error"
            className="mb-4 p-3 rounded-lg text-xs font-semibold bg-[var(--warns)] text-[var(--bad)]"
          >
            {errorMsg}
          </div>
        )}

        {step === "credentials" ? (
          <form onSubmit={handlePasswordSubmit} className="flex flex-col gap-3">
            <div>
              <label
                htmlFor="email"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                E-mail
              </label>
              <input
                id="email"
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="dono@geevheit.com"
                className="w-full px-3 py-2 text-sm rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              />
            </div>

            <div>
              <label
                htmlFor="password"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Senha
              </label>
              <input
                id="password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full px-3 py-2 text-sm rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              />
            </div>

            <button
              id="btn-login-submit"
              type="submit"
              disabled={loading}
              className="btn btn-primary mt-2 flex justify-center items-center font-bold"
            >
              {loading ? "Entrando..." : "Continuar"}
            </button>
          </form>
        ) : (
          <form onSubmit={handleTOTPSubmit} className="flex flex-col gap-3">
            <div className="text-xs text-[var(--mute)] mb-1">
              Digite o código de 6 dígitos do seu app autenticador:
            </div>
            <div>
              <label
                htmlFor="totp-code"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Código 2FA
              </label>
              <input
                id="totp-code"
                type="text"
                inputMode="numeric"
                maxLength={6}
                required
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
                placeholder="123456"
                className="w-full px-3 py-2 text-lg tracking-widest text-center font-mono rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              />
            </div>

            <button
              id="btn-verify-totp"
              type="submit"
              disabled={loading}
              className="btn btn-primary mt-2 flex justify-center items-center font-bold"
            >
              {loading ? "Validando..." : "Confirmar e Entrar"}
            </button>
          </form>
        )}

        {onDemoLogin && (
          <div className="mt-4 pt-3 border-t border-[var(--line)]">
            <button
              id="btn-demo-login"
              type="button"
              onClick={onDemoLogin}
              className="w-full py-2.5 px-3 rounded-lg border border-[var(--acc)] text-[var(--acc)] hover:bg-[var(--acc)]/10 font-bold text-xs flex items-center justify-center gap-2 transition-colors"
            >
              <span>⚡</span>
              <span>Acessar Central (Acesso Direto do Dono)</span>
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
