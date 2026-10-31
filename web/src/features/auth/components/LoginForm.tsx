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
  const [step, setStep] = useState<"credentials" | "totp" | "enroll">("credentials");
  const [enrollment, setEnrollment] = useState<{ qrCode: string; secret: string } | null>(null);
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
      if (factorsRes.error) throw factorsRes.error;
      setPassword("");
      const totpFactors = (factorsRes.data?.totp || []).filter((factor) => factor.status === "verified");

      if (totpFactors.length > 0) {
        // Exige o segundo fator (AAL2)
        setFactorId(totpFactors[0].id);
        setStep("totp");
      } else {
        const result = await supabase.auth.mfa.enroll({
          factorType: "totp",
          friendlyName: `GeevheitNeural ${Date.now()}`,
          issuer: "GeevheitNeural",
        });
        if (result.error) throw result.error;
        setFactorId(result.data.id);
        setEnrollment({ qrCode: result.data.totp.qr_code, secret: result.data.totp.secret });
        setStep("enroll");
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
    if (!/^\d{6}$/.test(totpCode)) {
      setErrorMsg("Digite os 6 números do aplicativo autenticador.");
      return;
    }

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

      // O segredo de cadastro permanece somente em memória até a confirmação.
      setEnrollment(null);
      setTotpCode("");
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
            {step === "enroll" && enrollment && (
              <section id="totp-enrollment" className="space-y-3">
                <h2 className="text-sm font-bold">Configure seu autenticador</h2>
                <p className="text-xs text-[var(--mute)]">
                  No Google Authenticator ou Microsoft Authenticator, adicione uma conta e leia este QR Code.
                  Depois, informe abaixo o código gerado pelo aplicativo.
                </p>
                <img
                  id="totp-qr-code"
                  src={enrollment.qrCode}
                  alt="QR Code para cadastrar o autenticador GeevheitNeural"
                  className="w-48 max-w-full mx-auto rounded-lg bg-white p-2"
                />
                <details className="text-xs text-[var(--mute)]">
                  <summary className="cursor-pointer">Não consigo ler o QR Code</summary>
                  <p className="my-2">No mesmo celular, escolha inserir uma chave de configuração no autenticador e copie esta chave. Não a compartilhe.</p>
                  <input id="totp-setup-key" aria-label="Chave de configuração do autenticador" readOnly value={enrollment.secret} className="w-full p-2 rounded-lg border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] font-mono select-all" />
                </details>
              </section>
            )}
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
                autoComplete="one-time-code"
                pattern="[0-9]{6}"
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
              {loading ? "Validando..." : step === "enroll" ? "Ativar 2FA e entrar" : "Confirmar e Entrar"}
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
              <span>Explorar demonstração (sem executar ordens)</span>
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
