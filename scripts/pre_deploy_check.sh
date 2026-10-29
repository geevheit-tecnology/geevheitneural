#!/usr/bin/env bash
set -euo pipefail

echo "=================================================="
echo "🚀 GeevheitNeural — Verificação Pré-Deploy em Produção"
echo "=================================================="

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "1. Verificando segredos no repositório Git..."
if git ls-files | grep -E '\.env$|\.env\.local$' > /dev/null; then
  echo "❌ ERRO: Arquivos .env reais estão rastreados no git!"
  exit 1
fi
echo "✅ Nenhum arquivo .env com segredos rastreado."

echo "2. Executando suíte completa de testes da API Go..."
cd "$ROOT_DIR/api"
go test -race ./...
echo "✅ Todos os testes da API Go passaram com sucesso!"

echo "3. Executando build e verificação do frontend Next.js..."
cd "$ROOT_DIR/web"
unset VSCODE_ESM_ENTRYPOINT ELECTRON_RUN_AS_NODE
npm run build
echo "✅ Build estático do frontend concluído sem erros!"

echo "4. Verificando migrations do banco de dados..."
cd "$ROOT_DIR"
if [ ! -d "db/migrations" ]; then
  echo "❌ ERRO: Diretório db/migrations não encontrado!"
  exit 1
fi
MIGRATIONS_COUNT=$(find db/migrations -name "*.sql" | wc -l | tr -d ' ')
echo "✅ $MIGRATIONS_COUNT migrações SQL identificadas e prontas para execução."

echo "=================================================="
echo "🎉 TUDO PRONTO PARA O DEPLOY DE AMANHÃ!"
echo "Siga o passo a passo em: docs/DEPLOYMENT.md"
echo "=================================================="
